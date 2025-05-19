package horizon_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/lands-horizon/horizon-server/horizon"
	"github.com/stretchr/testify/assert"
)

func setupSecurityUtils() horizon.SecurityUtils {
	return horizon.NewSecurityUtils(
		64*1024, // memory (e.g., 64MB)
		3,       // iterations
		2,       // parallelism
		16,      // salt length in bytes
		32,      // key length in bytes
	)
}

func TestGenerateUUID(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()

	uuid, err := sec.GenerateUUID(ctx)
	assert.NoError(t, err)
	assert.Len(t, uuid, 36) // UUID v4 standard length
}

func TestHashAndVerifyPassword(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()
	password := "MySecurePassword!@#"

	hashed, err := sec.HashPassword(ctx, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)

	isValid, err := sec.VerifyPassword(ctx, hashed, password)
	assert.NoError(t, err)
	assert.True(t, isValid)

	isValidWrong, err := sec.VerifyPassword(ctx, hashed, "WrongPassword")
	assert.NoError(t, err)
	assert.False(t, isValidWrong)
}

func TestEncryptAndDecrypt(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()

	plaintext := "Confidential Info"
	key := "dummy-key" // not used in current Encrypt/Decrypt

	encrypted, err := sec.Encrypt(ctx, plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// base64 decode to check if it's encoded properly
	_, err = base64.StdEncoding.DecodeString(encrypted)
	assert.NoError(t, err)

	decrypted, err := sec.Decrypt(ctx, encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
