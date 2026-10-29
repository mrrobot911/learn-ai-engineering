package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
)

func serveError(err error) *httptest.ResponseRecorder {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, err)
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	return w
}

func decodeErrorBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	return body
}

func TestWriteErrorMapsEveryCode(t *testing.T) {
	tests := []struct {
		name   string
		code   apperr.Code
		status int
	}{
		{"validation error", apperr.CodeValidation, http.StatusBadRequest},
		{"authentication error", apperr.CodeAuthentication, http.StatusUnauthorized},
		{"forbidden", apperr.CodeForbidden, http.StatusForbidden},
		{"authorization error", apperr.CodeAuthorization, http.StatusForbidden},
		{"not found", apperr.CodeNotFound, http.StatusNotFound},
		{"conflict", apperr.CodeConflict, http.StatusConflict},
		{"internal error", apperr.CodeInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveError(&apperr.Error{Code: tt.code, Message: "Visible message."})

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("expected content-type application/json, got %q", got)
			}

			body := decodeErrorBody(t, w)
			if body["code"] != string(tt.code) {
				t.Errorf("expected code %q, got %v", tt.code, body["code"])
			}
			if body["message"] != "Visible message." {
				t.Errorf("expected message %q, got %v", "Visible message.", body["message"])
			}
			if _, ok := body["details"]; ok {
				t.Errorf("expected no details key, got %v", body["details"])
			}
			if len(body) != 2 {
				t.Errorf("expected exactly code and message, got %v", body)
			}
		})
	}
}

func TestWriteErrorKeepsForbiddenCodesDistinct(t *testing.T) {
	forbidden := serveError(&apperr.Error{Code: apperr.CodeForbidden, Message: "Refresh the CSRF token."})
	authorization := serveError(&apperr.Error{Code: apperr.CodeAuthorization, Message: "Not a maintainer."})

	if forbidden.Code != http.StatusForbidden || authorization.Code != http.StatusForbidden {
		t.Fatalf("expected both responses to be 403, got %d and %d", forbidden.Code, authorization.Code)
	}

	forbiddenBody := decodeErrorBody(t, forbidden)
	authorizationBody := decodeErrorBody(t, authorization)
	if forbiddenBody["code"] != string(apperr.CodeForbidden) {
		t.Errorf("expected code %q, got %v", apperr.CodeForbidden, forbiddenBody["code"])
	}
	if authorizationBody["code"] != string(apperr.CodeAuthorization) {
		t.Errorf("expected code %q, got %v", apperr.CodeAuthorization, authorizationBody["code"])
	}
	if forbiddenBody["code"] == authorizationBody["code"] {
		t.Error("expected the two 403 responses to have different codes")
	}
}

func TestWriteErrorDetails(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		details := map[string]any{"email": "is required", "password": "is too short"}
		w := serveError(&apperr.Error{Code: apperr.CodeValidation, Message: "Check the form.", Details: details})

		body := decodeErrorBody(t, w)
		got, ok := body["details"].(map[string]any)
		if !ok {
			t.Fatalf("expected details object, got %v", body["details"])
		}
		if !reflect.DeepEqual(got, details) {
			t.Errorf("expected details %v, got %v", details, got)
		}
	})

	t.Run("absent", func(t *testing.T) {
		w := serveError(&apperr.Error{Code: apperr.CodeValidation, Message: "Check the form."})

		if _, ok := decodeErrorBody(t, w)["details"]; ok {
			t.Error("expected no details key")
		}
	})

	t.Run("empty", func(t *testing.T) {
		w := serveError(&apperr.Error{Code: apperr.CodeValidation, Message: "Check the form.", Details: map[string]any{}})

		if _, ok := decodeErrorBody(t, w)["details"]; ok {
			t.Error("expected no details key for empty details")
		}
	})
}

func TestWriteErrorFallsBackToDefaultMessage(t *testing.T) {
	w := serveError(&apperr.Error{Code: apperr.CodeNotFound})

	body := decodeErrorBody(t, w)
	message, ok := body["message"].(string)
	if !ok || message == "" {
		t.Errorf("expected a non-empty message, got %v", body["message"])
	}
}

func TestWriteErrorMapsWrappedDomainError(t *testing.T) {
	wrapped := fmt.Errorf("service: %w", &apperr.Error{Code: apperr.CodeConflict, Message: "Already exists."})

	w := serveError(wrapped)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
	if body := decodeErrorBody(t, w); body["code"] != string(apperr.CodeConflict) {
		t.Errorf("expected code %q, got %v", apperr.CodeConflict, body["code"])
	}
}

func TestWriteErrorHidesNonDomainError(t *testing.T) {
	cause := errors.New("query failed for user 42: password_hash=$2a$10$secret and password=hunter2")
	w := serveError(fmt.Errorf("load user: %w", cause))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	body := decodeErrorBody(t, w)
	if body["code"] != string(apperr.CodeInternal) {
		t.Errorf("expected code %q, got %v", apperr.CodeInternal, body["code"])
	}
	if message, ok := body["message"].(string); !ok || message == "" {
		t.Errorf("expected a non-empty message, got %v", body["message"])
	}
	if len(body) != 2 {
		t.Errorf("expected exactly code and message, got %v", body)
	}

	raw := w.Body.String()
	for _, leak := range []string{"load user", "query failed", "user 42", "password", "password_hash", "hunter2", "secret"} {
		if strings.Contains(raw, leak) {
			t.Errorf("error body leaks %q: %s", leak, raw)
		}
	}
}
