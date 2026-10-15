package evals

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedScenariosLoad(t *testing.T) {
	targets, targetsErr := ScenarioTargets()
	require.NoError(t, targetsErr)
	require.NotEmpty(t, targets)

	referenceTime := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	scenario, loadErr := loadScenario(targets[0].Name, referenceTime, newAgentHarnesses(&agentHarnessDeps{}))
	require.NoError(t, loadErr)
	require.NotContains(t, string(scenario.Events[0].Attributes), "{{", "placeholders render inside attributes")
}

func TestParseScenarioRejectsInvalidScenarios(t *testing.T) {
	valid := func() map[string]any {
		return map[string]any{
			"description": "A scenario.",
			"agent":       "investigation",
			"events": []any{map[string]any{
				"ref": "first", "kind": "alert_instance", "resource_ref": "alert:first", "attributes": map[string]any{},
			}},
			"setup":  map[string]any{"situation": map[string]any{"title": "First", "signals": []any{"first"}}},
			"phases": []any{map[string]any{"check": []any{map[string]any{"kind": "report_published"}}}},
		}
	}
	checks := func(s map[string]any) []any {
		return s["phases"].([]any)[0].(map[string]any)["check"].([]any)
	}
	cases := []struct {
		name   string
		change func(map[string]any)
		want   string
	}{
		{"an agent with no harness", func(s map[string]any) { s["agent"] = "chat" }, `agent "chat"`},
		{"an event without a kind", func(s map[string]any) { delete(s["events"].([]any)[0].(map[string]any), "kind") }, "no kind"},
		{"a setup naming no setup event", func(s map[string]any) {
			s["setup"] = map[string]any{"situation": map[string]any{"title": "First", "signals": []any{"later"}}}
		}, `names "later"`},
		{"an unknown check kind", func(s map[string]any) { checks(s)[0] = map[string]any{"kind": "report_is_good"} }, `no check kind "report_is_good"`},
		{"a count with no bound", func(s map[string]any) { checks(s)[0] = map[string]any{"kind": "provider_reads"} }, "no min or max"},
		{"a time that does not parse", func(s map[string]any) {
			s["events"].([]any)[0].(map[string]any)["occurred_at"] = "{{-3 minutes}}"
		}, "time placeholder"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scenario := valid()
			tc.change(scenario)
			encoded, encodeErr := json.Marshal(scenario)
			require.NoError(t, encodeErr)
			_, parseErr := parseScenario(encoded, time.Now(), newAgentHarnesses(&agentHarnessDeps{}))
			require.ErrorContains(t, parseErr, tc.want)
		})
	}
}
