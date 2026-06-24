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
