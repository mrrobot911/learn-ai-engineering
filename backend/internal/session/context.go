// Package session carries the authenticated session resolved from the session
// cookie through the request context.
package session

import (
	"context"

	"github.com/mrrobot911/learn-ai-engineering/internal/service"
)

// CookieName is the name of the session cookie defined by the API contract.
const CookieName = "session"

type contextKey struct{}

// With returns a context that carries the authenticated session.
func With(ctx context.Context, auth service.AuthSession) context.Context {
	return context.WithValue(ctx, contextKey{}, auth)
}

// From returns the authenticated session carried by ctx, if any.
func From(ctx context.Context) (service.AuthSession, bool) {
	auth, ok := ctx.Value(contextKey{}).(service.AuthSession)
	return auth, ok
}
