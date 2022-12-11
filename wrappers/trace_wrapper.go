package wrappers

import (
	"context"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/plog"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type TraceWrapper struct {
	Opts   clients.CommonOpts
	Tracer opentracing.Tracer
}

func (l *TraceWrapper) Fields(ctx context.Context, step pipeline.Step) logrus.Fields {
	fields := plog.DefaultFields(ctx, step, l.Opts)

	return fields
}

func TagSpan(span opentracing.Span, opts clients.CommonOpts, step pipeline.Step) {
	span.SetTag("job", "scribe")
	span.SetTag("build_id", opts.Args.BuildID)
}

func (l *TraceWrapper) WrapStep(step pipeline.Step) pipeline.Step {
	// Steps that provide a nil action should continue to provide a nil action.
	// There is nothing for us to trace in the execution of this action anyways, though there is an implication that
	// this step may execute something that is not defined in the pipeline.
	if step.Action == nil {
		return step
	}

	action := step.Action
	step.Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
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

	return step
}

func (l *TraceWrapper) Wrap(wf pipeline.StepWalkFunc) pipeline.StepWalkFunc {
	return func(ctx context.Context, step pipeline.Step) error {
		steps := l.WrapStep(step)

		if err := wf(ctx, steps); err != nil {
			return err
		}
		return nil
	}
}
