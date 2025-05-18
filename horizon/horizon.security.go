package horizon

import "context"

// SecurityUtils provides cryptographic and security-related functions
type SecurityUtils interface {
	// GenerateUUID creates a new UUIDv4
	GenerateUUID(ctx context.Context) (string, error)

	// HashPassword creates an Argon2 hashed password
	HashPassword(ctx context.Context, password string) (string, error)

	// VerifyPassword compares a password with its hash
	VerifyPassword(ctx context.Context, hash, password string) (bool, error)

	// SanitizeHTML removes potentially dangerous HTML content
	SanitizeHTML(ctx context.Context, input string) (string, error)

	// Encrypt performs AES encryption on plaintext
	Encrypt(ctx context.Context, plaintext, key string) (string, error)

	// Decrypt performs AES decryption on ciphertext
	Decrypt(ctx context.Context, ciphertext, key string) (string, error)
}
