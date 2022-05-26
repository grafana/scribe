package main

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

var (
	ArgumentRandomInt = pipeline.NewStringArgument("random_int")
)

func StepProduceRandom() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		r := rand.Int63n(10000)
		opts.State.Set(ArgumentRandomInt.Key, strconv.FormatInt(r, 10))
		return nil
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentRandomInt)
}

func StepPrintRandom() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		strVal, err := opts.State.Get(ArgumentRandomInt.Key)
		if err != nil {
			return err
		}

		v, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}

		opts.Logger.Println("Got value", v)
		return nil
	}

	step := pipeline.Step{
		Action: action,
		Arguments: []pipeline.Argument{
			ArgumentRandomInt,
		},
		Image: "test",
	}

	return step.WithArguments(ArgumentRandomInt)
}

// func init() {
// 	rand.Seed(time.Now().Unix())
// }

func main() {
	sw := shipwright.New("state-example")
	defer sw.Done()

	sw.Run(
		StepProduceRandom().WithName("test 1"),
		StepPrintRandom().WithName("test 2"),
	)
}
