package integrations

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const webhookTokenBytes = 32

// NewWebhookToken returns a random URL-safe token for an installation's webhook URL and the hash stored for it.
// The token itself is never stored.
func NewWebhookToken() (string, []byte, error) {
	raw := make([]byte, webhookTokenBytes)
	if _, readErr := rand.Read(raw); readErr != nil {
		return "", nil, fmt.Errorf("read random token: %w", readErr)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, WebhookTokenHash(token), nil
}

// WebhookTokenHash is the stored form of a webhook URL token.
func WebhookTokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
