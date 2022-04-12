package testutil

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

func ReadersEqual(t *testing.T, value io.Reader, expected io.Reader) {
	vScanner := bufio.NewScanner(value)
	eScanner := bufio.NewScanner(expected)
	i := 1
	for eScanner.Scan() {
		if !vScanner.Scan() {
			t.Fatal("expected has more lines than provided reader")
		}

		if !bytes.Equal(eScanner.Bytes(), vScanner.Bytes()) {
			t.Fatalf("[%d] Lines not equal: \n%s\n%s\n", i, string(eScanner.Bytes()), string(vScanner.Bytes()))
		}
		i++
	}

	if vScanner.Scan() {
		t.Fatal("provided reader has more lines than expected")
	}
}
