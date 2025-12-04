package customUtils

import (
	"crypto/sha256"
	"encoding/hex"
)

func StringToHash16(s string) string {
	return StringToHash(s)[:16]
}

func StringToHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func BytesToHash16(b []byte) string {
	return BytesToHash(b)[:16]
}

func BytesToHash(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
