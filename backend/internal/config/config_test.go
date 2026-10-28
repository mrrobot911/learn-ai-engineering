package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars
	t.Setenv("HOST", "")
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error with defaults, got %v", err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected environment development, got %s", cfg.Environment)
	}
	if cfg.Addr() != "0.0.0.0:8080" {
		t.Errorf("expected addr 0.0.0.0:8080, got %s", cfg.Addr())
	}
}

func TestLoadMissingRequiredProduction(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL in production")
	}
	if err.Error() != "DATABASE_URL is required in production environment" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadMissingRequiredTest(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL in test")
	}
	if err.Error() != "DATABASE_URL is required in test environment" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadInvalidEnv(t *testing.T) {
	t.Setenv("ENV", "staging")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid environment")
	}
	if err.Error() != "ENV must be development, test or production" {
		t.Errorf("unexpected error message: %v", err)
	}
}
