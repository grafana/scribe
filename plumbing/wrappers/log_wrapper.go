package wrappers

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

type LogWrapper struct {
	Log *logrus.Logger
}

func Fields(step pipeline.Step) logrus.Fields {
	fields := plog.StepFields(step)

	fields["time"] = time.Now()

	return fields
}

func (l *LogWrapper) WrapStep(step ...pipeline.Step) []pipeline.Step {
	for i, v := range step {
		action := step[i].Action
		step[i].Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
			l.Log.WithFields(Fields(v)).Infoln("starting step'")
			if err := action(ctx, opts); err != nil {
				l.Log.WithFields(Fields(v)).Infoln("encountered error", err.Error())
				return err
			}

			l.Log.WithFields(Fields(v)).Infoln("done running step without error")
			return nil
		}
	}

	return step
}

func (l *LogWrapper) Wrap(wf pipeline.WalkFunc) pipeline.WalkFunc {
	return func(ctx context.Context, step ...pipeline.Step) error {
		steps := l.WrapStep(step...)

		if err := wf(ctx, steps...); err != nil {
			return err
		}
		return nil
	}
}
