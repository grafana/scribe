package plog

import (
	"strings"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

func LogSteps[T pipeline.StepContent](logger logrus.FieldLogger, steps []pipeline.Step[T]) {
	s := make([]string, len(steps))
	for i, v := range steps {
		s[i] = v.Name
	}
	logger.Infof("[%d] step(s) %s", len(steps), strings.Join(s, " | "))
}
