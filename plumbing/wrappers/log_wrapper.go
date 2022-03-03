package wrappers

import (
	"context"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

type LogWrapper struct {
	Log *plog.Logger
}

func (l *LogWrapper) WrapStep(step ...pipeline.Step) []pipeline.Step {
	for i, v := range step {
		action := step[i].Action
		step[i].Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
			l.Log.Infof("[%s] starting step '%s'", time.Now().String(), v.Name)
			if err := action(ctx, opts); err != nil {
				l.Log.Infof("[%s] done running step '%s' with error %s", time.Now().String(), v.Name, err.Error())
				return err
			}

			l.Log.Infof("[%s] done running step '%s' without", time.Now().String(), v.Name)
			return nil
		}
	}

	return step
}

func (l *LogWrapper) Wrap(wf pipeline.WalkFunc) pipeline.WalkFunc {
	return func(ctx context.Context, step ...pipeline.Step) error {
		for _, v := range step {
			l.Log.Infoln(time.Now(), "running step", v.Name)
		}

		if err := wf(ctx, step...); err != nil {
			return err
		}

		for _, v := range step {
			l.Log.Infoln(time.Now(), "done running step", v.Name)
		}

		return nil
	}
}
