package main

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

var (
	ArgumentRandomInt = pipeline.NewStringArgument("random_int")
)

func StepProduceRandom(sw *shipwright.Shipwright[pipeline.Action]) pipeline.Step[pipeline.Action] {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		r := rand.Int63n(10000)
		sw.Opts.State.Set(ArgumentRandomInt.Key, strconv.FormatInt(r, 10))
		return nil
	}

	return pipeline.NewStep(action).Provides(ArgumentRandomInt)
}

func StepPrintRandom(sw *shipwright.Shipwright[pipeline.Action]) pipeline.Step[pipeline.Action] {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		strVal, err := sw.Opts.State.Get(ArgumentRandomInt.Key)
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

	return pipeline.NewStep(action).WithArguments(ArgumentRandomInt)
}
func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	sw := shipwright.New("state-example")
	defer sw.Done()

	sw.Run(
		StepProduceRandom(sw).WithName("create-random-number"),
		StepPrintRandom(sw).WithName("log-random-number"),
	)
}
