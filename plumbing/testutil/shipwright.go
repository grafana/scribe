package testutil

import (
	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

func NewShipwright(initializer shipwright.InitializerFunc) *shipwright.Shipwright[pipeline.Action] {
	log := logrus.New()

	opts := pipeline.CommonOpts{
		Log: log,
	}
	client := initializer(opts)

	return &shipwright.Shipwright[pipeline.Action]{
		Opts:       opts,
		Client:     client,
		Collection: shipwright.NewDefaultCollection(opts),
	}
}

func NewShipwrightMulti(initializer shipwright.InitializerFunc) *shipwright.Shipwright[pipeline.StepList] {
	return nil
}
