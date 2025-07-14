package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

func DeriveKeyFromSecret(keyString string) []byte {
	hash := sha256.Sum256([]byte(keyString))
	return hash[:]
}

func EncryptToken(aesKey []byte, token string) string {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		fmt.Println(err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println(err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
	}

	encrypted := aesGCM.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(encrypted)
}

func DecryptToken(aesKey []byte, encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", fmt.Errorf("decoded data is empty")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
