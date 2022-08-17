package jsonnet

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/formatter"
	"github.com/google/go-jsonnet/linter"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/tanka/pkg/kubernetes/util"
)

func Lint(path string) pipeline.Step {
	var errFiles []string
	vm := jsonnet.MakeVM()

	return pipeline.NewStep(
		func(ctx context.Context, opts pipeline.ActionOpts) error {
			path := filepath.Join(opts.State.MustGetDirectoryString(pipeline.ArgumentSourceFS), path)
			err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || (!strings.Contains(d.Name(), ".jsonnet") && !strings.Contains(d.Name(), ".libsonnet")) {
					return nil
				}
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				data, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				err = f.Close()
				if err != nil {
					return err
				}
				snippet := linter.Snippet{FileName: path, Code: string(data)}
				if !linter.LintSnippet(vm, opts.Stderr, []linter.Snippet{snippet}) {
					errFiles = append(errFiles, path)
				}
				return nil
			})
			if err != nil {
				return err
			}
			if len(errFiles) != 0 {
				return fmt.Errorf("jsonnetfmt found lint errors in files: %s", errFiles)
			}
			return nil
		},
	).WithArguments(pipeline.ArgumentSourceFS)
}

func Format(path string) pipeline.Step {
	return pipeline.NewStep(
		func(ctx context.Context, opts pipeline.ActionOpts) error {
			var errFiles []string

			path := filepath.Join(opts.State.MustGetDirectoryString(pipeline.ArgumentSourceFS), path)
			err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || (!strings.Contains(d.Name(), ".jsonnet") && !strings.Contains(d.Name(), ".libsonnet")) {
					return nil
				}
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				data, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				err = f.Close()
				if err != nil {
					return err
				}
				// snippet := linter.Snippet{FileName: path, Code: string(data)}
				out, err := formatter.Format(d.Name(), string(data), formatter.DefaultOptions())
				if err != nil {
					return fmt.Errorf("jsonnet linting failed for file: %s", path)
				}
				if out == string(data) {
					return nil
				}
				s, err := util.DiffStr(d.Name(), string(data), out)
				fmt.Printf("diff: %s\n", s)
				if err != nil {
					return err
				}
				errFiles = append(errFiles, path)
				return nil
			})
			if err != nil {
				return err
			}
			if len(errFiles) != 0 {
				return fmt.Errorf("jsonnetfmt found lint errors in files: %s", errFiles)
			}
			return nil
		},
	).WithArguments(pipeline.ArgumentSourceFS)
}
