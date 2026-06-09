package integrations

import (
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeInstallationTargetOptions(t *testing.T) {
	options := []rez.IntegrationInstallationTarget{{
		ExternalRef: "installation-123",
		DisplayName: "Rezible",
		InstallationConfig: map[string]any{
			"org":             "rezible",
			"installation_id": int64(123),
		},
	}}

	encoded, err := EncodeInstallationTargetOptions(options)
	require.NoError(t, err)
	require.Len(t, encoded, 1)

	assert.Equal(t, "installation-123", encoded[0]["ExternalRef"])
	assert.Equal(t, "Rezible", encoded[0]["DisplayName"])
	assert.Equal(t, map[string]any{
		"org":             "rezible",
		"installation_id": int64(123),
	}, encoded[0]["InstallationConfig"])
}

func TestDecodeInstallationTargetOptions(t *testing.T) {
	options := []map[string]any{{
		"ExternalRef": "installation-123",
		"DisplayName": "Rezible",
		"InstallationConfig": map[string]any{
			"org":             "rezible",
			"installation_id": int64(123),
		},
	}}

	decoded, err := DecodeInstallationTargetOptions(options)
	require.NoError(t, err)

	assert.Equal(t, []rez.IntegrationInstallationTarget{{
		ExternalRef: "installation-123",
		DisplayName: "Rezible",
		InstallationConfig: map[string]any{
			"org":             "rezible",
			"installation_id": int64(123),
		},
	}}, decoded)
}
