package service

import (
	"context"
	"testing"
)

func TestHealthService(t *testing.T) {
	s := NewHealthService()
	if err := s.Check(context.Background()); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
