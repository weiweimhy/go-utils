package idutil

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// FormatSequence formats n with an optional prefix and zero padding width.
func FormatSequence(prefix string, n uint64, width int) string {
	if width <= 0 {
		return prefix + strconv.FormatUint(n, 10)
	}
	return fmt.Sprintf("%s%0*d", prefix, width, n)
}

// ParseSequence parses an ID created by FormatSequence.
func ParseSequence(id, prefix string) (uint64, bool) {
	if !strings.HasPrefix(id, prefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(id, prefix)
	if raw == "" {
		return 0, false
	}
	for _, r := range raw {
		if !unicode.IsDigit(r) {
			return 0, false
		}
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	return n, err == nil
}

// RandomHex returns prefix followed by nBytes of crypto-random data encoded as hex.
func RandomHex(prefix string, nBytes int) (string, error) {
	b, err := randomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

// RandomBase64URL returns prefix followed by nBytes of crypto-random data encoded as raw URL-safe base64.
func RandomBase64URL(prefix string, nBytes int) (string, error) {
	b, err := randomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomDigits returns length crypto-random decimal digits without modulo bias.
func RandomDigits(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("idutil: length must be non-negative")
	}
	if length == 0 {
		return "", nil
	}

	digits := make([]byte, 0, length)
	buffer := make([]byte, length)
	for len(digits) < length {
		if _, err := rand.Read(buffer); err != nil {
			return "", err
		}
		for _, value := range buffer {
			// 250 is the greatest multiple of ten below 256, so values in
			// [0, 250) map evenly to decimal digits.
			if value >= 250 {
				continue
			}
			digits = append(digits, '0'+value%10)
			if len(digits) == length {
				break
			}
		}
	}
	return string(digits), nil
}

func randomBytes(nBytes int) ([]byte, error) {
	if nBytes < 0 {
		return nil, fmt.Errorf("idutil: nBytes must be non-negative")
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
