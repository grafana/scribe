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
