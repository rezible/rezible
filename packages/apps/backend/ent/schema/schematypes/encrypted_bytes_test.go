package schematypes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var testEncryptionKey = bytes.Repeat([]byte{1}, 32)

func TestSetFieldEncryptionKeyIsSetOnce(t *testing.T) {
	require.NoError(t, SetFieldEncryptionKey(testEncryptionKey))
	require.NoError(t, SetFieldEncryptionKey(bytes.Clone(testEncryptionKey)), "same key")
	require.Error(t, SetFieldEncryptionKey(bytes.Repeat([]byte{2}, 32)), "different key")
}

func TestEncryptedBytesRejectsUnreadableValues(t *testing.T) {
	scanner := EncryptedBytes[[]byte]{}
	require.NoError(t, SetFieldEncryptionKey(testEncryptionKey))
	plaintext := []byte(`{"token":"secret"}`)
	sealedValue, valueErr := scanner.Value(plaintext)
	require.NoError(t, valueErr)
	sealed := sealedValue.([]byte)

	read := func(stored []byte) error {
		_, readErr := scanner.FromValue(&sql.Null[[]byte]{V: stored, Valid: true})
		return readErr
	}
	require.NoError(t, read(sealed))

	tampered := bytes.Clone(sealed)
	tampered[len(tampered)-1] ^= 1
	require.Error(t, read(tampered), "tampered")

	unknownVersion := bytes.Clone(sealed)
	unknownVersion[0] = 2
	require.Error(t, read(unknownVersion), "unknown version")

	require.Error(t, read(sealed[:20]), "short")

	block, blockErr := aes.NewCipher(bytes.Repeat([]byte{2}, 32))
	require.NoError(t, blockErr)
	otherAead, aeadErr := cipher.NewGCM(block)
	require.NoError(t, aeadErr)
	nonce := sealed[1 : 1+otherAead.NonceSize()]
	otherKey := otherAead.Seal(bytes.Clone(sealed[:1+otherAead.NonceSize()]), nonce, plaintext, nil)
	require.Error(t, read(otherKey), "different key")
}

func TestEncryptedBytesStoresNilAsNull(t *testing.T) {
	scanner := EncryptedBytes[[]byte]{}
	require.NoError(t, SetFieldEncryptionKey(testEncryptionKey))

	stored, valueErr := scanner.Value(nil)
	require.NoError(t, valueErr)
	require.Nil(t, stored)

	read, readErr := scanner.FromValue(&sql.Null[[]byte]{})
	require.NoError(t, readErr)
	require.Nil(t, read)
}
