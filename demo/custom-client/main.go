package main

import (
	"context"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

type MyClient struct {
	Log logrus.FieldLogger
}

func (c *MyClient) Validate(step pipeline.Step[pipeline.Action]) error {
	return nil
}

func (c *MyClient) Done(ctx context.Context, w pipeline.Walker) error {
	return w.WalkPipelines(ctx, func(ctx context.Context, pipelines ...pipeline.Step[pipeline.Pipeline]) error {
		c.Log.Infoln("pipelines:", pipeline.StepNames(pipelines))
		for _, v := range pipelines {
			err := w.WalkSteps(ctx, v.Serial, func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
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
	shipwright.RegisterClient("my-custom-client", func(opts pipeline.CommonOpts) pipeline.Client {
		return &MyClient{
			Log: opts.Log,
		}
	})
}

func main() {
	sw := shipwright.New("custom-client")
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
