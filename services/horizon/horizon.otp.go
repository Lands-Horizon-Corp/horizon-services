package horizon

import "context"

// OTPService manages one-time password generation and validation
type OTPService interface {
	// Generate creates a new OTP code for a key
	Generate(ctx context.Context, key string) (string, error)

	// Verify checks a code against the stored OTP
	Verify(ctx context.Context, key, code string) (bool, error)

	// Revoke invalidates an existing OTP code
	Revoke(ctx context.Context, key string) error
}
