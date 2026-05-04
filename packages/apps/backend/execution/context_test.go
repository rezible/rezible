package execution

import (
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	userID := uuid.New()
	tenantID := 42
	exec := Context{
		Actor: Actor{
			Kind:     KindUser,
			TenantID: &tenantID,
			UserID:   &userID,
		},
		Auth: &rez.AuthSession{
			Scopes:    []string{"document:read"},
			ExpiresAt: time.Unix(100, 0).UTC(),
		},
		Provenance: Provenance{
			Source:        SourceHTTP,
			RequestID:     "req-1",
			CorrelationID: "corr-1",
			ParentKind:    "job",
			ParentID:      "123",
		},
	}

	encoded, err := exec.Encode()
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, exec, decoded)
}

func TestAnonymousValidateRejectsAuth(t *testing.T) {
	exec := Context{
		Actor: Actor{Kind: KindAnonymous},
		Auth:  &rez.AuthSession{},
	}
	require.Error(t, exec.validate())
}
