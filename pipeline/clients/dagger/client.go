package dagger

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dagger.io/dagger"
	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/cmdutil"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/syncutil"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Opts  clients.CommonOpts
	Log   *logrus.Logger
	State *state.Observer
}

func getArgMap(r state.Reader, extra map[string]string, required state.Arguments) (args.ArgMap, error) {
	m := args.ArgMap{}
	for _, v := range required {
		if val, ok := extra[v.Key]; ok {
			if err := m.Set(fmt.Sprintf("%s=%s", v.Key, val)); err != nil {
				return nil, err
			}
			continue
		}

		ok, err := r.Exists(v)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("'%s' does not exist in state", v.Key)
		}
		value, err := state.GetValueAsString(r, v)
		if err != nil {
			return nil, fmt.Errorf("arg: '%s', error: %w", v.Key, err)
		}

		if err := m.Set(fmt.Sprintf("%s=%s", v.Key, value)); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func New(opts clients.CommonOpts) (pipeline.Client, error) {
	s, err := state.NewDefaultState(opts.Log, opts.Args)
	if err != nil {
		return nil, err
	}

	return &Client{
		Opts:  opts,
		Log:   opts.Log,
		State: state.NewObserver(s),
	}, nil
}

func (c *Client) WaitForArgs(log logrus.FieldLogger, args state.Arguments) {
	log = log.WithField("arguments", args.String())
	log.Debugln("Waiting for arguments...")
	defer log.Debugln("Done for arguments")
	wg := &sync.WaitGroup{}
	for _, arg := range args {
		wg.Add(1)
		done := make(chan bool)
		go func(arg state.Argument, done chan bool) {
			c.Log.Debugf("Waiting for argument '%s'...", arg.Key)
			ticker := time.NewTicker(time.Second * 5)

			for {
				select {
				case <-ticker.C:
					log.WithField("argument", arg.Key).Debugln("Waiting for argument in state...")
				case <-done:
					log.WithField("argument", arg.Key).Debugln("Argument value received")
					wg.Done()
					return
				}
			}
		}(arg, done)
		cond := c.State.CondFor(arg)
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
		done <- true
	}
	wg.Wait()
}

func (c *Client) HandleState(d *dagger.Client, container *dagger.Container, step pipeline.Step) (*dagger.Container, map[string]string, error) {
	m := map[string]string{}

	for _, v := range step.RequiredArgs {
		switch v.Type {
		case state.ArgumentTypeFile:
			file, err := c.State.GetFile(v)
			if err != nil {
				return nil, nil, err
			}
			hPath := file.Name()
			cPath := path.Join("/var/scribe-state", hPath)
			container = container.WithMountedFile(cPath, d.Host().Directory(filepath.Dir(hPath)).File(filepath.Base(hPath)))
			m[v.Key] = cPath
		case state.ArgumentTypeFS:
			fallthrough
		case state.ArgumentTypeUnpackagedFS:
			hDir, err := c.State.GetDirectoryString(v)
			if err != nil {
				return nil, nil, err
			}
			cDir := path.Join("/var/scribe-state", hDir)
			container = container.WithMountedDirectory(cDir, d.Host().Directory(hDir))

			m[v.Key] = cDir
		}
	}
	return container, m, nil
}

// StepWalkFunc executes the contents of the step using the CLI client and is called once per step.
func (c *Client) StepWalkFunc(d *dagger.Client, wg *syncutil.WaitGroup, bin *dagger.Directory, src *dagger.Directory, path string) pipeline.StepWalkFunc {
	return func(ctx context.Context, step pipeline.Step) error {
		wg.Add(func(ctx context.Context) error {
			log := c.Log.WithFields(logrus.Fields{
				"step": step.Name,
			})

			c.WaitForArgs(log, state.Without(step.RequiredArgs, pipeline.ClientProvidedArguments))

			binPath := "/opt/scribe/pipeline"
			runner := d.Container().From(step.Image).
				WithMountedDirectory("/opt/scribe", bin).
				WithMountedDirectory("/var/scribe", src).
				WithEntrypoint([]string{}).
				WithWorkdir("/var/scribe")

			r, m, err := c.HandleState(d, runner, step)
			if err != nil {
				return err
			}
			runner = r

			argmap, err := getArgMap(c.State, m, step.RequiredArgs)
			if err != nil {
				return err
			}

			cmd, err := cmdutil.StepCommand(cmdutil.CommandOpts{
				CompiledPipeline: binPath,
				Step:             step,
				PipelineArgs: args.PipelineArgs{
					Path:   path,
					ArgMap: argmap,
				},
			})
			if err != nil {
				return err
			}

			// Some containers have entrypoints that can make `Exec` inconsistent. This attempts to disable / override that behavior.
			//runner = runner.WithEntrypoint([]string{})
			log.WithField("command", strings.Join(cmd, " ")).Debugln("Registering container with command...")
			runner = runner.WithExec(cmd)

			if stderr, err := runner.Stderr(ctx); err == nil {
				log.WithField("stream", "stderr").Infoln(stderr)
			} else {
				log.Errorln("Failed to get stderr from container. Dagger currently doesn't support streaming stdout/stderr directly; try re-running with `--log-level=debug` for more information")
			}

			stdout, err := runner.Stdout(ctx)
			if err != nil {
				return fmt.Errorf("failed to get stdout from container. The container is likely stopped due to an error. Consider re-running the pipeline with `--log-level=debug` for more information")
			}

			if _, err := runner.ExitCode(ctx); err != nil {
				return err
			}

			updates := map[string]state.StateValueJSON{}

			if err := json.Unmarshal([]byte(stdout), &updates); err != nil {
				return fmt.Errorf("error unmarshaling state JSON from CLI client: %w", err)
			}

			for _, v := range updates {
				log.Debugf("Setting arg '%s' in state", v.Argument.Key)
				if err := state.SetValueFromJSON(c.State, v); err != nil {
					return err
				}
			}

			return nil
		})
		return nil
	}
}

// PipelineWalkFunc is executed once for every set of parallel functions.
func (c *Client) PipelineWalkFunc(w pipeline.Walker, d *dagger.Client) pipeline.PipelineWalkFunc {
	return func(ctx context.Context, pipelines ...pipeline.Pipeline) error {
		// This is where all of the source code for the project lives, including the pipeline.
		src, err := c.State.GetDirectoryString(pipeline.ArgumentSourceFS)
		if err != nil {
			return err
		}
		// Some projects might not have the go.mod in the root or might have a separate go.mod for the pipeline itself.
		// If that's the case, then we need to provide that to the go build command.
		gomod, err := c.State.GetDirectoryString(pipeline.ArgumentPipelineGoModFS)
		if err != nil {
			return err
		}

		// Compile the pipeline so that individual steps can be ran in each container
		bin, err := CompilePipeline(ctx, d, src, gomod, c.Opts.Args.Path)
		if err != nil {
			return err
		}

		wg := syncutil.NewWaitGroup()
		for _, pipeline := range pipelines {
			wg.Add(func(ctx context.Context) error {
				stepwg := syncutil.NewWaitGroup()
				wf := c.StepWalkFunc(d, stepwg, bin, d.Host().Directory(src), c.Opts.Args.Path)
				// Walk through each step, add it to the waitgroup for this set of steps
				if err := w.WalkSteps(ctx, pipeline.ID, wf); err != nil {
					return err
				}

				if err := stepwg.Wait(ctx); err != nil {
					return err
				}

				return nil
			})
		}

		return wg.Wait(ctx)
	}
}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	d, err := dagger.Connect(
		ctx,
		// Until dagger has the ability to provide log streams per-container for stdout/stderr, we have to include the whole thing
		dagger.WithLogOutput(
			c.Log.WithField("stream", "dagger").WriterLevel(logrus.DebugLevel),
		),
	)

	if err != nil {
		return err
	}
	defer d.Close()
	return w.WalkPipelines(ctx, c.PipelineWalkFunc(w, d))
}

// Validate is ran internally before calling Run or Parallel and allows the client to effectively configure per-step requirements
// For example, Drone steps MUST have an image so the Drone client returns an error in this function when the provided step does not have an image.
// If the error encountered is not critical but should still be logged, then return a plumbing.ErrorSkipValidation.
// The error is checked with `errors.Is` so the error can be wrapped with fmt.Errorf.
func (c *Client) Validate(step pipeline.Step) error {
	return nil
}
