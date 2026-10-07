package grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rezible/rezible/pkg/integrations"
)

const maxProviderErrorLength = 300 // runes

// dataSource is one Loki or Prometheus data source reached through Grafana's data source proxy.
type dataSource struct {
	kind string // the Grafana data source type, "loki" or "prometheus"
	uid  string
}

// client reads Loki and Prometheus through Grafana's data source proxy.
type client struct {
	http    *http.Client
	baseURL string
	token   string
}

// apiResponse is the envelope shared by Loki and Prometheus responses, and the error bodies of Grafana, Loki
// and Prometheus.
type apiResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`   // Prometheus
	Message string          `json:"message"` // Grafana and Loki
}

// get reads a data source API path and decodes the response's data into result.
func (c *client) get(ctx context.Context, ds dataSource, path string, query url.Values, result any) error {
	ctx, cancel := context.WithTimeout(ctx, integrations.TelemetryReadTimeout)
	defer cancel()

	endpoint := c.baseURL + "/api/datasources/proxy/uid/" + url.PathEscape(ds.uid) + path + "?" + query.Encode()
	request, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if requestErr != nil {
		return fmt.Errorf("build request: %w", requestErr)
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/json")

	response, doErr := c.http.Do(request)
	if doErr != nil {
		if errors.Is(doErr, context.DeadlineExceeded) {
			return fmt.Errorf("grafana did not respond within %s", integrations.TelemetryReadTimeout)
		}
		return fmt.Errorf("request grafana: %w", doErr)
	}
	defer response.Body.Close()

	rawBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		if errors.Is(readErr, context.DeadlineExceeded) {
			return fmt.Errorf("grafana did not respond within %s", integrations.TelemetryReadTimeout)
		}
		return fmt.Errorf("read grafana response: %w", readErr)
	}
	var body apiResponse
	decodeErr := json.Unmarshal(rawBody, &body)
	failed := response.StatusCode < 200 || response.StatusCode >= 300 || body.Status != "success"
	if decodeErr != nil || failed {
		return c.responseError(response.StatusCode, body, decodeErr)
	}
	if dataErr := json.Unmarshal(body.Data, result); dataErr != nil {
		return fmt.Errorf("decode grafana response: %w", dataErr)
	}
	return nil
}

// responseError carries the provider's message only when the body parses as a Grafana, Loki or Prometheus
// error, so an arbitrary response from the configured URL is never shown. The token is redacted before the
// message is bounded, in case the provider echoes it.
func (c *client) responseError(status int, body apiResponse, decodeErr error) error {
	message := body.Error
	if message == "" {
		message = body.Message
	}
	if decodeErr != nil || message == "" {
		return fmt.Errorf("grafana responded with status %d", status)
	}
	message = strings.ReplaceAll(message, c.token, "[redacted]")
	runes := []rune(message)
	if len(runes) > maxProviderErrorLength {
		message = string(runes[:maxProviderErrorLength]) + "…"
	}
	return fmt.Errorf("grafana responded with status %d: %s", status, message)
}

type explorePane struct {
	Datasource string         `json:"datasource"`
	Queries    []exploreQuery `json:"queries"`
	Range      exploreRange   `json:"range"`
}

type exploreQuery struct {
	RefID      string            `json:"refId"`
	Datasource exploreDatasource `json:"datasource"`
	Expr       string            `json:"expr"`
	EditorMode string            `json:"editorMode"`
}

type exploreDatasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

type exploreRange struct {
	From string `json:"from"` // unix milliseconds
	To   string `json:"to"`
}

// exploreURL opens the query and range in Grafana Explore, using the `panes` URL format.
func (c *client) exploreURL(ds dataSource, query string, start, end time.Time) string {
	pane := explorePane{
		Datasource: ds.uid,
		Queries: []exploreQuery{{
			RefID:      "A",
			Datasource: exploreDatasource{Type: ds.kind, UID: ds.uid},
			Expr:       query,
			EditorMode: "code",
		}},
		Range: exploreRange{
			From: strconv.FormatInt(start.UnixMilli(), 10),
			To:   strconv.FormatInt(end.UnixMilli(), 10),
		},
	}
	panes, encodeErr := json.Marshal(map[string]explorePane{"a": pane})
	if encodeErr != nil {
		return ""
	}
	params := url.Values{
		"schemaVersion": {"1"},
		"panes":         {string(panes)},
	}
	return c.baseURL + "/explore?" + params.Encode()
}
