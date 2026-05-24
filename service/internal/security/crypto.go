package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type CredentialCodec struct {
	aead cipher.AEAD
}

func NewCredentialCodec(key string) (CredentialCodec, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return CredentialCodec{}, errors.New("credential encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return CredentialCodec{}, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return CredentialCodec{}, fmt.Errorf("create gcm: %w", err)
	}

	return CredentialCodec{aead: aead}, nil
}

func (codec CredentialCodec) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, codec.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	sealed := codec.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (codec CredentialCodec) Decrypt(ciphertext string) (string, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := codec.aead.NonceSize()
	if len(sealed) <= nonceSize {
		return "", errors.New("ciphertext is too short")
	}

	nonce := sealed[:nonceSize]
	payload := sealed[nonceSize:]
	plaintext, err := codec.aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}
