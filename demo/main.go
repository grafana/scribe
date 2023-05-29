package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/grafana/scribe/v2"
)

var (
	ArgumentRandomString = scribe.NewStringArgument("random-string")
)

var pipelines = []scribe.Pipeline{
	{
		Metadata: scribe.Metadata{
			Name: "create-random-string",
		},
		Steps: []scribe.Step{
			{
				Metadata: scribe.Metadata{
					Name: "produce-random-string",
				},
				Action: ProduceRandomString,
			},
		},
		Provides: []scribe.Argument{
			ArgumentRandomString,
		},
	},
	{
		Metadata: scribe.Metadata{
			Name: "print-random-string",
		},
		Steps:    []scribe.Step{},
		Requires: []scribe.Argument{ArgumentRandomString},
	},
}

func ProduceRandomString(ctx context.Context, args scribe.ActionOpts) error {
	str := fmt.Sprintf("%d", rand.Intn(100))

	return args.State.SetString(ArgumentRandomString, str)
}

func PrintRandomString(ctx context.Context, args scribe.ActionOpts) error {
	return nil
}

func main() {
	s := scribe.New()

	if err := s.Add(pipelines...); err != nil {
		panic(err)
	}

	if err := s.Run(); err != nil {
		panic(err)
	}
}
