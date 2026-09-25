package watermill

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/stretchr/testify/require"
)

func TestMessageExecutionContextRoundTripsTenantActor(t *testing.T) {
	ctx := execution.NewTenantContext(t.Context(), 42)
	expected := execution.GetContext(ctx)
	msg := message.NewMessage(uuid.NewString(), []byte("event"))
	msg.SetContext(ctx)
	require.NoError(t, setMessageExecutionContext(msg))

	msg.SetContext(t.Context())
	require.NoError(t, (&MessageQueue{}).restoreMessageContext(msg))
	require.Equal(t, expected, execution.GetContext(msg.Context()))
}
