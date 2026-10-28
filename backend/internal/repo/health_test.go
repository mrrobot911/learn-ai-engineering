package repo

import (
	"context"
	"testing"
)

func TestHealthRepository(t *testing.T) {
	r := NewHealthRepository()
	if err := r.Ping(context.Background()); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
