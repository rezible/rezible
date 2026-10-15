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
			Reports: []RecordedOutput{
				{Text: "Checkout may be failing.", NextChecks: []string{"Open the dashboard — metrics are not read"}},
				{Text: "Checkout is failing.", Evidence: []CitedEvidence{{Entities: []RecordedEntity{deployment}}}},
			},
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
		{CheckSpec{Kind: CheckNextChecks, Min: count(1)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.check.id(), func(t *testing.T) {
			require.NoError(t, tc.check.validate())
			require.Equal(t, tc.passes, tc.check.Evaluate("id", records).Passed)
		})
	}

	t.Run("next_checks counts the latest report's", func(t *testing.T) {
		withChecks := Records{Investigation: InvestigationRecords{Reports: []RecordedOutput{
			{Text: "Checkout is failing.", NextChecks: []string{"Open the dashboard — metrics are not read", "Ask payments — they changed the pool"}},
		}}}
		cases := []struct {
			min    int
			passes bool
		}{
			{min: 1, passes: true},
			{min: 2, passes: true},
			{min: 3, passes: false},
		}
		for _, tc := range cases {
			check := CheckSpec{Kind: CheckNextChecks, Min: count(tc.min)}
			require.Equal(t, tc.passes, check.Evaluate("id", withChecks).Passed, "min %d", tc.min)
		}
		require.False(t, CheckSpec{Kind: CheckNextChecks, Min: count(1)}.Evaluate("id", Records{}).Passed, "no report has none")
	})

	t.Run("next_checks without min means at least one", func(t *testing.T) {
		var check CheckSpec
		require.NoError(t, decodeStrict([]byte(`{"kind": "next_checks"}`), &check))
		require.NoError(t, check.validate())
		require.Equal(t, 1, *check.Min)
		require.Error(t, decodeStrict([]byte(`{"kind": "next_checks", "minimum": 1}`), &check), "checks stay strict")
	})

	t.Run("outputs_cite with nothing published", func(t *testing.T) {
		require.False(t, CheckSpec{Kind: CheckOutputsCite}.Evaluate("id", Records{}).Passed)
	})
}
