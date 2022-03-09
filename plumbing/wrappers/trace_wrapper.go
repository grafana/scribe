package wrappers

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/plog"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type TraceWrapper struct {
	Opts   pipeline.CommonOpts
	Tracer opentracing.Tracer
}

func (l *TraceWrapper) Fields(ctx context.Context, step pipeline.Step) logrus.Fields {
	fields := plog.DefaultFields(ctx, step, l.Opts)

	return fields
}

func TagSpan(span opentracing.Span, opts pipeline.CommonOpts, step pipeline.Step) {
	span.SetTag("job", "shipwright")
	span.SetTag("build_id", opts.Args.BuildID)
}

func (l *TraceWrapper) WrapStep(steps ...pipeline.Step) []pipeline.Step {
	for i := range steps {
		step := steps[i]
		action := step.Action
		steps[i].Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
			parent := opentracing.SpanFromContext(ctx)
			span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, l.Tracer, step.Name, opentracing.ChildOf(parent.Context()))
			TagSpan(span, l.Opts, step)
			defer span.Finish()

			if err := action(ctx, opts); err != nil {
				span.SetTag("error", err)
				return err
			}

			return nil
		}
	}

	return steps
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
