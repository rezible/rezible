package projections

import (
	"testing"

	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeIncidentObservedEvent(t *testing.T) {
	ev := &ent.NormalizedEvent{
		Kind: ne.KindIncidentObserved,
		Attributes: IncidentObservedAttributes{
			Title:        "Checkout search lookups timing out",
			Summary:      "Checkout requests are timing out.",
			SeverityName: "SEV-1",
			SeverityRank: 1,
			TypeName:     "Customer Impact",
		}.Encode(),
	}

	incEv, err := DecodeEvent[IncidentObservedAttributes](ev)
	require.NoError(t, err)
	assert.Equal(t, "Checkout search lookups timing out", incEv.Attributes.Title)
	assert.Equal(t, "SEV-1", incEv.Attributes.SeverityName)
	assert.Equal(t, 1, incEv.Attributes.SeverityRank)
	assert.Equal(t, "Customer Impact", incEv.Attributes.TypeName)
}

func TestDecodeAlertObservedEvent(t *testing.T) {
	ev := &ent.NormalizedEvent{
		Kind: ne.KindAlertObserved,
		Attributes: AlertObservedAttributes{
			Title:       "Search API response time high",
			Description: "p95 latency is above threshold.",
			Definition:  "avg(last_5m):p95:search.api.response_time > 2000",
		}.Encode(),
	}

	alertEv, err := DecodeEvent[AlertObservedAttributes](ev)
	require.NoError(t, err)
	assert.Equal(t, "Search API response time high", alertEv.Attributes.Title)
	assert.Equal(t, "p95 latency is above threshold.", alertEv.Attributes.Description)
	assert.Equal(t, "avg(last_5m):p95:search.api.response_time > 2000", alertEv.Attributes.Definition)
}
