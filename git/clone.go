package git

import (
	"net/url"
	"os"
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

	remoteURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	workDir, err := c.Configurer.Value(types.ArgumentWorkingDir)
	if err != nil {
		return nil, err
	}

	return &CloneOpts{
		Folder: workDir,
		URL:    remoteURL,
		Ref:    ref,
	}, nil
}

func (c *Client) clone(depth int) types.StepAction {
	return func() error {
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
			cloneArgs    = []string{"clone", opts.URL.String(), opts.Folder}
			checkoutArgs = []string{"checkout", opts.Ref}
		)

		if depth == 0 {
			cloneArgs = append(cloneArgs, "--depth", strconv.Itoa(depth))
		}

		// Don't run `git clone` or checkout if the working directory already exists
		if _, err := os.Stat(opts.Folder); !os.IsNotExist(err) {
			return nil
		}

		if err := exec.RunCommand(cmd, cloneArgs...); err != nil {
			return err
		}

		if err := exec.RunCommandAt(opts.Folder, cmd, checkoutArgs...); err != nil {
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
