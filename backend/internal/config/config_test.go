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
	t.Setenv("STATIC_DIR", "")

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
	if cfg.StaticDir != "" {
		t.Errorf("expected empty static dir, got %s", cfg.StaticDir)
	}
	if cfg.Addr() != "0.0.0.0:8080" {
		t.Errorf("expected addr 0.0.0.0:8080, got %s", cfg.Addr())
	}
}

func TestLoadStaticDir(t *testing.T) {
	t.Setenv("STATIC_DIR", "/app/static")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.StaticDir != "/app/static" {
		t.Errorf("expected static dir /app/static, got %s", cfg.StaticDir)
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

func TestLoadQuizSubmissionFaultInjectionEnabledInTest(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://example/test")
	t.Setenv("FAULT_INJECT_QUIZ_SUBMISSION_5XX", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cfg.FaultInjectQuizSubmission5xx {
		t.Fatal("expected quiz submission fault injection to be enabled")
	}
}

func TestLoadQuizSubmissionFaultInjectionRejectedInProduction(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example/prod")
	t.Setenv("FAULT_INJECT_QUIZ_SUBMISSION_5XX", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when enabling fault injection in production")
	}
	if err.Error() != "FAULT_INJECT_QUIZ_SUBMISSION_5XX can only be enabled when ENV=test" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadQuizSubmissionFaultInjectionRejectedOutsideTest(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("FAULT_INJECT_QUIZ_SUBMISSION_5XX", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when enabling fault injection outside test environment")
	}
	if err.Error() != "FAULT_INJECT_QUIZ_SUBMISSION_5XX can only be enabled when ENV=test" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadQuizSubmissionFaultInjectionDisabledByDefault(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://example/test")
	t.Setenv("FAULT_INJECT_QUIZ_SUBMISSION_5XX", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.FaultInjectQuizSubmission5xx {
		t.Fatal("expected quiz submission fault injection to be disabled by default")
	}
}
