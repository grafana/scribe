package plog

import (
	"github.com/sirupsen/logrus"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func StepFields(step pipeline.Step) logrus.Fields {
	return logrus.Fields{
		"step":   step.Name,
		"serial": step.Serial,
	}
}

func PipelineFields(opts pipeline.CommonOpts) logrus.Fields {
	return logrus.Fields{
		"id": opts.Args.BuildID,
	}
}

func Combine(field ...logrus.Fields) logrus.Fields {
	fields := logrus.Fields{}

	for _, m := range field {
		for k, v := range m {
			fields[k] = v
		}
	}

	return fields
}

func DefaultFields(step pipeline.Step, opts pipeline.CommonOpts) logrus.Fields {
	return Combine(StepFields(step), PipelineFields(opts))
}
