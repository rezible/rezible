package schematypes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync/atomic"

	"entgo.io/ent/schema/field"
)

// encryptedBytesVersion marks a value sealed with AES-256-GCM: the version byte, the nonce, then the ciphertext.
const encryptedBytesVersion byte = 1

type encryptionKey struct {
	key  []byte
	aead cipher.AEAD
}

var currentEncryptionKey atomic.Pointer[encryptionKey]

// SetFieldEncryptionKey sets the 32-byte key that EncryptedBytes fields are sealed under. Ent value scanners are
// static, so the key is process-global and set once: setting the same key again does nothing, and setting a
// different key is an error.
func SetFieldEncryptionKey(key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, blockErr := aes.NewCipher(key)
	if blockErr != nil {
		return fmt.Errorf("create cipher: %w", blockErr)
	}
	aead, aeadErr := cipher.NewGCM(block)
	if aeadErr != nil {
		return fmt.Errorf("create gcm: %w", aeadErr)
	}
	if currentEncryptionKey.CompareAndSwap(nil, &encryptionKey{key: bytes.Clone(key), aead: aead}) {
		return nil
	}
	if subtle.ConstantTimeCompare(currentEncryptionKey.Load().key, key) != 1 {
		return errors.New("a different encryption key is already set")
	}
	return nil
}

func currentEncryptionAead() (cipher.AEAD, error) {
	current := currentEncryptionKey.Load()
	if current == nil {
		return nil, errors.New("encryption key not set")
	}
	return current.aead, nil
}

// EncryptedBytes stores a bytes field encrypted at rest. A nil value is stored as null and null reads as nil.
type EncryptedBytes[T ~[]byte] struct{}

func (EncryptedBytes[T]) Value(plaintext T) (driver.Value, error) {
	if plaintext == nil {
		return nil, nil
	}
	aead, aeadErr := currentEncryptionAead()
	if aeadErr != nil {
		return nil, aeadErr
	}
	sealed := make([]byte, 1+aead.NonceSize(), 1+aead.NonceSize()+len(plaintext)+aead.Overhead())
	sealed[0] = encryptedBytesVersion
	nonce := sealed[1:]
	rand.Read(nonce)
	return aead.Seal(sealed, nonce, plaintext, nil), nil
}

func (EncryptedBytes[T]) ScanValue() field.ValueScanner {
	return &sql.Null[[]byte]{}
}

func (EncryptedBytes[T]) FromValue(value driver.Value) (T, error) {
	scanned, ok := value.(*sql.Null[[]byte])
	if !ok {
		return nil, fmt.Errorf("unexpected encrypted value type %T", value)
	}
	if !scanned.Valid {
		return nil, nil
	}
	aead, aeadErr := currentEncryptionAead()
	if aeadErr != nil {
		return nil, aeadErr
	}
	sealed := scanned.V
	if len(sealed) < 1+aead.NonceSize()+aead.Overhead() {
		return nil, errors.New("encrypted value too short")
	}
	if sealed[0] != encryptedBytesVersion {
		return nil, fmt.Errorf("unknown encrypted value version %d", sealed[0])
	}
	nonce, ciphertext := sealed[1:1+aead.NonceSize()], sealed[1+aead.NonceSize():]
	plaintext, openErr := aead.Open(nil, nonce, ciphertext, nil)
	if openErr != nil {
		return nil, fmt.Errorf("decrypt value: %w", openErr)
	}
	return T(plaintext), nil
}
