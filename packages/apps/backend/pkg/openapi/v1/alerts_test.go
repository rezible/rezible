package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertinstance"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

func TestAlertEpisodeIdentityAndWindows(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	entityID := uuid.New()
	episode := &ent.AlertEpisode{ID: uuid.New(), KnowledgeEntityID: &entityID, StartedAt: start, HighestSeverity: schematypes.SignalSeverityCritical}
	empty := AlertEpisodeFromEnt(episode)
	require.Equal(t, episode.ID, empty.Id)
	require.Equal(t, &entityID, empty.Attributes.KnowledgeEntityId)
	require.NotEqual(t, empty.Id, *empty.Attributes.KnowledgeEntityId)
	require.NotNil(t, empty.Attributes.Instances)
	require.Empty(t, empty.Attributes.Instances)
	encoded, marshalErr := json.Marshal(empty)
	require.NoError(t, marshalErr)
	require.Contains(t, string(encoded), `"instances":[]`)
	window := &ent.AlertInstance{
		ID: uuid.New(), InstanceKey: "host-a", GroupingKey: "cluster-a", Labels: map[string]string{"host": "a"},
		Summary: "failed", Severity: schematypes.SignalSeverityCritical, FiredAt: start, LastObservedAt: start,
		EndedAt: &end, EndReason: new(alertinstance.EndReasonTimeout),
	}
	episode.Edges.Instances = []*ent.AlertInstance{window, {ID: uuid.New(), FiredAt: end}}
	converted := AlertEpisodeFromEnt(episode)
	require.Equal(t, "open", converted.Attributes.Status)
	require.Equal(t, AlertInstance{Id: window.ID, Attributes: AlertInstanceAttributes{
		InstanceKey: "host-a", GroupingKey: "cluster-a", Labels: window.Labels, Summary: "failed", Severity: "critical",
		FiredAt: start, LastObservedAt: start, EndedAt: &end, EndReason: new("timeout"),
	}}, converted.Attributes.Instances[0])
	require.Nil(t, converted.Attributes.Instances[1].Attributes.EndedAt)
	require.Nil(t, converted.Attributes.Instances[1].Attributes.EndReason)
	episode.ClosedAt = &end
	require.Equal(t, "closed", AlertEpisodeFromEnt(episode).Attributes.Status)
}
