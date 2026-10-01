package execution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	exec := Context{
		ActorKind: KindUser,
		Auth: Auth{
			TenantID: new(42),
			UserID:   new(uuid.New()),
		},
		Provenance: Provenance{
			ID:       "456",
			Source:   SourceHTTP,
			ParentID: new("123"),
		},
	}

	encoded, encErr := exec.Encode()
	require.NoError(t, encErr)

	decoded, restErr := DecodeContext(encoded)
	require.NoError(t, restErr)
	require.Equal(t, exec, decoded)
}

func TestAnonymousValidateRejectsAuth(t *testing.T) {
	exec := Context{
		ActorKind: KindAnonymous,
		Auth:      Auth{UserID: new(uuid.New())},
	}
	require.Error(t, exec.validate())
}

func TestTenantContextReplacesParentAuthAndRoundTrips(t *testing.T) {
	parent := SetContext(t.Context(), Context{
		ActorKind: KindUser,
		Auth: Auth{
			TenantID:            new(7),
			UserID:              new(uuid.New()),
			AgentSessionID:      new(uuid.New()),
			ImpersonatingUserID: new(uuid.New()),
			Scopes:              []string{"admin"},
		},
		Provenance: Provenance{
			ID:       "parent-id",
			Source:   SourceJob,
			ParentID: new("root-id"),
		},
	})

	ctx := NewTenantContext(parent, 42)
	got := GetContext(ctx)
	require.Equal(t, KindTenant, got.ActorKind)
	require.Equal(t, Auth{TenantID: new(42)}, got.Auth)
	require.Equal(t, GetContext(parent).Provenance, got.Provenance)
	require.False(t, got.IsSystem())
	_, hasUser := got.UserID()
	require.False(t, hasUser)

	encoded, encodeErr := got.Encode()
	require.NoError(t, encodeErr)
	decoded, decodeErr := DecodeContext(encoded)
	require.NoError(t, decodeErr)
	require.Equal(t, got, decoded)
}

func TestTenantValidationRequiresTenantIDWithoutRequiringPositiveID(t *testing.T) {
	require.Error(t, (Context{ActorKind: KindTenant}).validate())
	require.NoError(t, (Context{ActorKind: KindTenant, Auth: Auth{TenantID: new(0)}}).validate())
	require.Error(t, (Context{ActorKind: KindTenant, Auth: Auth{
		TenantID: new(42),
		UserID:   new(uuid.New()),
	}}).validate())
}

func TestDecodeKeepsStoredSystemKindExplicit(t *testing.T) {
	legacySystem := Context{ActorKind: KindSystem, Auth: Auth{TenantID: new(42)}}
	encoded, encodeErr := legacySystem.Encode()
	require.NoError(t, encodeErr)
	decoded, decodeErr := DecodeContext(encoded)
	require.NoError(t, decodeErr)
	require.Equal(t, KindSystem, decoded.ActorKind)
	require.True(t, decoded.IsSystem())
	require.Equal(t, legacySystem, decoded)
}
