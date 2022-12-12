package wrappers

import (
	"context"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/plog"
	"github.com/sirupsen/logrus"
)

type LogWrapper struct {
	Opts clients.CommonOpts
	Log  logrus.FieldLogger
}

func (l *LogWrapper) Fields(ctx context.Context, step pipeline.Step) logrus.Fields {
	fields := plog.DefaultFields(ctx, step, l.Opts)

	return fields
}

func (l *LogWrapper) WrapStep(step pipeline.Step) pipeline.Step {
	action := step.Action

	// Steps that provide a nil action should continue to provide a nil action.
	// There is nothing for us to log in the execution of this action anyways, though there is an implication that
	// this step may execute something that is not defined in the pipeline.
	if step.Action == nil {
		return step
	}

	step.Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
		l.Log.WithFields(l.Fields(ctx, step)).Infoln("starting step")

		stdoutFields := l.Fields(ctx, step)
		stdoutFields["stream"] = "stdout"

		stderrFields := l.Fields(ctx, step)
		stderrFields["stream"] = "stderr"

		opts.Stdout = l.Log.WithFields(stdoutFields).Writer()
		opts.Stderr = l.Log.WithFields(stderrFields).Writer()

		if err := action(ctx, opts); err != nil {
			l.Log.WithFields(l.Fields(ctx, step)).Infoln("encountered error", err.Error())
			return err
		}

		l.Log.WithFields(l.Fields(ctx, step)).Infoln("done running step without error")
		return nil
	}

	return step
}

func (l *LogWrapper) Wrap(wf pipeline.StepWalkFunc) pipeline.StepWalkFunc {
	return func(ctx context.Context, step pipeline.Step) error {
		steps := l.WrapStep(step)

		if err := wf(ctx, steps); err != nil {
			return err
		}
		return nil
	}
}
