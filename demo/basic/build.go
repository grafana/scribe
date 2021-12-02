package main

import (
	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
)

func main() {
	sw := shipwright.New(git.EventCommit{})
	repo := sw.Git.Clone()

	sw.Parallel(
		sw.Cache(sw.Yarn.Install(), fs.Cache("node_modules", fs.FileHasChanged("yarn.lock"))),
		sw.Cache(sw.Golang.Modules.Download(), fs.Cache("$GOPATH/pkg", fs.FileHasChanged("go.sum"))),
	)

	// equivalent of `git describe --tags --dirty --always`
	version := repo.Describe(&git.DescribeOpts{
		Tags:   true,
		Dirty:  true,
		Always: true,
	})

	sw.Run(
		// write the version string in the `.version` file.
		sw.FS.ReplaceString(".version", version),
		sw.Make.Target("build"),
		sw.Make.Target("package"),
	)
	sw.Output(shipwright.NewArtifact(
		"example:tarball",
		fs.Glob("bin/*.tar.gz"),
	))
}
