package horizon

import (
	"context"

	"github.com/labstack/echo/v4"
)

// TokenService manages JWT token lifecycle
type TokenService[T any] interface {
	// GetToken extracts and validates token from request context
	GetToken(ctx context.Context, c echo.Context) (T, error)

	// CleanToken removes token from response context
	CleanToken(ctx context.Context, c echo.Context)

	// VerifyToken validates a token string and returns claims
	VerifyToken(ctx context.Context, value string) T

	// SetToken creates and sets a new token in response context
	SetToken(ctx context.Context, c echo.Context, claim T) error

	// GenerateToken creates a new signed token with claims
	GenerateToken(ctx context.Context, claims T) (string, error)
}
