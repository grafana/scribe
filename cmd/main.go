// Package main contains the CLI logic for the `scribe` command
// The scribe command's main responsibility is to run a pipeline.
package main

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/grafana/scribe/cmd/commands"
	"github.com/grafana/scribe/cmdutil"
	"github.com/grafana/scribe/plog"
	"github.com/sirupsen/logrus"
)

// Arguments provided at compile-time
var (
	Version = "latest"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func handleSignal(log *logrus.Logger, cmd *exec.Cmd, sig os.Signal) int {
	log.Debugln("Received OS signal", sig.String())

	log.Debugf("Sending pipeline '%s' signal...", sig.String())
	cmd.Process.Signal(sig)

	log.Debugln("Waiting for pipeline to exit...")
	p, err := cmd.Process.Wait()
	if err != nil {
		log.Error(err)
		return 0
	}

	return p.ExitCode()
}

func main() {
	log := plog.New(logrus.InfoLevel)

	log.Debugln("Running version", Version)
	var (
		ctx = context.Background()
	)

	args := commands.MustParseArgs(os.Args[1:])

	cmd := commands.Run(ctx, &commands.RunOpts{
		Version: Version,
		State:   args.State,
		Path:    args.Path,
		Args:    args,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
	})

	var (
		c        = make(chan os.Signal, 1)
		errChan  = make(chan error)
		doneChan = make(chan bool)
	)

	go func(cmd *exec.Cmd) {
		if err := cmd.Run(); err != nil {
			errChan <- err
			return
		}
		doneChan <- true
	}(cmd)

	log.Debugln("Watching for OS signals...")
	cmdutil.NotifySignals(c)

	select {
	case sig := <-c:
		os.Exit(handleSignal(log, cmd, sig))
	case err := <-errChan:
		log.Error(err)
		os.Exit(1)
	case <-doneChan:
		os.Exit(0)
	}
}
