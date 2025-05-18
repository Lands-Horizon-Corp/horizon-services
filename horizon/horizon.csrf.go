package horizon

import (
	"context"

	"github.com/labstack/echo/v4"
)

// CSRFService manages Cross-Site Request Forgery protection
type CSRFService interface {
	// GenerateToken creates a new CSRF token for a session
	GenerateToken(ctx context.Context, sessionID string) (string, error)

	// VerifyToken validates a CSRF token against session ID
	VerifyToken(ctx context.Context, sessionID string, token string) (bool, error)

	// RevokeToken invalidates all CSRF tokens for a session
	RevokeToken(ctx context.Context, sessionID string) error

	// Middleware provides Echo middleware for CSRF protection
	Middleware() echo.MiddlewareFunc
}
