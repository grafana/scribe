package wrappers

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

type TraceWrapper struct {
	Opts   pipeline.CommonOpts
	Tracer opentracing.Tracer
}

func (l *TraceWrapper) Fields(ctx context.Context, step pipeline.Step) logrus.Fields {
	fields := plog.DefaultFields(ctx, step, l.Opts)

	return fields
}

func (l *TraceWrapper) WrapStep(step ...pipeline.Step) []pipeline.Step {
	for i, v := range step {
		action := step[i].Action
		step[i].Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
			parent := opentracing.SpanFromContext(ctx)
			span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, l.Tracer, v.Name, opentracing.ChildOf(parent.Context()))
			span.SetTag("serial", v.Serial)
			defer span.Finish()

			if err := action(ctx, opts); err != nil {
				span.SetTag("error", err)
				return err
			}

			return nil
		}
	}

	return step
}

func (l *TraceWrapper) Wrap(wf pipeline.WalkFunc) pipeline.WalkFunc {
	return func(ctx context.Context, step ...pipeline.Step) error {
		steps := l.WrapStep(step...)

		if err := wf(ctx, steps...); err != nil {
			return err
		}
		return nil
	}
}
