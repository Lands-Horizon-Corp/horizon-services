package horizon

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

/*

type OwnerClaim struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
	jwt.RegisteredClaims
}

func (c OwnerClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key")

// Owner tokens
ownerTokenService := NewTokenService[OwnerClaim]("owner_token", jwtSecret)

// Member tokens
memberTokenService := NewTokenService[MemberClaim]("member_token", jwtSecret)
*/

type JWTClaims interface {
	jwt.Claims
	GetRegisteredClaims() *jwt.RegisteredClaims
}

// TokenService manages JWT token lifecycle
type TokenService[T JWTClaims] interface {
	// GetToken extracts and validates token from request context
	GetToken(ctx context.Context, c echo.Context) (*T, error)

	// CleanToken removes token from response context
	CleanToken(ctx context.Context, c echo.Context)

	// VerifyToken validates a token string and returns claims
	VerifyToken(ctx context.Context, value string) (*T, error)

	// SetToken creates and sets a new token in response context
	SetToken(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error

	// GenerateToken creates a new signed token with claims
	GenerateToken(ctx context.Context, claims T, expiry time.Duration) (string, error)
}

type HorizonTokenService[T JWTClaims] struct {
	name   string
	secret []byte
}

func NewTokenService[T JWTClaims](value *HorizonTokenService[T]) TokenService[T] {
	return value
}

// CleanToken implements TokenService.
func (h *HorizonTokenService[T]) CleanToken(ctx context.Context, c echo.Context) {
	cookie := &http.Cookie{
		Name:     h.name,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// GenerateToken implements TokenService.
func (h *HorizonTokenService[T]) GenerateToken(ctx context.Context, claims T, expiry time.Duration) (string, error) {
	now := time.Now()
	rc := claims.GetRegisteredClaims()

	if rc.NotBefore == nil {
		rc.NotBefore = jwt.NewNumericDate(now)
	}
	if rc.IssuedAt == nil {
		rc.IssuedAt = jwt.NewNumericDate(now)
	}
	if rc.ExpiresAt == nil {
		rc.ExpiresAt = jwt.NewNumericDate(now.Add(expiry))
	}
	if rc.Subject == "" {
		rc.Subject = rc.ID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.secret)
	if err != nil {
		return "", eris.Wrap(err, "signing token failed")
	}
	return base64.StdEncoding.EncodeToString([]byte(signed)), nil
}

// GetToken implements TokenService.
func (h *HorizonTokenService[T]) GetToken(ctx context.Context, c echo.Context) (*T, error) {
	cookie, err := c.Cookie(h.name)
	if err != nil {
		h.CleanToken(ctx, c)
		return nil, eris.New("authentication token not found")
	}
	rawToken := cookie.Value
	if rawToken == "" {
		h.CleanToken(ctx, c)
		return nil, eris.New("authentication token is empty")
	}

	claim, err := h.VerifyToken(ctx, rawToken)
	if err != nil {
		h.CleanToken(ctx, c)
		return nil, eris.Wrap(err, "invalid or expired authentication token")
	}
	return claim, nil
}

// SetToken implements TokenService.
func (h *HorizonTokenService[T]) SetToken(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	tok, err := h.GenerateToken(ctx, claim, expiry)
	if err != nil {
		return eris.Wrap(err, "GenerateToken failed")
	}
	cookie := &http.Cookie{
		Name:     h.name,
		Value:    tok,
		Path:     "/",
		Expires:  time.Now().Add(expiry),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

// VerifyToken implements TokenService.
func (h *HorizonTokenService[T]) VerifyToken(ctx context.Context, value string) (*T, error) {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, eris.Wrap(err, "invalid base64 token")
	}
	var claims T
	tok, err := jwt.ParseWithClaims(string(raw), claims, func(tkn *jwt.Token) (any, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, eris.Wrap(jwt.ErrSignatureInvalid, "unexpected signing method")
		}
		return h.secret, nil
	})
	if err != nil {
		return nil, eris.Wrap(err, "token parse failed")
	}
	parsedClaims, ok := tok.Claims.(T)
	if !ok || !tok.Valid {
		return nil, eris.New("invalid token")
	}
	if rc := parsedClaims.GetRegisteredClaims(); rc.Issuer != h.name {
		return nil, eris.New("invalid token issuer")
	}
	return &parsedClaims, nil
}
