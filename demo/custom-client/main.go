package main

import (
	"context"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/sirupsen/logrus"
)

type MyClient struct {
	Log logrus.FieldLogger
}

func (c *MyClient) Validate(step pipeline.Step) error {
	return nil
}

func (c *MyClient) Done(ctx context.Context, w pipeline.Walker) error {
	return w.WalkPipelines(ctx, func(ctx context.Context, pipelines ...pipeline.Pipeline) error {
		c.Log.Infoln("pipelines:", pipeline.PipelineNames(pipelines))
		for _, v := range pipelines {
			err := w.WalkSteps(ctx, v.ID, func(ctx context.Context, steps ...pipeline.Step) error {
				c.Log.Infoln("steps:", pipeline.StepNames(steps))
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func init() {
	scribe.RegisterClient("my-custom-client", func(opts clients.CommonOpts) pipeline.Client {
		return &MyClient{
			Log: opts.Log,
		}
	})
}

func main() {
	sw := scribe.New("custom-client")
	defer sw.Done()

	sw.Run(
		pipeline.NoOpStep.WithName("step 1"),
		pipeline.NoOpStep.WithName("step 2"),
	)
	sw.Parallel(
		pipeline.NoOpStep.WithName("step 3"),
		pipeline.NoOpStep.WithName("step 4"),
	)
}
