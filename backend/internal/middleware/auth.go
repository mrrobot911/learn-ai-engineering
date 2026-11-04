package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/handler"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

// Session resolves the session cookie into an authenticated session in the
// request context. A missing, unknown or expired session leaves the request
// anonymous; an unexpected lookup error answers 500.
func Session(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(session.CookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			resolved, err := auth.CurrentSession(r.Context(), cookie.Value)
			if err != nil {
				var domainErr *apperr.Error
				if errors.As(err, &domainErr) && domainErr.Code == apperr.CodeAuthentication {
					next.ServeHTTP(w, r)
					return
				}
				handler.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(session.With(r.Context(), resolved)))
		})
	}
}

// CSRF requires a valid session and the matching X-CSRF-Token header. It
// answers 401 without a session and 403 for a missing or wrong token.
func CSRF(next http.Handler) http.Handler {
	return csrf(true, next)
}

// OptionalCSRF checks the token only when a valid session is present, so an
// anonymous request passes through. It never answers 401.
func OptionalCSRF(next http.Handler) http.Handler {
	return csrf(false, next)
}

func csrf(requireSession bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, ok := session.From(r.Context())
		if !ok {
			if !requireSession {
				next.ServeHTTP(w, r)
				return
			}
			handler.WriteError(w, &apperr.Error{
				Code:    apperr.CodeAuthentication,
				Message: "Authentication is required.",
			})
			return
		}

		token := r.Header.Get("X-CSRF-Token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(auth.Session.CSRFToken)) != 1 {
			handler.WriteError(w, &apperr.Error{
				Code:    apperr.CodeForbidden,
				Message: "The request could not be verified. Refresh the page and try again.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
