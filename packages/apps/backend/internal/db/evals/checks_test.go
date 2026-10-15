package evals

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCheckKinds(t *testing.T) {
	latestTurn := uuid.New()
	alert := RecordedEntity{Kind: "alert", Name: "CheckoutHighErrorRate"}
	deployment := RecordedEntity{Kind: "deployment", Properties: map[string]any{"version": "checkout-api@2041"}}
	read := RecordedRead{Tool: "search_logs", Service: "payments-api", Outcome: "succeeded"}
	records := Records{
		Reads: []RecordedRead{read},
		Investigation: InvestigationRecords{
			Reports:               []RecordedOutput{{Text: "Checkout is failing.", Evidence: []CitedEvidence{{Entities: []RecordedEntity{deployment}}}}},
			Findings:              []RecordedOutput{{Title: "Errors", Evidence: []CitedEvidence{{Entities: []RecordedEntity{alert}}}}},
			Hypotheses:            []RecordedOutput{{Title: "Upstream timeout"}},
			Answers:               []RecordedOutput{{Title: "Search is fine", TurnID: latestTurn, Reads: []RecordedRead{read}}},
			LatestCompletedTurnID: latestTurn,
		},
	}
	count := func(value int) *int { return &value }

	cases := []struct {
		check  CheckSpec
		passes bool
	}{
		{CheckSpec{Kind: CheckReportPublished}, true},
		{CheckSpec{Kind: CheckAnswerPublished}, true},
		{CheckSpec{Kind: CheckOutputCount, Output: "hypothesis", Max: count(0)}, false},
		{CheckSpec{Kind: CheckOutputsCite}, false},
		{CheckSpec{Kind: CheckCitesEntity, EntityKind: "alert", Name: "CheckoutHighErrorRate"}, true},
		{CheckSpec{Kind: CheckCitesEntity, EntityKind: "deployment", Property: map[string]any{"version": "checkout-api@2040"}}, false},
		{CheckSpec{Kind: CheckProviderReads, Service: "payments-api", Min: count(1)}, true},
		{CheckSpec{Kind: CheckCitesProviderRead, Tool: "count_logs"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.check.id(), func(t *testing.T) {
			require.NoError(t, tc.check.validate())
			require.Equal(t, tc.passes, tc.check.Evaluate("id", records).Passed)
		})
	}

	t.Run("outputs_cite with nothing published", func(t *testing.T) {
		require.False(t, CheckSpec{Kind: CheckOutputsCite}.Evaluate("id", Records{}).Passed)
	})
}
