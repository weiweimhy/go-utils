package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strings"
)

// PKCS7Padding 对数据进行 PKCS7 填充
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7UnPadding 移除 PKCS7 填充
func PKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	unpadding := int(data[length-1])
	if unpadding > length || unpadding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:length-unpadding], nil
}

// AesEncryptCBC 使用 AES CBC 模式加密，返回大写 HEX 字符串
func AesEncryptCBC(plain string, key, iv []byte) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid key length: must be 16, 24 or 32 bytes")
	}
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid iv length: must be %d bytes", aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := PKCS7Padding([]byte(plain), aes.BlockSize)
	encrypted := make([]byte, len(plaintext))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, plaintext)

	return strings.ToUpper(hex.EncodeToString(encrypted)), nil
}

// AesDecryptCBC 使用 AES CBC 模式解密 HEX 字符串
func AesDecryptCBC(cipherHex string, key, iv []byte) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid key length: must be 16, 24 or 32 bytes")
	}
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid iv length: must be %d bytes", aes.BlockSize)
	}

	cipherBytes, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}

	if len(cipherBytes)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherBytes))
	mode.CryptBlocks(decrypted, cipherBytes)

	result, err := PKCS7UnPadding(decrypted)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

