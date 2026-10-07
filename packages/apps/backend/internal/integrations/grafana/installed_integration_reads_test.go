package grafana

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/integrations"
)

func TestLogQueriesAreBoundToTheService(t *testing.T) {
	t.Run("search names the service label and escapes the text", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, lokiStreamsBody(t))
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{
			Service:  "checkout-api",
			Start:    testStart,
			End:      testEnd,
			Contains: `say "hi" \ there`,
		}
		_, searchErr := installed.SearchLogs(t.Context(), params)

		require.NoError(t, searchErr)
		request := fake.onlyRequest()
		require.Equal(t, `{service_name="checkout-api"} |= "say \"hi\" \\ there"`, request.URL.Query().Get("query"))
		require.Equal(t, "backward", request.URL.Query().Get("direction"))
		require.Equal(t, "Bearer "+testToken, request.Header.Get("Authorization"))
	})

	t.Run("a service value cannot widen the selector", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, matrixBody(t))
		installed := fake.install(testSettings)

		params := integrations.LogCountParams{
			Service: `checkout-api", service_name=~".+`,
			Start:   testStart,
			End:     testEnd,
		}
		_, countErr := installed.CountLogs(t.Context(), params)

		require.NoError(t, countErr)
		query := fake.onlyRequest().URL.Query().Get("query")
		require.Equal(t, `sum(count_over_time({service_name="checkout-api\", service_name=~\".+"} [12s]))`, query)
	})

	t.Run("the configured label names the service", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("prometheus", "/api/v1/label/__name__/values", http.StatusOK, labelValuesBody(t))
		settings := map[string]any{
			"metrics_data_source_uid": "prometheus",
			"metric_service_label":    "app",
		}
		installed := fake.install(settings)

		params := integrations.MetricNamesParams{
			Service: "checkout-api",
			Start:   testStart,
			End:     testEnd,
		}
		_, listErr := installed.ListMetricNames(t.Context(), params)

		require.NoError(t, listErr)
		require.Equal(t, []string{`{app="checkout-api"}`}, fake.onlyRequest().URL.Query()["match[]"])
	})

	t.Run("an empty service is an error", func(t *testing.T) {
		fake := newFakeGrafana(t)
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{
			Service: " ",
			Start:   testStart,
			End:     testEnd,
		}
		_, searchErr := installed.SearchLogs(t.Context(), params)

		require.ErrorIs(t, searchErr, rez.ErrInvalidInput)
		require.ErrorContains(t, searchErr, "service is required")
		require.Empty(t, fake.requests)
	})
}

func TestReadLimits(t *testing.T) {
	t.Run("requests outside the limits are errors naming the limit", func(t *testing.T) {
		fake := newFakeGrafana(t)
		installed := fake.install(testSettings)

		tooLong := integrations.LogSearchParams{Service: "checkout-api", Start: testEnd.Add(-7 * time.Hour), End: testEnd}
		_, tooLongErr := installed.SearchLogs(t.Context(), tooLong)
		require.ErrorIs(t, tooLongErr, rez.ErrInvalidInput)
		require.ErrorContains(t, tooLongErr, "longer than the 6h0m0s limit")

		backwards := integrations.MetricRangeParams{Query: "up", Start: testEnd, End: testEnd}
		_, backwardsErr := installed.QueryMetricRange(t.Context(), backwards)
		require.ErrorIs(t, backwardsErr, rez.ErrInvalidInput)
		require.ErrorContains(t, backwardsErr, "end must be after start")

		tooMany := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd, Limit: 201}
		_, tooManyErr := installed.SearchLogs(t.Context(), tooMany)
		require.ErrorIs(t, tooManyErr, rez.ErrInvalidInput)
		require.ErrorContains(t, tooManyErr, "maximum of 200 lines")

		require.Empty(t, fake.requests)
	})

	t.Run("a search reaching its limit says so and keeps the newest lines", func(t *testing.T) {
		fake := newFakeGrafana(t)
		checkout := testStream{
			labels: map[string]string{"service_name": "checkout-api", "level": "error"},
			lines: map[time.Time]string{
				testEnd.Add(-3 * time.Minute): "oldest",
				testEnd.Add(-1 * time.Minute): "newest",
			},
		}
		checkoutInfo := testStream{
			labels: map[string]string{"service_name": "checkout-api", "level": "info"},
			lines: map[time.Time]string{
				testEnd.Add(-2 * time.Minute): "middle",
			},
		}
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, lokiStreamsBody(t, checkout, checkoutInfo))
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd, Limit: 2}
		result, searchErr := installed.SearchLogs(t.Context(), params)

		require.NoError(t, searchErr)
		require.Equal(t, "2", fake.onlyRequest().URL.Query().Get("limit"))
		require.Len(t, result.Lines, 2)
		require.Equal(t, "newest", result.Lines[0].Line)
		require.Equal(t, "middle", result.Lines[1].Line)
		require.Equal(t, "info", result.Lines[1].Labels["level"])
		require.Equal(t, []string{"reached the limit of 2 lines; more may exist"}, result.Read.Limits)
	})

	t.Run("long lines are truncated by rune", func(t *testing.T) {
		fake := newFakeGrafana(t)
		stream := testStream{
			labels: map[string]string{"service_name": "checkout-api"},
			lines: map[time.Time]string{
				testEnd.Add(-time.Minute): strings.Repeat("é", 2500),
			},
		}
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, lokiStreamsBody(t, stream))
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd}
		result, searchErr := installed.SearchLogs(t.Context(), params)

		require.NoError(t, searchErr)
		require.Equal(t, strings.Repeat("é", 2000), result.Lines[0].Line)
		require.Equal(t, []string{"truncated 1 lines longer than 2000 characters"}, result.Read.Limits)
	})

	t.Run("steps are raised to fit 300 points", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, matrixBody(t))
		fake.respond("prometheus", "/api/v1/query_range", http.StatusOK, matrixBody(t))
		installed := fake.install(testSettings)
		sixHours := testEnd.Add(-6 * time.Hour)

		countParams := integrations.LogCountParams{Service: "checkout-api", Start: sixHours, End: testEnd, Step: time.Second}
		count, countErr := installed.CountLogs(t.Context(), countParams)
		require.NoError(t, countErr)

		rangeParams := integrations.MetricRangeParams{Query: "up", Start: sixHours, End: testEnd, Step: 10 * time.Second}
		metrics, rangeErr := installed.QueryMetricRange(t.Context(), rangeParams)
		require.NoError(t, rangeErr)

		require.Len(t, fake.requests, 2)
		require.Equal(t, "72", fake.requests[0].URL.Query().Get("step"))
		require.Equal(t, 72*time.Second, count.Read.Step)
		require.Equal(t, []string{"step raised from 1s to 1m12s to stay within 300 points"}, count.Read.Limits)
		require.Equal(t, "72", fake.requests[1].URL.Query().Get("step"))
		require.Equal(t, []string{"step raised from 10s to 1m12s to stay within 300 points"}, metrics.Read.Limits)
	})

	t.Run("series beyond 20 are dropped by label set", func(t *testing.T) {
		fake := newFakeGrafana(t)
		series := make([]testSeries, 0, 25)
		for i := 24; i >= 0; i-- {
			s := testSeries{
				labels: map[string]string{"__name__": "up", "instance": fmt.Sprintf("host-%02d", i)},
				points: map[time.Time]float64{testEnd: 1},
			}
			series = append(series, s)
		}
		fake.respond("prometheus", "/api/v1/query_range", http.StatusOK, matrixBody(t, series...))
		installed := fake.install(testSettings)

		params := integrations.MetricRangeParams{Query: "up", Start: testStart, End: testEnd}
		result, rangeErr := installed.QueryMetricRange(t.Context(), params)

		require.NoError(t, rangeErr)
		require.Len(t, result.Series, 20)
		require.Equal(t, "host-00", result.Series[0].Labels["instance"])
		require.Equal(t, "host-19", result.Series[19].Labels["instance"])
		require.Equal(t, []integrations.MetricPoint{{Time: testEnd, Value: 1}}, result.Series[0].Points)
		require.Equal(t, []string{"kept the first 20 of 25 series, ordered by labels"}, result.Read.Limits)
	})

	t.Run("metric names beyond 200 are dropped in sorted order", func(t *testing.T) {
		fake := newFakeGrafana(t)
		names := make([]string, 0, 250)
		for i := 249; i >= 0; i-- {
			names = append(names, "metric_"+strconv.Itoa(1000+i))
		}
		fake.respond("prometheus", "/api/v1/label/__name__/values", http.StatusOK, labelValuesBody(t, names...))
		installed := fake.install(testSettings)

		params := integrations.MetricNamesParams{Service: "checkout-api", Start: testStart, End: testEnd}
		result, listErr := installed.ListMetricNames(t.Context(), params)

		require.NoError(t, listErr)
		require.Len(t, result.Names, 200)
		require.Equal(t, "metric_1000", result.Names[0])
		require.Equal(t, "metric_1199", result.Names[199])
		require.Equal(t, []string{"kept the first 200 of 250 metric names, sorted by name"}, result.Read.Limits)
	})
}

func TestReadRecordDescribesTheQuery(t *testing.T) {
	t.Run("log count", func(t *testing.T) {
		fake := newFakeGrafana(t)
		counts := testSeries{
			labels: map[string]string{},
			points: map[time.Time]float64{testEnd: 7},
		}
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusOK, matrixBody(t, counts))
		installed := fake.install(testSettings)

		params := integrations.LogCountParams{Service: "checkout-api", Start: testStart, End: testEnd, Contains: "upstream timeout"}
		result, countErr := installed.CountLogs(t.Context(), params)

		require.NoError(t, countErr)
		sent := fake.onlyRequest().URL.Query()
		require.Equal(t, `sum(count_over_time({service_name="checkout-api"} |= "upstream timeout" [12s]))`, result.Read.Query)
		require.Equal(t, sent.Get("query"), result.Read.Query)
		require.Equal(t, testStart.Format(time.RFC3339Nano), sent.Get("start"))
		require.Equal(t, testEnd.Format(time.RFC3339Nano), sent.Get("end"))
		require.Equal(t, "loki", result.Read.DataSourceUID)
		require.Equal(t, testStart, result.Read.Start)
		require.Equal(t, testEnd, result.Read.End)
		require.Equal(t, 12*time.Second, result.Read.Step)
		require.Equal(t, testRanAt, result.Read.RanAt)
		require.Empty(t, result.Read.Limits)
		require.Equal(t, []integrations.MetricPoint{{Time: testEnd, Value: 7}}, result.Points)

		base, pane := decodeExploreURL(t, result.Read.ExploreURL)
		require.Equal(t, fake.server.URL+"/explore", base)
		expectedPane := map[string]any{
			"datasource": "loki",
			"queries": []any{
				map[string]any{
					"refId":      "A",
					"datasource": map[string]any{"type": "loki", "uid": "loki"},
					"expr":       result.Read.Query,
					"editorMode": "code",
				},
			},
			"range": map[string]any{
				"from": strconv.FormatInt(testStart.UnixMilli(), 10),
				"to":   strconv.FormatInt(testEnd.UnixMilli(), 10),
			},
		}
		require.Equal(t, expectedPane, pane)
	})

	t.Run("metric range", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("prometheus", "/api/v1/query_range", http.StatusOK, matrixBody(t))
		installed := fake.install(testSettings)
		query := `sum(rate(http_server_request_duration_seconds_count{service_name="checkout-api"}[5m]))`

		params := integrations.MetricRangeParams{Query: query, Start: testStart, End: testEnd, Step: time.Minute}
		result, rangeErr := installed.QueryMetricRange(t.Context(), params)

		require.NoError(t, rangeErr)
		require.Equal(t, query, fake.onlyRequest().URL.Query().Get("query"))
		require.Equal(t, query, result.Read.Query)
		require.Equal(t, time.Minute, result.Read.Step)
		require.Empty(t, result.Read.Limits)

		_, pane := decodeExploreURL(t, result.Read.ExploreURL)
		expectedPane := map[string]any{
			"datasource": "prometheus",
			"queries": []any{
				map[string]any{
					"refId":      "A",
					"datasource": map[string]any{"type": "prometheus", "uid": "prometheus"},
					"expr":       query,
					"editorMode": "code",
				},
			},
			"range": map[string]any{
				"from": strconv.FormatInt(testStart.UnixMilli(), 10),
				"to":   strconv.FormatInt(testEnd.UnixMilli(), 10),
			},
		}
		require.Equal(t, expectedPane, pane)
	})

	t.Run("metric names have no step or link", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("prometheus", "/api/v1/label/__name__/values", http.StatusOK, labelValuesBody(t, "up"))
		installed := fake.install(testSettings)

		params := integrations.MetricNamesParams{Service: "checkout-api", Start: testStart, End: testEnd}
		result, listErr := installed.ListMetricNames(t.Context(), params)

		require.NoError(t, listErr)
		require.Equal(t, `{service_name="checkout-api"}`, result.Read.Query)
		require.Zero(t, result.Read.Step)
		require.Empty(t, result.Read.ExploreURL)
	})
}

func TestGrafanaErrorsAreReported(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		body     string
		contains string
	}{
		{
			name:     "grafana rejects the token",
			status:   http.StatusUnauthorized,
			body:     `{"message":"Invalid API key","traceID":""}`,
			contains: "grafana responded with status 401: Invalid API key",
		},
		{
			name:     "loki rejects the query",
			status:   http.StatusBadRequest,
			body:     `{"code":400,"status":"error","message":"parse error at line 1, col 2: syntax error"}`,
			contains: "grafana responded with status 400: parse error at line 1, col 2: syntax error",
		},
		{
			name:     "prometheus reports an error with a success status",
			status:   http.StatusOK,
			body:     `{"status":"error","errorType":"execution","error":"query timed out"}`,
			contains: "grafana responded with status 200: query timed out",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeGrafana(t)
			fake.respond("loki", "/loki/api/v1/query_range", tc.status, tc.body)
			installed := fake.install(testSettings)

			params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd}
			_, searchErr := installed.SearchLogs(t.Context(), params)

			require.ErrorContains(t, searchErr, tc.contains)
		})
	}

	t.Run("a body that is not a provider error carries only its status", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusBadGateway, "<html>internal proxy page</html>")
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd}
		_, searchErr := installed.SearchLogs(t.Context(), params)

		require.EqualError(t, searchErr, "search logs: grafana responded with status 502")
	})

	t.Run("the token is redacted from provider messages", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusUnauthorized, `{"message":"invalid token `+testToken+`"}`)
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd}
		_, searchErr := installed.SearchLogs(t.Context(), params)

		require.ErrorContains(t, searchErr, "invalid token [redacted]")
		require.NotContains(t, searchErr.Error(), testToken)
	})

	t.Run("long messages are bounded", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/query_range", http.StatusBadRequest, `{"message":"`+strings.Repeat("x", 1000)+`"}`)
		installed := fake.install(testSettings)

		params := integrations.LogSearchParams{Service: "checkout-api", Start: testStart, End: testEnd}
		_, searchErr := installed.SearchLogs(t.Context(), params)

		require.ErrorContains(t, searchErr, strings.Repeat("x", 300)+"…")
		require.NotContains(t, searchErr.Error(), strings.Repeat("x", 301))
	})

	t.Run("a read without a data source is an error", func(t *testing.T) {
		fake := newFakeGrafana(t)
		installed := fake.install(map[string]any{})

		params := integrations.MetricRangeParams{Query: "up", Start: testStart, End: testEnd}
		_, rangeErr := installed.QueryMetricRange(t.Context(), params)

		require.EqualError(t, rangeErr, "no metrics data source configured")
	})
}
