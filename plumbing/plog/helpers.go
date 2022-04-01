package plog

import (
	"strings"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

func LogSteps(logger logrus.FieldLogger, steps []pipeline.Step) {
	s := make([]string, len(steps))
	for i, v := range steps {
		s[i] = v.Name
	}
	logger.Infof("[%d] step(s) %s", len(steps), strings.Join(s, " | "))
}
