package oidc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type cookieWriter struct {
	path string
	aead cipher.AEAD
}

func newCookieWriter(secret []byte) (*cookieWriter, error) {
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

	return &cookieWriter{
		path: "/api",
		aead: aead,
	}, nil
}

func (c *cookieWriter) write(w http.ResponseWriter, name string, value any, maxAge time.Duration) error {
	encoded, encErr := c.encode(value)
	if encErr != nil {
		return fmt.Errorf("encode value: %w", encErr)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    encoded,
		Path:     c.path,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (c *cookieWriter) read(r *http.Request, name string, dst any) error {
	cookie, cookieErr := r.Cookie(name)
	if cookieErr != nil {
		return cookieErr
	}
	return c.decode(cookie.Value, dst)
}

func (c *cookieWriter) clear(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     c.path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *cookieWriter) encode(v any) (string, error) {
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

func (c *cookieWriter) decode(encoded string, dst any) error {
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

func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, randErr := rand.Read(b); randErr != nil {
		return "", randErr
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func createRandomValues() (string, string, error) {
	state, stateErr := randomURLSafe(32)
	nonce, nonceErr := randomURLSafe(32)
	if randErr := errors.Join(stateErr, nonceErr); randErr != nil {
		return "", "", fmt.Errorf("creating random state: %w", randErr)
	}
	return state, nonce, nil
}
