package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorWrapsCause(t *testing.T) {
	cause := errors.New("connection refused")
	err := &Error{Code: CodeInternal, Message: "Could not save progress.", Err: cause}

	if got, want := err.Error(), "Could not save progress.: connection refused"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if !errors.Is(err, cause) {
		t.Error("expected the cause to be reachable with errors.Is")
	}

	var target *Error
	if !errors.As(fmt.Errorf("handler: %w", err), &target) {
		t.Error("expected the error to be reachable with errors.As")
	}
}
