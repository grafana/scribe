package dagger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	"github.com/grafana/scribe/stringutil"
	"github.com/grafana/scribe/syncutil"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Opts  clients.CommonOpts
	Log   *logrus.Logger
	State *state.Observer
}

// getArgMap builds an argument map to supply to the step.
// Since the steps are executed using the CLI mode, the state arguments that they need are simply passed as CLI arguments and mounted into the container's filesystem if necessary.
func getArgMap(ctx context.Context, r state.Reader, extra map[string]string, required state.Arguments) (args.ArgMap, error) {
	m := args.ArgMap{}
	for _, v := range required {
		if val, ok := extra[v.Key]; ok {
			m[v.Key] = val
			continue
		}

		ok, err := r.Exists(ctx, v)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("'%s' does not exist in state", v.Key)
		}
		value, err := state.GetValueAsString(ctx, r, v)
		if err != nil {
			return nil, fmt.Errorf("arg: '%s', error: %w", v.Key, err)
		}

		if err := m.Set(fmt.Sprintf("%s=%s", v.Key, value)); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func New(ctx context.Context, opts clients.CommonOpts) (pipeline.Client, error) {
	s, err := state.NewDefaultState(ctx, opts.Log, opts.Args)
	if err != nil {
		return nil, err
	}

	return &Client{
		Opts:  opts,
		Log:   opts.Log,
		State: state.NewObserver(s),
	}, nil
}

func (c *Client) WaitForArgs(ctx context.Context, log logrus.FieldLogger, args state.Arguments) {
	log = log.WithField("arguments", args.String())
	log.Infoln("Checking for argument existence in state before waiting...")

	wait := state.Arguments{}
	for _, arg := range args {
		exists, err := c.State.Exists(ctx, arg)
		// If it does exist and there was no error returned, then we don't need to wait for it.
		if exists && err == nil {
			log.Infof("We do not need to wait for '%s'", arg.Key)
			continue
		}

		if err != nil {
			log.WithError(err).Infoln("Got error from state when checking for existence")
		}

		wait = append(wait, arg)
	}

	if len(wait) == 0 {
		return
	}

	log.Infof("Waiting for '%d' arguments...", len(wait))
	defer log.Infoln("Done waiting for arguments")
	wg := &sync.WaitGroup{}
	for _, arg := range wait {
		wg.Add(1)
		done := make(chan bool)
		go func(arg state.Argument, done chan bool) {
			c.Log.Infof("Waiting for argument '%s'...", arg.Key)
			ticker := time.NewTicker(time.Second * 5)

			for {
				select {
				case <-ticker.C:
					log.WithField("argument", arg.Key).Infoln("Waiting for argument in state...")
					// Hack: checking if it exists in state yet if we were somehow not notified.
					if exists, err := c.State.Exists(ctx, arg); exists && err == nil {
						log.WithField("argument", arg.Key).Infoln("Argument exists in state now")
						c.State.CondFor(ctx, arg).Broadcast()
					}
				case <-done:
					log.WithField("argument", arg.Key).Infoln("Argument value received")
					wg.Done()
					return
				}
			}
		}(arg, done)
		cond := c.State.CondFor(ctx, arg)
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
		done <- true
	}
	wg.Wait()
}

// HandleRequiredArgs modifies the provided container to account for the arguments provided by and required by the provided step, then returns the modified container.
func (c *Client) HandleRequiredArgs(ctx context.Context, d *dagger.Client, container *dagger.Container, step pipeline.Step) (*dagger.Container, map[string]string, error) {
	m := map[string]string{}

	for _, v := range step.RequiredArgs {
		switch v.Type {
		case state.ArgumentTypeFile:
			file, err := c.State.GetFile(ctx, v)
			if err != nil {
				return nil, nil, err
			}

			if err := file.Close(); err != nil {
				c.Log.WithError(err).Warnln("error file retrieved from state")
			}

			// Where the file exists on the host.
			// file.Name() definitely exists since it came from 'container.Export', and is a full, absolute path.
			// Example: /tmp/sCQdrMlS/1597600537
			hPath := file.Name() // Definitely exists

			// The path on the container. It should always be in '/var/scribe-state'.
			// Example: /var/scribe-state/tmp/sCQdrMlS/1597600537
			var (
				cPath = path.Join("/var/scribe-state", hPath)
				dir   = d.Host().Directory(filepath.Dir(hPath))
			)

			// Add the mount to the container
			// Maybe not the best solution here but for some reason getting the file directly from the dir wasn't working as expected...
			container = container.WithMountedDirectory(filepath.Dir(cPath), dir)
			// Tell the container to look for this argument at "cPath"
			m[v.Key] = cPath
		case state.ArgumentTypeFS:
			fallthrough
		case state.ArgumentTypeUnpackagedFS:
			hDir, err := c.State.GetDirectoryString(ctx, v)
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

func (c *Client) HandleStep(ctx context.Context, step pipeline.Step, d *dagger.Client, wg *syncutil.WaitGroup, bin *dagger.Directory, src *dagger.Directory, path string) error {
	wg.Add(func(ctx context.Context) error {
		log := c.Log.WithFields(logrus.Fields{
			"step": step.Name,
		})

		log.Infoln("Waiting for pipeline arguments before registering step...")
		c.WaitForArgs(ctx, log, state.Without(step.RequiredArgs, pipeline.ClientProvidedArguments))
		log.Infoln("Done waiting for pipeline arguments")

		binPath := "/opt/scribe/pipeline"
		runner := d.Container().From(step.Image).
			WithMountedDirectory("/opt/scribe", bin).
			WithMountedDirectory("/var/scribe", src).
			WithEntrypoint([]string{}).
			WithWorkdir("/var/scribe")

		r, m, err := c.HandleRequiredArgs(ctx, d, runner, step)
		if err != nil {
			return err
		}
		runner = r

		argmap, err := getArgMap(ctx, c.State, m, step.RequiredArgs)
		if err != nil {
			return err
		}

		for k, v := range argmap {
			log.Infoln("ArgMap", k, v)
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
		log.WithField("command", strings.Join(cmd, " ")).Infoln("Registering container with command...")
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
			if v.Argument.Type == state.ArgumentTypeFile {
				containerPath := v.Value.(string)
				hostPath := filepath.Join(os.TempDir(), filepath.Base(containerPath))
				dir := runner.Directory(filepath.Dir(containerPath))
				_, err := dir.File(filepath.Base(containerPath)).Export(ctx, hostPath)
				if err != nil {
					return err
				}

				v.Value = hostPath
			}

			// If the container gives us a Filesystem argument, we must mount it in a temporary location in order to create the tar.gz so the state
			// can properly handle it.
			if v.Argument.Type == state.ArgumentTypeFS {
				var (
					hostPath      = filepath.Join(os.TempDir(), stringutil.Random(8))
					containerPath = v.Value.(string)
					dir           = runner.Directory(containerPath)
				)
				if _, err := dir.Export(ctx, hostPath); err != nil {
					return err
				}
				v.Value = hostPath
			}

			if err := state.SetValueFromJSON(ctx, c.State, v); err != nil {
				return err
			}
		}

		return nil
	})
	return nil
}

// StepWalkFunc executes the contents of the step using the CLI client and is called once per step.
func (c *Client) StepWalkFunc(d *dagger.Client, wg *syncutil.WaitGroup, bin *dagger.Directory, src *dagger.Directory, path string) pipeline.StepWalkFunc {
	return func(ctx context.Context, step pipeline.Step) error {
		return c.HandleStep(ctx, step, d, wg, bin, src, path)
	}
}

// PipelineWalkFunc is executed once for every set of parallel functions.
func (c *Client) PipelineWalkFunc(w *pipeline.Collection, wg *syncutil.WaitGroup, bin, src *dagger.Directory, d *dagger.Client) pipeline.PipelineWalkFunc {
	return func(ctx context.Context, p pipeline.Pipeline) error {
		if p.ID == 0 {
			return nil
		}

		log := c.Log.WithFields(logrus.Fields{
			"pipeline": p.Name,
		})

		// notFoundArgs := []state.Argument{}
		// for _, v := range p.RequiredArgs {
		// 	ok, err := c.State.Exists(ctx, v)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// If this required arg doesn't exist in the state, then wait for it...
		// 	if !ok {
		// 		notFoundArgs = append(notFoundArgs, v)
		// 	}
		// }

		// // Wait for not found arguments
		// //c.WaitForArgs(ctx, c.Log, notFoundArgs)
		log.Infoln("Processing pipeline with Dagger")
		wg.Add(func(ctx context.Context) error {
			swg := syncutil.NewWaitGroup()
			// Before running the steps in the pipeline, wait for the arguments to be in the state that this pipeline is requesting

			log.Infoln("Waiting for arguments to be ready before registering containers...")
			c.WaitForArgs(ctx, log, p.RequiredArgs)
			log.Infoln("Done waiting for arguments")

			wf := c.StepWalkFunc(d, swg, bin, src, c.Opts.Args.Path)
			log.Infoln("Walking through steps and registering containers...")
			// Walk through each step, add it to the waitgroup for this set of steps
			if err := w.WalkSteps(ctx, p.ID, wf); err != nil {
				return err
			}
			log.Infoln("Done walking through")

			log.Infoln("Waiting for steps to complete")
			return swg.Wait(ctx)
		})

		return nil
	}
}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (c *Client) Done(ctx context.Context, w *pipeline.Collection) error {
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

	c.Log.Infoln("Getting source directory from state...")

	// This is where all of the source code for the project lives, including the pipeline.
	src, err := c.State.GetDirectoryString(ctx, pipeline.ArgumentSourceFS)
	if err != nil {
		return err
	}

	// Some projects might not have the go.mod in the root or might have a separate go.mod for the pipeline itself.
	// If that's the case, then we need to provide that to the go build command.
	gomod, err := c.State.GetDirectoryString(ctx, pipeline.ArgumentPipelineGoModFS)
	if err != nil {
		return err
	}
	c.Log.Infoln("Done setting up pipeline")

	// Compile the pipeline so that individual steps can be ran in each container
	bin, err := CompilePipeline(ctx, d, c.Opts.Name, src, gomod, c.Opts.Args.Path)
	if err != nil {
		return err
	}

	wg := syncutil.NewWaitGroup()
	wf := c.PipelineWalkFunc(w, wg, bin, d.Host().Directory(src), d)

	if err := w.WalkPipelines(ctx, wf); err != nil {
		return err
	}

	return wg.Wait(ctx)
}

// Validate is ran internally before calling Run or Parallel and allows the client to effectively configure per-step requirements
// For example, Drone steps MUST have an image so the Drone client returns an error in this function when the provided step does not have an image.
// If the error encountered is not critical but should still be logged, then return a plumbing.ErrorSkipValidation.
// The error is checked with `errors.Is` so the error can be wrapped with fmt.Errorf.
func (c *Client) Validate(step pipeline.Step) error {
	return nil
}
