package grafana

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGrafanaHealthCheck(t *testing.T) {
	t.Run("both data sources have service label values", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/label/service_name/values", http.StatusOK, labelValuesBody(t, "checkout-api"))
		fake.respond("prometheus", "/api/v1/label/service_name/values", http.StatusOK, labelValuesBody(t, "checkout-api"))
		installed := fake.install(testSettings)

		checkErr := installed.CheckHealth(t.Context())

		require.NoError(t, checkErr)
		require.Len(t, fake.requests, 2)
		for _, request := range fake.requests {
			require.Equal(t, testRanAt.Add(-15*time.Minute).Format(time.RFC3339Nano), request.URL.Query().Get("start"))
			require.Equal(t, testRanAt.Format(time.RFC3339Nano), request.URL.Query().Get("end"))
		}
	})

	t.Run("empty data source UIDs are both reported", func(t *testing.T) {
		fake := newFakeGrafana(t)
		installed := fake.install(map[string]any{})

		checkErr := installed.CheckHealth(t.Context())

		require.EqualError(t, checkErr, "logs: no logs data source configured\nmetrics: no metrics data source configured")
		require.Empty(t, fake.requests)
	})

	t.Run("a label with no values names its data source", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/label/app/values", http.StatusOK, labelValuesBody(t))
		fake.respond("prometheus", "/api/v1/label/service_name/values", http.StatusOK, labelValuesBody(t, "checkout-api"))
		settings := map[string]any{
			"logs_data_source_uid":    "loki",
			"metrics_data_source_uid": "prometheus",
			"log_service_label":       "app",
		}
		installed := fake.install(settings)

		checkErr := installed.CheckHealth(t.Context())

		require.EqualError(t, checkErr, "logs: no values for label `app` in the last 15 minutes")
	})

	t.Run("a provider error and a missing data source are reported together", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", "/loki/api/v1/label/service_name/values", http.StatusUnauthorized, `{"message":"Invalid API key"}`)
		settings := map[string]any{
			"logs_data_source_uid": "loki",
		}
		installed := fake.install(settings)

		checkErr := installed.CheckHealth(t.Context())

		require.EqualError(t, checkErr, "logs: grafana responded with status 401: Invalid API key\nmetrics: no metrics data source configured")
	})
}
