package git

import (
	"context"
	"os"
	osexec "os/exec"
	"strconv"

	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/exec"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

func GetCloneOpts(sw shipwright.Shipwright) (*CloneOpts, error) {
	ref, err := sw.Value(pipeline.ArgumentCommitRef)
	if err != nil {
		return nil, err
	}

	u, err := sw.Value(pipeline.ArgumentRemoteURL)
	if err != nil {
		return nil, err
	}

	workDir, err := sw.Value(pipeline.ArgumentWorkingDir)
	if err != nil {
		return nil, err
	}

	return &CloneOpts{
		Folder: workDir,
		URL:    u,
		Ref:    ref,
	}, nil
}

func clone(sw shipwright.Shipwright, depth int) pipeline.StepAction {
	return func(ctx context.Context, aopts pipeline.ActionOpts) error {
		opts, err := GetCloneOpts(sw)
		if err != nil {
			return err
		}

		sw.Log.Infoln("Got git opts:", opts)
		// Instead of using a re-implementation of `git` in Go (like go-git)
		// we will just delegate git steps into sub-shells.
		// This can introduce some possibly variable effects on different machines depending on which version of git is installed, or if one is installed at all.
		// If we find that happens a lot, then we can consider something like go-git.

		var (
			cmd          = "git"
			cloneArgs    = []string{"clone", opts.URL, opts.Folder}
			checkoutArgs = []string{"checkout", opts.Ref}
		)

		if depth == 0 {
			cloneArgs = append(cloneArgs, "--depth", strconv.Itoa(depth))
		}

		// Don't run `git clone` or checkout if the working directory already exists
		if _, err := os.Stat(opts.Folder); !os.IsNotExist(err) {
			return nil
		}

		// Don't run `git clone` or checkout if we're already in a git repository
		// An error will not be returned (and the word true will be printed) if we're in a git repository.
		if _, err := osexec.Command("git", "rev-parse", "--is-inside-git-dir").CombinedOutput(); err == nil {
			plog.Warnln("Skipping git clone because we're already in a git repository...")
			return nil
		}

		if err := exec.RunCommand(ctx, aopts.Stdout, aopts.Stderr, cmd, cloneArgs...); err != nil {
			return err
		}

		if err := exec.RunCommandAt(ctx, aopts.Stdout, aopts.Stderr, opts.Folder, cmd, checkoutArgs...); err != nil {
			return err
		}

		return nil
	}
}

func Clone(sw shipwright.Shipwright, depth int) pipeline.Step {
	return pipeline.NewStep(clone(sw, depth)).
		WithArguments(
			pipeline.ArgumentCommitRef,
			pipeline.ArgumentRemoteURL,
			pipeline.ArgumentWorkingDir,
		).
		WithImage(
			plumbing.SubImage("git", sw.Opts.Version),
		)
}
