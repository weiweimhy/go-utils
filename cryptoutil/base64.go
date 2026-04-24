package cryptoutil

import "encoding/base64"

// Base64FromBytes encodes bytes with standard base64 encoding.
func Base64FromBytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
