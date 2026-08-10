package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Encrypt(plaintext []byte, secretKey []byte, aad ...[]byte) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	additional := additionalData(aad)

	return gcm.Seal(nonce, nonce, plaintext, additional), nil
}

func Decrypt(ciphertext []byte, secretKey []byte, aad ...[]byte) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	additional := additionalData(aad)

	return gcm.Open(nil, nonce, ciphertext, additional)
}

func additionalData(aad [][]byte) []byte {
	if len(aad) > 0 {
		return aad[0]
	}
	return nil
}
