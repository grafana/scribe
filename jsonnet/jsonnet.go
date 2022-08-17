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
	"github.com/google/go-jsonnet/linter"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func Lint(path string) pipeline.Step {
	path, err := filepath.Abs(path)
	if err != nil {
		return pipeline.NewStep(nil)
	}
	vm := jsonnet.MakeVM()
	return pipeline.NewStep(
		func(ctx context.Context, opts pipeline.ActionOpts) error {
			return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
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
				g := linter.LintSnippet(vm, os.Stderr, []linter.Snippet{snippet})
				if !g {
					return fmt.Errorf("jsonnet linting failed for file: %s", path)
				}
				return nil
			})
		},
	)
}
