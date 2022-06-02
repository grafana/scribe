package main

import (
	"context"
	"io"
	"io/fs"
	"math/rand"
	"path/filepath"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/scribe/plumbing/stringutil"
)

var (
	ArgumentRandomString  = pipeline.NewStringArgument("random_string")
	ArgumentRandomInt     = pipeline.NewInt64Argument("random_int")
	ArgumentRandomFloat64 = pipeline.NewFloat64Argument("random_float")
	ArgumentTextFile      = pipeline.NewFileArgument("text_file")
	ArgumentDirectory     = pipeline.NewDirectoryArgument("example_directory")
)

func StepProduceRandomString() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		r := stringutil.Random(12)
		opts.State.SetString(ArgumentRandomString, r)
		return nil
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentRandomString)
}

func StepProduceRandomFloat64() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		r := rand.Float64() * 10000
		opts.State.SetFloat64(ArgumentRandomFloat64, r)
		return nil
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentRandomFloat64)
}

func StepProduceRandomInt64() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		r := rand.Int63n(10000)
		opts.State.SetInt64(ArgumentRandomInt, r)
		return nil
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentRandomInt)
}

func StepStoreFile() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		opts.Logger.Infoln("Storing file ./example-state-file.txt in state")
		return opts.State.SetFile(ArgumentTextFile, filepath.Join(opts.Path, "./example-state-file.txt"))
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentTextFile)
}

func StepStoreDirectory() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		return opts.State.SetDirectory(ArgumentDirectory, filepath.Join(opts.Path, "./example-directory"))
	}

	step := pipeline.NewStep(action)

	return step.Provides(ArgumentDirectory)
}

func StepPrintRandomInt64() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		v, err := opts.State.GetInt64(ArgumentRandomInt)
		if err != nil {
			return err
		}

		opts.Logger.Println("Got random int", v)
		return nil
	}

	step := pipeline.NewStep(action)
	return step.WithArguments(ArgumentRandomInt)
}

func StepPrintRandomFloat64() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		v, err := opts.State.GetFloat64(ArgumentRandomFloat64)
		if err != nil {
			return err
		}

		opts.Logger.Println("Got random float", v)
		return nil
	}

	step := pipeline.NewStep(action)
	return step.WithArguments(ArgumentRandomFloat64)
}

func StepPrintRandomString() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		v, err := opts.State.GetString(ArgumentRandomString)
		if err != nil {
			return err
		}

		opts.Logger.Println("Got random string", v)
		return nil
	}

	step := pipeline.NewStep(action)
	return step.WithArguments(ArgumentRandomString)
}

func StepPrintFile() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		v, err := opts.State.GetFile(ArgumentTextFile)
		if err != nil {
			return err
		}

		w := opts.Logger.WithField("file", ArgumentTextFile.Key).Writer()
		if _, err := io.Copy(w, v); err != nil {
			return err
		}

		return nil
	}

	step := pipeline.NewStep(action)
	return step.WithArguments(ArgumentTextFile)
}

func StepPrintDirectory() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		dir, err := opts.State.GetDirectory(ArgumentDirectory)
		if err != nil {
			return err
		}

		fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
			opts.Logger.Infoln(path)
			return nil
		})

		return nil
	}

	step := pipeline.NewStep(action)
	return step.WithArguments(ArgumentDirectory)
}

// func init() {
// 	rand.Seed(time.Now().Unix())
// }

func main() {
	sw := scribe.New("state-example")
	defer sw.Done()

	sw.Run(
		StepProduceRandomInt64().WithName("create random int64"),
		StepProduceRandomFloat64().WithName("create random float64"),
		StepProduceRandomString().WithName("create random string"),
		StepStoreFile().WithName("store file"),
		StepStoreDirectory().WithName("store directory"),
	)

	sw.Run(
		StepPrintRandomInt64().WithName("print random int64"),
		StepPrintRandomFloat64().WithName("print random float64"),
		StepPrintRandomString().WithName("print random string"),
		StepPrintFile().WithName("print file"),
		StepPrintDirectory().WithName("print directory"),
	)
}
