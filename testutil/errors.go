package testutil

import (
	"errors"
	"testing"
)

func EnsureError(t *testing.T, err, expect error) {
	t.Helper()
	if err == nil && expect != nil {
		t.Fatal("Expected error but none was received")
	}

	if expect == nil && err != nil {
		t.Fatalf("Expected no error but received '%s'", err.Error())
	}

	if !errors.Is(err, expect) {
		t.Fatalf("Expected error '%s' but received '%s'", expect.Error(), err.Error())
	}
}
