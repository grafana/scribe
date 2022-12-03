package dagger

import (
	"context"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/cmdutil"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/syncutil"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Opts clients.CommonOpts

	Log *logrus.Logger
}

// WalkSteps is the handler for walking steps provided to the pipeline.Walker.
// It is called once per parallel group of steps.
// Every step pper pipeline with Dagger is executed using the same connection.
func (c *Client) StepWalkFunc(d *dagger.Client, bin *dagger.Directory, src *dagger.Directory, state *dagger.CacheVolume, path string) pipeline.StepWalkFunc {
	return func(ctx context.Context, steps ...pipeline.Step) error {
		for _, step := range steps {
			log := c.Log.WithFields(logrus.Fields{
				"step": step.Name,
			})

			log.Infoln("Running steps using dagger client...")
			binPath := "/opt/scribe/pipeline"
			runner := d.Container().From(step.Image).
				WithMountedDirectory("/opt/scribe", bin).
				WithMountedDirectory("/var/scribe", src).
				WithMountedCache("/var/scribe-state", state, dagger.ContainerWithMountedCacheOpts{}).
				WithEntrypoint([]string{}).
				WithWorkdir("/var/scribe")

			cmd, err := cmdutil.StepCommand(cmdutil.CommandOpts{
				CompiledPipeline: binPath,
				Step:             step,
				PipelineArgs: args.PipelineArgs{
					Path:  path,
					State: "file:///var/scribe-state/state.json",
				},
			})
			if err != nil {
				return err
			}

			// Some containers have entrypoints that can make `Exec` inconsistent. This attempts to disable / override that behavior.
			//runner = runner.WithEntrypoint([]string{})
			log.WithField("command", strings.Join(cmd, " ")).Debugln("Registering container with command...")
			runner = runner.WithExec(cmd)

			if stdout, err := runner.Stderr(ctx); err == nil {
				log.WithField("stream", "stdout").Infoln(stdout)
			}

			if stderr, err := runner.Stderr(ctx); err == nil {
				log.WithField("stream", "stderr").Infoln(stderr)
			}

			if _, err := runner.ExitCode(ctx); err != nil {
				return err
			}
		}

		return nil
	}
}

// WalkPipelines is the handler for walking pipelines provided to the pipeline.Walker.
// It is called once per parallel group of pipelines.
func (c *Client) PipelineWalkFunc(w pipeline.Walker, d *dagger.Client, cache *dagger.CacheVolume) pipeline.PipelineWalkFunc {
	return func(ctx context.Context, pipelines ...pipeline.Pipeline) error {
		// This is where all of the source code for the project lives, including the pipeline.
		src, err := c.Opts.State.GetDirectoryString(pipeline.ArgumentSourceFS)
		if err != nil {
			return err
		}
		// Some projects might not have the go.mod in the root or might have a separate go.mod for the pipeline itself.
		// If that's the case, then we need to provide that to the go build command.
		gomod, err := c.Opts.State.GetDirectoryString(pipeline.ArgumentPipelineGoModFS)
		if err != nil {
			return err
		}

		log := c.Log.WithFields(logrus.Fields{
			"source":   src,
			"go.mod":   gomod,
			"pipeline": c.Opts.Args.Path,
		})

		log.Infoln("Compiling pipeline...")
		// Compile the pipeline so that individual steps can be ran in each container
		bin, err := CompilePipeline(ctx, d, src, gomod, c.Opts.Args.Path)
		if err != nil {
			return err
		}

		log.Infoln("Done compiling pipeline")

		wg := syncutil.NewPipelineWaitGroup()
		for _, pipeline := range pipelines {
			wf := c.StepWalkFunc(d, bin, d.Host().Directory(src), cache, c.Opts.Args.Path)
			wg.Add(pipeline, w, wf)
		}

		return wg.Wait(ctx)
	}
}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	d, err := dagger.Connect(ctx, dagger.WithLogOutput(c.Log.Writer()))
	if err != nil {
		return err
	}
	defer d.Close()

	cache := d.CacheVolume("scribe-state")
	return w.WalkPipelines(ctx, c.PipelineWalkFunc(w, d, cache))
}

// Validate is ran internally before calling Run or Parallel and allows the client to effectively configure per-step requirements
// For example, Drone steps MUST have an image so the Drone client returns an error in this function when the provided step does not have an image.
// If the error encountered is not critical but should still be logged, then return a plumbing.ErrorSkipValidation.
// The error is checked with `errors.Is` so the error can be wrapped with fmt.Errorf.
func (c *Client) Validate(step pipeline.Step) error {
	return nil
}
