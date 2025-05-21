package horizon

import (
	"crypto/rand"
	"errors"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
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
