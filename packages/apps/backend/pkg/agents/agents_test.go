package agents

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent"
)

type examplePayload struct {
	FooBar string `json:"foo_bar" validate:"required"`
}

func exampleWorkflowRunner(context.Context, *ent.AgentTask, *ent.AgentRun) (*RunResult, error) {
	return &RunResult{Content: "ok"}, nil
}

func TestEncodeInputUsesJSONTags(t *testing.T) {
	encoded, err := EncodeInput(examplePayload{FooBar: "baz"})

	require.NoError(t, err)
	assert.Equal(t, "baz", encoded["foo_bar"])
	assert.NotContains(t, encoded, "FooBar")
}

func TestEncodeWorkflowInputEncodesSubjectRefs(t *testing.T) {
	alertID := uuid.New()

	encoded, err := EncodeInput(AlertInvestigationInput{
		Subjects: []SubjectRef{{
			Type: "alert",
			ID:   alertID,
		}},
	})

	require.NoError(t, err)
	decoded, decodeErr := DecodeInput[AlertInvestigationInput](encoded)
	require.NoError(t, decodeErr)
	assert.Equal(t, alertID, FindSubjectRefID(decoded.Subjects, "alert"))
}

func TestDecodeInputRejectsUnknownFields(t *testing.T) {
	_, err := DecodeInput[examplePayload](map[string]any{
		"foo_bar":    "baz",
		"unexpected": true,
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown field")
}

func TestDecodeInputRejectsMissingRequiredFields(t *testing.T) {
	_, err := DecodeInput[examplePayload](map[string]any{
		"foo_bar": "",
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "required")
}

func TestValidateWorkflowInputAcceptsAlertInvestigation(t *testing.T) {
	alertID := uuid.New()

	err := ValidateWorkflowInput(WorkflowKindAlertInvestigation, mustEncodeInput(AlertInvestigationInput{
		Subjects: []SubjectRef{{
			Type: "alert",
			ID:   alertID,
		}},
	}))

	require.NoError(t, err)
}

func TestValidateWorkflowInputAcceptsIncidentContextPack(t *testing.T) {
	incidentID := uuid.New()

	err := ValidateWorkflowInput(WorkflowKindIncidentContextPack, mustEncodeInput(IncidentContextPackInput{
		Subjects: []SubjectRef{{
			Type: "incident",
			ID:   incidentID,
		}},
	}))

	require.NoError(t, err)
}

func TestValidateWorkflowInputRejectsMissingWorkflowSubject(t *testing.T) {
	err := ValidateWorkflowInput(WorkflowKindIncidentContextPack, mustEncodeInput(IncidentContextPackInput{
		Subjects: []SubjectRef{{
			Type: "alert",
			ID:   uuid.New(),
		}},
	}))

	require.Error(t, err)
	assert.ErrorContains(t, err, "incident subject")
}

func TestValidateWorkflowInputRejectsUnknownWorkflowKind(t *testing.T) {
	err := ValidateWorkflowInput(WorkflowKind("unknown"), map[string]any{})

	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown agent workflow kind")
}

func TestValidateWorkflowInputRejectsUnknownFields(t *testing.T) {
	err := ValidateWorkflowInput(WorkflowKindAlertInvestigation, map[string]any{
		"subjects": []any{map[string]any{
			"type": "alert",
			"id":   uuid.NewString(),
		}},
		"unexpected": true,
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown field")
}

func TestWorkflowRegistryRegistersRunners(t *testing.T) {
	reg := NewWorkflowRegistry()
	reg.Register(WorkflowKindAlertInvestigation, exampleWorkflowRunner)

	runner, ok := reg.Get(WorkflowKindAlertInvestigation)

	require.True(t, ok)
	result, err := runner(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Content)
}

func mustEncodeInput[T any](input T) map[string]any {
	encoded, err := EncodeInput(input)
	if err != nil {
		panic(err)
	}
	return encoded
}

func TestValidateRunCitationInputRequiresExactlyOneTarget(t *testing.T) {
	err := ValidateRunCitationInput(RunCitationInput{
		CitationKind:      "primary_subject",
		DomainEntityType:  "alert",
		DomainEntityID:    uuid.New(),
		KnowledgeEntityID: uuid.New(),
		Summary:           "Alert",
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "exactly one target")
}

func TestValidateRunCitationInputRequiresDomainTypeAndID(t *testing.T) {
	err := ValidateRunCitationInput(RunCitationInput{
		CitationKind:     "primary_subject",
		DomainEntityType: "alert",
		Summary:          "Alert",
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "type and id")
}

func TestValidateRunCitationInputRequiresSummary(t *testing.T) {
	err := ValidateRunCitationInput(RunCitationInput{
		CitationKind:     "primary_subject",
		DomainEntityType: "alert",
		DomainEntityID:   uuid.New(),
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "summary")
}

func TestRedactPayloadRecurses(t *testing.T) {
	redacted := RedactPayload(map[string]any{
		"accessToken": "secret",
		"nested": map[string]any{
			"authorization": "Bearer abc",
		},
		"items": []any{map[string]any{
			"password": "pw",
		}},
	})

	assert.Equal(t, "redacted", redacted["accessToken"])
	assert.Equal(t, "redacted", redacted["nested"].(map[string]any)["authorization"])
	assert.Equal(t, "redacted", redacted["items"].([]any)[0].(map[string]any)["password"])
}
