package v1

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent"
	apr "github.com/rezible/rezible/ent/agentproviderread"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func TestAgentProviderReadFromEnt(t *testing.T) {
	earliest := time.Date(2026, 10, 15, 9, 57, 0, 0, time.UTC)
	value := 12.0
	read := &ent.AgentProviderRead{
		ID:       uuid.New(),
		ToolName: rezai.SearchLogsTool.Name(),
		Status:   apr.StatusSucceeded,
		Input:    []byte(`{"start":"2026-10-15T09:00:00Z","end":"2026-10-15T10:00:00Z"}`),
		Read:     []byte(`{"query":"{service=\"checkout-api\"}","source_url":"https://logs.example/explore","limits":[]}`),
		Highlight: []schematypes.ProviderReadItem{
			{Ref: "1", Time: time.Date(2026, 10, 15, 9, 59, 0, 0, time.UTC), Text: "upstream timeout", Labels: map[string]string{"pod": "a"}, Count: 3, Earliest: &earliest},
			{Ref: "4", Time: time.Date(2026, 10, 15, 9, 50, 0, 0, time.UTC), Value: &value},
		},
		Edges: ent.AgentProviderReadEdges{
			ServiceEntity: &ent.KnowledgeEntity{State: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout-api"}},
			Integration:   &ent.Integration{Provider: "provider-a", DisplayName: "Provider A"},
		},
	}

	converted := AgentProviderReadFromEnt(read)

	require.Equal(t, read.ID, converted.Id)
	attrs := converted.Attributes
	require.Equal(t, "checkout-api", attrs.ServiceName)
	require.Equal(t, "provider-a", attrs.IntegrationProvider)
	require.Equal(t, "Provider A", attrs.IntegrationDisplayName)
	require.Equal(t, `{service="checkout-api"}`, attrs.Query)
	require.Equal(t, "https://logs.example/explore", attrs.SourceUrl)
	require.True(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC).Equal(attrs.Start))
	require.Len(t, attrs.Highlight, 2)
	require.Equal(t, "upstream timeout", attrs.Highlight[0].Text)
	require.Equal(t, 3, attrs.Highlight[0].Count)
	require.Equal(t, map[string]string{"pod": "a"}, attrs.Highlight[0].Labels)
	require.True(t, earliest.Equal(*attrs.Highlight[0].Earliest))
	require.Equal(t, "4", attrs.Highlight[1].Ref)
	require.Equal(t, 12.0, *attrs.Highlight[1].Value)
	require.Equal(t, 1, attrs.Highlight[1].Count, "a single item counts one")
	require.Empty(t, attrs.Highlight[1].Labels)
}
