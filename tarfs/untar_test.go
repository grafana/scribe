package tarfs_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/scribe/tarfs"
)

func ensureEqualFS(t *testing.T, a fs.FS, b fs.FS) {
	t.Helper()

	fs.WalkDir(a, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		info, err := fs.Stat(b, path)
		if errors.Is(err, fs.ErrNotExist) {
			t.Fatal("Expected file or folder at", path, "but was not found", err)
		}
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			// Nothing left to do with directories
			return nil
		}
		if _, err := b.Open(path); err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

func TestUntar(t *testing.T) {
	dir := os.DirFS("testdir")
	buf := bytes.NewBuffer(nil)
	if err := tarfs.Write(buf, dir); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	out := filepath.Join(tmp, "testdir")
	if err := tarfs.Untar(out, buf); err != nil {
		t.Fatal(err)
	}

	ensureEqualFS(t, dir, os.DirFS(out))
}
