package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
)

const servicesPath = "/loki/api/v1/label/service_name/values"

// queryServices runs the installation's services querier and collects what it yields.
func queryServices(t *testing.T, installed *InstalledIntegration, cursors rez.ProviderEventSourceCursors) ([]*rez.ProviderEventQueryResult, error) {
	querier := &eventQuerier{ii: installed}
	var results []*rez.ProviderEventQueryResult
	for result, queryErr := range querier.QueryProviderEvents(t.Context(), cursors) {
		if queryErr != nil {
			return results, queryErr
		}
		results = append(results, result)
	}
	return results, nil
}

func decodeListing(t *testing.T, event rez.ProviderEvent) serviceListing {
	var listing serviceListing
	require.NoError(t, json.Unmarshal(event.Attributes, &listing))
	return listing
}

func TestServicesQuerier(t *testing.T) {
	t.Run("lists the log service label's values as one event", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", servicesPath, http.StatusOK, labelValuesBody(t, "payments", "checkout-api", "Checkout_API"))
		installed := fake.install(testSettings)

		// Cursors for other sources, or none, still list services.
		results, queryErr := queryServices(t, installed, rez.ProviderEventSourceCursors{"other": "cursor"})

		require.NoError(t, queryErr)
		require.Len(t, results, 1)
		event := results[0].Event
		require.Equal(t, ProviderName, event.Provider)
		require.Equal(t, installed.intg.ID.String(), event.ProviderNamespace)
		require.Equal(t, SourceServices, event.ProviderEventSource)
		require.Equal(t, testRanAt, event.ReceivedAt)
		require.Regexp(t, regexp.MustCompile(`^logs:service_name:[0-9a-f]{64}$`), event.ProviderEventRef)
		require.Equal(t, serviceListing{
			Signal: "logs",
			Label:  "service_name",
			Values: []string{"Checkout_API", "checkout-api", "payments"},
		}, decodeListing(t, event))
		require.NotNil(t, results[0].ProviderEventSourceCursorAfter)
		require.Equal(t, "logs", *results[0].ProviderEventSourceCursorAfter)

		request := fake.onlyRequest()
		require.Equal(t, testRanAt.Add(-24*time.Hour).Format(time.RFC3339Nano), request.URL.Query().Get("start"))
		require.Equal(t, testRanAt.Format(time.RFC3339Nano), request.URL.Query().Get("end"))
	})

	t.Run("the event ref depends only on the label and its values", func(t *testing.T) {
		refFor := func(settings map[string]any, path string, values ...string) string {
			fake := newFakeGrafana(t)
			fake.respond("loki", path, http.StatusOK, labelValuesBody(t, values...))
			results, queryErr := queryServices(t, fake.install(settings), nil)
			require.NoError(t, queryErr)
			require.Len(t, results, 1)
			return results[0].Event.ProviderEventRef
		}
		appSettings := map[string]any{"logs_data_source_uid": "loki", "log_service_label": "app"}

		first := refFor(testSettings, servicesPath, "b", "a")
		require.Equal(t, first, refFor(testSettings, servicesPath, "a", "b"), "the same values in another order")
		require.NotEqual(t, first, refFor(testSettings, servicesPath, "a", "c"), "other values")
		require.NotEqual(t, first, refFor(appSettings, "/loki/api/v1/label/app/values", "a", "b"), "another label")
	})

	t.Run("without a logs data source it lists nothing", func(t *testing.T) {
		fake := newFakeGrafana(t)
		installed := fake.install(map[string]any{"metrics_data_source_uid": "prometheus"})

		results, queryErr := queryServices(t, installed, nil)

		require.NoError(t, queryErr)
		require.Empty(t, results)
		require.Empty(t, fake.requests)
	})

	t.Run("values past the limit are dropped", func(t *testing.T) {
		values := make([]string, MaxServiceLabelValues+2)
		for i := range values {
			values[i] = fmt.Sprintf("service-%04d", len(values)-i)
		}
		fake := newFakeGrafana(t)
		fake.respond("loki", servicesPath, http.StatusOK, labelValuesBody(t, values...))
		installed := fake.install(testSettings)

		results, queryErr := queryServices(t, installed, nil)

		require.NoError(t, queryErr)
		require.Len(t, results, 1)
		listed := decodeListing(t, results[0].Event).Values
		require.Len(t, listed, MaxServiceLabelValues)
		require.Equal(t, "service-0001", listed[0])
		require.Equal(t, fmt.Sprintf("service-%04d", MaxServiceLabelValues), listed[MaxServiceLabelValues-1])
	})

	t.Run("a Grafana error is returned", func(t *testing.T) {
		fake := newFakeGrafana(t)
		fake.respond("loki", servicesPath, http.StatusUnauthorized, `{"message":"Invalid API key"}`)
		installed := fake.install(testSettings)

		results, queryErr := queryServices(t, installed, nil)

		require.ErrorContains(t, queryErr, "401")
		require.Empty(t, results)
	})
}
