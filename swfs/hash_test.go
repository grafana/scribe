package swfs_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/grafana/scribe/swfs"
)

func TestHashFile(t *testing.T) {
	file := `{
  "example": "json",
  "file": []
}`

	expect := "c443cfbc0ed9c6097346eed4fe6581999c82032f2d81065e095fd407a1d20fd7"

	buf := bytes.NewBufferString(file)
	b, err := swfs.HashFile(buf)
	if err != nil {
		t.Fatal(err)
	}

	hash := hex.EncodeToString(b)
	if hash != expect {
		t.Fatalf("Unexpected result from hashfile:\nExpected: '%s'\nReceived: '%s'", expect, hash)
	}
}

func TestEncodeDir(t *testing.T) {
	dir := filepath.Clean("testdata")

	b, err := swfs.HashDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}

	hash := sha256.New()
	hash.Sum([]byte("a.json"))
	hash.Sum([]byte("e8f1fae1d192acff9666ffb429757fb60cb92dfa39e3ac074777fd01e1bfabbf"))
	hash.Sum([]byte("b.json"))
	hash.Sum([]byte("497ec934da3f4dc5708e4be58a11f72224b23127b8b402256c114a892ae2aba2"))
	hash.Sum([]byte("c/c.json"))
	hash.Sum([]byte("bbd82e48900b9f9bbe1a00eca6a9ec646eb7126a442dc60b6dd0255de6abd48c"))

	expect := hash.Sum(nil)

	if !bytes.Equal(b, expect) {
		t.Fatalf("Unexpected result from HashDirectory:\nExpected: '%x'\nReceived: '%x'", expect, b)
	}
}
