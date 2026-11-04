package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

// AuthHandler handles registration, login, logout and the current session.
type AuthHandler struct {
	service      *service.AuthService
	secureCookie bool
}

// NewAuthHandler creates an auth handler. secureCookie sets the Secure flag of
// the session cookie and must be true only in production.
func NewAuthHandler(authService *service.AuthService, secureCookie bool) *AuthHandler {
	return &AuthHandler{service: authService, secureCookie: secureCookie}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userSummaryResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type authResponse struct {
	User      userSummaryResponse `json:"user"`
	CSRFToken string              `json:"csrf_token"`
}

type logoutResponse struct {
	OK bool `json:"ok"`
}

// Register creates an account, starts a session and issues a CSRF token.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	auth, err := h.service.Register(r.Context(), request.Email, request.Password)
	if err != nil {
		WriteError(w, err)
		return
	}

	h.setSessionCookie(w, auth)
	writeJSON(w, http.StatusCreated, newAuthResponse(auth))
}

// Login authenticates with email and password, starts a session and issues a
// CSRF token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	auth, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		WriteError(w, err)
		return
	}

	h.setSessionCookie(w, auth)
	writeJSON(w, http.StatusOK, newAuthResponse(auth))
}

// CurrentSession returns the session resolved by the session middleware and
// issues its CSRF token again. A GET does not require the CSRF header.
func (h *AuthHandler) CurrentSession(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}
	writeJSON(w, http.StatusOK, newAuthResponse(auth))
}

// Logout ends the session resolved by the session middleware and clears its
// cookie. The CSRF token is checked by the middleware before this handler.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	if err := h.service.Logout(r.Context(), auth.Session.ID); err != nil {
		WriteError(w, err)
		return
	}

	h.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, logoutResponse{OK: true})
}

func newAuthResponse(auth service.AuthSession) authResponse {
	return authResponse{
		User: userSummaryResponse{
			ID:    auth.User.ID,
			Email: auth.User.Email,
		},
		CSRFToken: auth.Session.CSRFToken,
	}
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, auth service.AuthSession) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    auth.Session.ID,
		Path:     "/",
		Expires:  auth.Session.ExpiresAt,
		MaxAge:   int(service.SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}
