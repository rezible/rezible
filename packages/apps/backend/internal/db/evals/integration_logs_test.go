package evals

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

var logsReferenceTime = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

func logsAt(offset time.Duration) *time.Time {
	return new(logsReferenceTime.Add(offset))
}

func TestValidateLogsRejectsInvalidEntries(t *testing.T) {
	valid := func() ScenarioLog {
		return ScenarioLog{Service: "service:checkout-api", Line: "request completed", At: logsAt(0)}
	}
	cases := []struct {
		name   string
		change func(*ScenarioLog)
		want   string
	}{
		{"no service", func(l *ScenarioLog) { l.Service = " " }, "entry 1: no service"},
		{"no line", func(l *ScenarioLog) { l.Line = "" }, "entry 1: no line"},
		{"at and a series", func(l *ScenarioLog) { l.From, l.To, l.Every = logsAt(-time.Hour), logsAt(0), "5s" }, "both at and"},
		{"neither at nor a series", func(l *ScenarioLog) { l.At = nil }, "neither at nor"},
		{"a series missing every", func(l *ScenarioLog) { l.At, l.From, l.To = nil, logsAt(-time.Hour), logsAt(0) }, "neither at nor"},
		{"every under a second", func(l *ScenarioLog) {
			l.At, l.From, l.To, l.Every = nil, logsAt(-time.Hour), logsAt(0), "500ms"
		}, "under one second"},
		{"to not after from", func(l *ScenarioLog) {
			l.At, l.From, l.To, l.Every = nil, logsAt(0), logsAt(0), "5s"
		}, "to is not after from"},
		{"too many lines", func(l *ScenarioLog) {
			l.At, l.From, l.To, l.Every = nil, logsAt(-3*time.Hour), logsAt(0), "1s"
		}, "more than the 10000 allowed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := valid()
			tc.change(&entry)
			require.ErrorContains(t, validateLogs([]ScenarioLog{entry}), tc.want)
		})
	}
	require.NoError(t, validateLogs([]ScenarioLog{valid()}))
}

func TestScenarioLogSeriesTimes(t *testing.T) {
	series := ScenarioLog{Service: "s", Line: "l", From: logsAt(-10 * time.Second), To: logsAt(0), Every: "4s"}
	require.Equal(t, []time.Time{*logsAt(-10 * time.Second), *logsAt(-6 * time.Second), *logsAt(-2 * time.Second)}, series.times())
	require.Equal(t, 3, series.lineCount())
}

// newLogsInstallation installs the integration with the logs, as the base harness does.
func newLogsInstallation(t *testing.T, logs []ScenarioLog) (rez.InstalledIntegration, error) {
	t.Helper()
	config, encodeErr := json.Marshal(integrationConfig{Logs: logs})
	require.NoError(t, encodeErr)
	return NewIntegration(fixedClock{now: logsReferenceTime}, true).GetInstalledIntegration(&ent.Integration{InstallationConfig: config})
}

func TestInstalledLogs(t *testing.T) {
	logs := []ScenarioLog{
		{Service: "service:checkout-api", Line: "upstream timeout", Labels: map[string]string{"level": "ERROR"},
			From: logsAt(-30 * time.Second), To: logsAt(0), Every: "10s"},
		{Service: "service:checkout-api", Line: "request completed", At: logsAt(-25 * time.Second)},
		{Service: "service:payments-api", Line: "upstream timeout", At: logsAt(-5 * time.Second)},
	}
	installed, installErr := newLogsInstallation(t, logs)
	require.NoError(t, installErr)
	querier, isQuerier := installed.(integrations.LogQuerier)
	require.True(t, isQuerier)
	require.NoError(t, querier.CheckLogsAvailable())
	checkout := integrations.LogService{DisplayName: "checkout-api", Aliases: []string{"other", "service:checkout-api"}}

	t.Run("search filters by alias, range and text, newest first", func(t *testing.T) {
		params := integrations.LogSearchParams{Service: checkout, Start: *logsAt(-25 * time.Second), End: *logsAt(0), Contains: "timeout"}
		result, searchErr := querier.SearchLogs(t.Context(), params)
		require.NoError(t, searchErr)
		require.Equal(t, integrations.LogServiceName{Value: "service:checkout-api", FromAlias: true}, result.Service)
		require.Equal(t, []integrations.LogLine{
			{Time: *logsAt(-10 * time.Second), Line: "upstream timeout", Labels: map[string]string{"level": "ERROR"}},
			{Time: *logsAt(-20 * time.Second), Line: "upstream timeout", Labels: map[string]string{"level": "ERROR"}},
		}, result.Lines)
		require.Equal(t, `service="service:checkout-api" contains="timeout"`, result.Read.Query)
		require.Empty(t, result.Read.Limits)
		require.Empty(t, result.Read.ExploreURL)
		require.Equal(t, logsReferenceTime, result.Read.RanAt)
	})

	t.Run("search records the limit when it cuts", func(t *testing.T) {
		params := integrations.LogSearchParams{Service: checkout, Start: *logsAt(-time.Minute), End: *logsAt(0), Limit: 2}
		result, searchErr := querier.SearchLogs(t.Context(), params)
		require.NoError(t, searchErr)
		require.Len(t, result.Lines, 2)
		require.Equal(t, *logsAt(-10 * time.Second), result.Lines[0].Time)
		require.Equal(t, []string{"reached the limit of 2 lines; 2 more matched"}, result.Read.Limits)
	})

	t.Run("a service without a matching alias reads nothing", func(t *testing.T) {
		search := integrations.LogService{DisplayName: "checkout-api"}
		params := integrations.LogSearchParams{Service: search, Start: *logsAt(-time.Minute), End: *logsAt(0)}
		result, searchErr := querier.SearchLogs(t.Context(), params)
		require.NoError(t, searchErr)
		require.Equal(t, integrations.LogServiceName{Value: "checkout-api"}, result.Service)
		require.Empty(t, result.Lines)
	})

	t.Run("count buckets with zeros", func(t *testing.T) {
		params := integrations.LogCountParams{Service: checkout, Start: *logsAt(-40 * time.Second), End: *logsAt(0), Step: 20 * time.Second}
		result, countErr := querier.CountLogs(t.Context(), params)
		require.NoError(t, countErr)
		require.Equal(t, []integrations.MetricPoint{
			{Time: *logsAt(-40 * time.Second), Value: 2},
			{Time: *logsAt(-20 * time.Second), Value: 2},
		}, result.Points)
		require.Equal(t, 20*time.Second, result.Read.Step)

		params.Contains, params.Step = "timeout", 0
		params.Start = *logsAt(-time.Hour)
		result, countErr = querier.CountLogs(t.Context(), params)
		require.NoError(t, countErr)
		require.Len(t, result.Points, 300)
		require.Zero(t, result.Points[0].Value)

		params.Step = time.Second
		result, countErr = querier.CountLogs(t.Context(), params)
		require.NoError(t, countErr)
		require.Equal(t, []string{"step raised from 1s to 12s to stay within 300 points"}, result.Read.Limits)
	})
}

func TestInstallationWithoutLogs(t *testing.T) {
	installed, installErr := newLogsInstallation(t, nil)
	require.NoError(t, installErr)
	_, isQuerier := installed.(integrations.LogQuerier)
	require.False(t, isQuerier, "a scenario without logs has no log source")
	require.Equal(t, []string{"deployments"}, installed.Capabilities())
}
