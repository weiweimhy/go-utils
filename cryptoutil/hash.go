package cryptoutil

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex16FromString returns the first 16 hex characters of a SHA-256 hash.
func SHA256Hex16FromString(s string) string {
	return SHA256HexFromString(s)[:16]
}

// SHA256HexFromString returns the SHA-256 hash of a string in hex form.
func SHA256HexFromString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256Hex16FromBytes returns the first 16 hex characters of a SHA-256 hash.
func SHA256Hex16FromBytes(b []byte) string {
	return SHA256HexFromBytes(b)[:16]
}

// SHA256HexFromBytes returns the SHA-256 hash of bytes in hex form.
func SHA256HexFromBytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
