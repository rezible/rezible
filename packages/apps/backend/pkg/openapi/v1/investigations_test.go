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

func TestInvestigationProviderReadFromEnt(t *testing.T) {
	displayNames := func(name string) string { return "Provider " + name }
	service := &ent.KnowledgeEntity{State: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout-api"}}
	newRead := func(tool string, status apr.Status) *ent.AgentProviderRead {
		return &ent.AgentProviderRead{
			ID:       uuid.New(),
			ToolName: tool,
			Status:   status,
			Input:    []byte(`{"start":"2026-10-15T09:00:00Z","end":"2026-10-15T10:00:00Z"}`),
			Read:     []byte(`{"integration":"logs","query":"{service=\"checkout-api\"}","step":"5m0s","explore_url":"https://logs.example/explore","limits":[]}`),
			Edges:    ent.AgentProviderReadEdges{ServiceEntity: service},
		}
	}

	t.Run("a log search", func(t *testing.T) {
		read := newRead(rezai.SearchLogsTool.Name(), apr.StatusSucceeded)
		read.Result = []byte(`{
			"labels": {"service": "checkout-api"},
			"lines": [
				{"time": "2026-10-15T09:59:00Z", "line": "upstream timeout", "labels": {"pod": "a"}, "count": 3, "earliest": "2026-10-15T09:57:00Z"},
				{"time": "2026-10-15T09:50:00Z", "line": "retrying"}
			]
		}`)

		converted := InvestigationProviderReadFromEnt(read, displayNames)

		require.Equal(t, "checkout-api", converted.Attributes.Summary.ServiceName)
		require.Equal(t, "Provider logs", converted.Attributes.Summary.ProviderDisplayName)
		require.Equal(t, "https://logs.example/explore", converted.Attributes.Summary.ExploreUrl)
		require.Nil(t, converted.Attributes.LogCount)
		require.NotNil(t, converted.Attributes.LogSearch)
		require.Equal(t, map[string]string{"service": "checkout-api"}, converted.Attributes.LogSearch.Labels)
		lines := converted.Attributes.LogSearch.Lines
		require.Len(t, lines, 2)
		require.Equal(t, "upstream timeout", lines[0].Line)
		require.Equal(t, 3, lines[0].Count)
		require.Equal(t, map[string]string{"pod": "a"}, lines[0].Labels)
		require.True(t, time.Date(2026, 10, 15, 9, 57, 0, 0, time.UTC).Equal(*lines[0].Earliest))
		require.Equal(t, 1, lines[1].Count, "a single line counts one")
		require.Empty(t, lines[1].Labels)
	})

	t.Run("a log count", func(t *testing.T) {
		read := newRead(rezai.CountLogsTool.Name(), apr.StatusSucceeded)
		read.Result = []byte(`{"points": [{"time": "2026-10-15T09:55:00Z", "value": 4}, {"time": "2026-10-15T10:00:00Z", "value": 12}]}`)

		converted := InvestigationProviderReadFromEnt(read, displayNames)

		require.Nil(t, converted.Attributes.LogSearch)
		require.NotNil(t, converted.Attributes.LogCount)
		require.Len(t, converted.Attributes.LogCount.Points, 2)
		require.Equal(t, 12.0, converted.Attributes.LogCount.Points[1].Value)
		require.Equal(t, 300, converted.Attributes.LogCount.StepSeconds, "the step comes from the read's record")
	})

	t.Run("a failed read has no result", func(t *testing.T) {
		read := newRead(rezai.SearchLogsTool.Name(), apr.StatusFailed)
		read.Read = nil
		errorText := "dial tcp: connection refused"
		read.Error = &errorText

		converted := InvestigationProviderReadFromEnt(read, displayNames)

		require.Equal(t, "failed", converted.Attributes.Summary.Status)
		require.Nil(t, converted.Attributes.LogSearch)
		require.Nil(t, converted.Attributes.LogCount)
	})

	t.Run("a result stored in the earlier shape is left out", func(t *testing.T) {
		read := newRead(rezai.SearchLogsTool.Name(), apr.StatusSucceeded)
		read.Result = []byte(`{"lines": [{"time": "2026-10-15T09:59:00Z", "line": "upstream timeout"}], "read": {"query": "{}"}}`)

		converted := InvestigationProviderReadFromEnt(read, displayNames)

		require.Equal(t, "checkout-api", converted.Attributes.Summary.ServiceName)
		require.Nil(t, converted.Attributes.LogSearch)
	})
}
