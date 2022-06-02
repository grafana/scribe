package x_test

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/scribe/fs/x"
)

func TestCopy(t *testing.T) {
	var (
		content = `test file`
		tmp     = t.TempDir()
		from    = filepath.Join(tmp, "test-from.txt")
		to      = filepath.Join(tmp, "test-to.txt")
	)

	f, err := os.Create(from)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, content); err != nil {
		t.Fatal(err)
	}

	if err := x.CopyFile(from, to); err != nil {
		t.Fatal(err)
	}

	r, err := os.Open(to)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if content != string(b) {
		t.Fatalf("Copied file did not have expected content.\nExpected: '%s'\nReceived: '%s'", content, string(b))
	}
}
