package swfs_test

import (
	"os"
	"testing"

	"github.com/grafana/scribe/swfs"
)

func TestExtract(t *testing.T) {
	tmp := t.TempDir()
	fs := os.DirFS("testdata")

	if err := swfs.CopyFS(fs, tmp); err != nil {
		t.Fatal(err)
	}

	copied := os.DirFS(tmp)
	equal, err := swfs.Equal(fs, copied)
	if err != nil {
		t.Fatal(err)
	}

	if !equal {
		t.Fatal("Expected copied filesystem to equal the original one")
	}
}
