package testutil

import (
	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/sirupsen/logrus"
)

func NewScribe(initializer scribe.InitializerFunc) *scribe.Scribe {
	log := logrus.New()

	opts := clients.CommonOpts{
		Log: log,
	}
	client := initializer(opts)

	return &scribe.Scribe{
		Opts:       opts,
		Client:     client,
		Collection: scribe.NewDefaultCollection(opts),
	}
}

func NewScribeMulti(initializer scribe.InitializerFunc) *scribe.Scribe {
	return nil
}
