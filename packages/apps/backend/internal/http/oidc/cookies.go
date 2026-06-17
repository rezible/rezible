package oidc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

type cookieCodec struct {
	aead cipher.AEAD
}

func newCookieCodec(secret []byte) (*cookieCodec, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("secret must be at least 32 bytes")
	}

	// Derive fixed 32-byte AES-256 key from config secret.
	key := sha256.Sum256(secret)

	block, cipherErr := aes.NewCipher(key[:])
	if cipherErr != nil {
		return nil, fmt.Errorf("aes: %w", cipherErr)
	}

	aead, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return nil, fmt.Errorf("gcm: %w", gcmErr)
	}

	return &cookieCodec{aead: aead}, nil
}

func (c *cookieCodec) encode(v any) (string, error) {
	plaintext, jsonErr := json.Marshal(v)
	if jsonErr != nil {
		return "", fmt.Errorf("json encode: %w", jsonErr)
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, readErr := io.ReadFull(rand.Reader, nonce); readErr != nil {
		return "", readErr
	}
	ciphertext := c.aead.Seal(nil, nonce, plaintext, nil)

	b := bytes.NewBuffer(nonce)
	b.Write(ciphertext)
	return base64.RawURLEncoding.EncodeToString(b.Bytes()), nil
}

func (c *cookieCodec) decode(encoded string, dst any) error {
	raw, decodeErr := base64.RawURLEncoding.DecodeString(encoded)
	if decodeErr != nil {
		return fmt.Errorf("decode: %w", decodeErr)
	}

	nonceSize := c.aead.NonceSize()
	if len(raw) < nonceSize {
		return fmt.Errorf("invalid cookie payload")
	}

	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]
	plaintext, decryptErr := c.aead.Open(nil, nonce, ciphertext, nil)
	if decryptErr != nil {
		return fmt.Errorf("decrypt: %w", decryptErr)
	}

	return json.Unmarshal(plaintext, dst)
}
