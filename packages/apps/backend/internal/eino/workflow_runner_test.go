package eino

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/agents"
	"github.com/rezible/rezible/ent"
)

func TestNewWorkflowRunnerRegistersRegistryHandlers(t *testing.T) {
	reg := agents.NewWorkflowRegistry()

	_, err := NewWorkflowRunner(rez.Config{}, reg, nil, nil)

	require.NoError(t, err)
	incidentRunner, incidentOK := reg.Get(agents.WorkflowKindIncidentContextPack)
	require.True(t, incidentOK)
	require.NotNil(t, incidentRunner)
	alertRunner, alertOK := reg.Get(agents.WorkflowKindAlertInvestigation)
	require.True(t, alertOK)
	require.NotNil(t, alertRunner)
}

func TestWorkflowRunnerHandlerDecodesTypedInput(t *testing.T) {
	reg := agents.NewWorkflowRegistry()
	_, err := NewWorkflowRunner(rez.Config{}, reg, nil, nil)
	require.NoError(t, err)
	handler, ok := reg.Get(agents.WorkflowKindAlertInvestigation)
	require.True(t, ok)
	task := &ent.AgentTask{
		WorkflowInput: map[string]any{
			"subjects": []any{map[string]any{
				"type": "alert",
				"id":   uuid.NewString(),
			}},
			"unexpected": true,
		},
	}

	_, runErr := handler(context.Background(), task, &ent.AgentRun{})

	require.Error(t, runErr)
	require.ErrorContains(t, runErr, "decode alert investigation input")
}
