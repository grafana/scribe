package plog

import (
	"github.com/sirupsen/logrus"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

func StepFields(step pipeline.Step) logrus.Fields {
	return logrus.Fields{
		"step":   step.Name,
		"serial": step.Serial,
	}
}
