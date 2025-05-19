package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lands-horizon/horizon-server/horizon"
)

func main() {
	ctx := context.Background()

	// Initialize SecurityUtils with configuration
	sec := horizon.NewSecurityUtils(
		64*1024, // memory in KB (e.g. 64MB)
		3,       // iterations
		2,       // parallelism
		16,      // salt length in bytes
		32,      // key length in bytes
	)

	// Generate UUID
	uuid, err := sec.GenerateUUID(ctx)
	if err != nil {
		log.Fatalf("Error generating UUID: %v", err)
	}
	fmt.Println("Generated UUID:", uuid)

	// Hash password
	password := "mySecurePassword123"
	hashedPassword, err := sec.HashPassword(ctx, password)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	fmt.Println("Hashed Password:", hashedPassword)

	// Verify password
	isValid, err := sec.VerifyPassword(ctx, hashedPassword, password)
	if err != nil {
		log.Fatalf("Error verifying password: %v", err)
	}
	fmt.Println("Password is valid:", isValid)

	// Encrypt plaintext
	plaintext := "Secret Message"
	key := "dummy-key" // Not used in your current implementation
	encrypted, err := sec.Encrypt(ctx, plaintext, key)
	if err != nil {
		log.Fatalf("Error encrypting: %v", err)
	}
	fmt.Println("Encrypted:", encrypted)

	// Decrypt ciphertext
	decrypted, err := sec.Decrypt(ctx, encrypted, key)
	if err != nil {
		log.Fatalf("Error decrypting: %v", err)
	}
	fmt.Println("Decrypted:", decrypted)

}
