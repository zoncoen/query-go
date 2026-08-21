package query

import (
	"errors"
	"testing"
)

func TestNotFoundError(t *testing.T) {
	err := error(&NotFoundError{Query: ".a.b[1]", FailedAt: ".a.b"})
	if got, expect := err.Error(), `".a.b[1]" not found`; got != expect {
		t.Errorf("expected %q but got %q", expect, got)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Error("expected errors.Is(err, ErrNotFound) to be true")
	}
	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatal("expected errors.As to match *NotFoundError")
	}
	if nfe.FailedAt != ".a.b" {
		t.Errorf("unexpected FailedAt: %q", nfe.FailedAt)
	}
}
