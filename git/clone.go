package git

import (
	"os"
	osexec "os/exec"
	"strconv"

	"pkg.grafana.com/shipwright/v1/exec"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

func (c *Client) CloneOpts() (*CloneOpts, error) {
	ref, err := c.Configurer.Value(types.ArgumentCommitRef)
	if err != nil {
		return nil, err
	}

	u, err := c.Configurer.Value(types.ArgumentRemoteURL)
	if err != nil {
		return nil, err
	}

	workDir, err := c.Configurer.Value(types.ArgumentWorkingDir)
	if err != nil {
		return nil, err
	}

	return &CloneOpts{
		Folder: workDir,
		URL:    u,
		Ref:    ref,
	}, nil
}

func (c *Client) clone(depth int) types.StepAction {
	return func(aopts types.ActionOpts) error {
		opts, err := c.CloneOpts()
		if err != nil {
			return err
		}

		plog.Infoln("Got git opts:", opts)
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

		if err := exec.RunCommand(aopts.Stdout, aopts.Stderr, cmd, cloneArgs...); err != nil {
			return err
		}

		if err := exec.RunCommandAt(aopts.Stdout, aopts.Stderr, opts.Folder, cmd, checkoutArgs...); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) Clone(depth int) types.Step {
	return types.NewStep(c.clone(depth)).
		WithArguments(
			types.ArgumentCommitRef,
			types.ArgumentRemoteURL,
			types.ArgumentWorkingDir,
		).
		WithImage(
			plumbing.SubImage("git", c.Opts.Version),
		)
}
