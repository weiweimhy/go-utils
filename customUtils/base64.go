package customUtils

import "encoding/base64"

func GetBase64FromBytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
