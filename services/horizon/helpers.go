package horizon

import (
	"errors"
	"math/rand"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func isValidFilePath(p string) error {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("not exist")
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("is dir")
	}
	return nil
}

// IsValidEmail checks if the provided string is a valid email address format
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`)
	return re.MatchString(phoneNumber)
}

func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Create32ByteKey(key []byte) string {
	if len(key) > 32 {
		return string(key[:32])
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return string(padded)
}

func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	if strings.ContainsAny(rawURL, " <>\"") {
		return false
	}

	return true
}

func GenerateRandomDigits(size int) (int, error) {
	if size > 8 {
		return 0, errors.New("size must not exceed 8 digits")
	}
	if size <= 0 {
		return 0, errors.New("size must be a positive integer")
	}

	min := intPow(10, size-1)
	max := intPow(10, size) - 1

	// Use a local source (deterministic if seeded)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min, nil
}

func intPow(a, b int) int {
	result := 1
	for i := 0; i < b; i++ {
		result *= a
	}
	return result
}
