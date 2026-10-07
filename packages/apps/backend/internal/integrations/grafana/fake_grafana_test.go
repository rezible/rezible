package grafana

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/test"
)

const testToken = "glsa_test_token"

var (
	testRanAt = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	testStart = testRanAt.Add(-time.Hour)
	testEnd   = testRanAt
)

var testSettings = map[string]any{
	"logs_data_source_uid":    "loki",
	"metrics_data_source_uid": "prometheus",
}

type fakeResponse struct {
	status int
	body   string
}

// fakeGrafana answers data source proxy paths with canned responses and records each request.
type fakeGrafana struct {
	t         *testing.T
	server    *httptest.Server
	responses map[string]fakeResponse
	requests  []*http.Request
}

func newFakeGrafana(t *testing.T) *fakeGrafana {
	fake := &fakeGrafana{t: t, responses: map[string]fakeResponse{}}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeGrafana) serve(w http.ResponseWriter, r *http.Request) {
	f.requests = append(f.requests, r)
	response, found := f.responses[r.URL.Path]
	if !found {
		f.t.Errorf("unexpected request to %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(response.status)
	_, _ = w.Write([]byte(response.body))
}

// respond sets the response for a path under the data source proxy.
func (f *fakeGrafana) respond(uid, path string, status int, body string) {
	f.responses["/api/datasources/proxy/uid/"+uid+path] = fakeResponse{status: status, body: body}
}

// onlyRequest returns the single request the fake received.
func (f *fakeGrafana) onlyRequest() *http.Request {
	require.Len(f.t, f.requests, 1)
	return f.requests[0]
}

// install returns an installation of Grafana pointed at the fake, with the given user settings.
func (f *fakeGrafana) install(settings map[string]any) *InstalledIntegration {
	config, encodeErr := json.Marshal(InstallationConfig{URL: f.server.URL + "/", Token: testToken})
	require.NoError(f.t, encodeErr)
	integration, makeErr := MakeIntegration(test.NewClock(testRanAt))
	require.NoError(f.t, makeErr)

	intg := &ent.Integration{
		InstallationConfig: config,
		UserSettings:       settings,
	}
	installed, installErr := integration.GetInstalledIntegration(intg)
	require.NoError(f.t, installErr)
	return installed.(*InstalledIntegration)
}

type testStream struct {
	labels map[string]string
	lines  map[time.Time]string
}

func lokiStreamsBody(t *testing.T, streams ...testStream) string {
	result := make([]map[string]any, len(streams))
	for i, stream := range streams {
		values := make([][]string, 0, len(stream.lines))
		for at, line := range stream.lines {
			values = append(values, []string{strconv.FormatInt(at.UnixNano(), 10), line})
		}
		result[i] = map[string]any{"stream": stream.labels, "values": values}
	}
	return successBody(t, map[string]any{"resultType": "streams", "result": result})
}

type testSeries struct {
	labels map[string]string
	points map[time.Time]float64
}

func matrixBody(t *testing.T, series ...testSeries) string {
	result := make([]map[string]any, len(series))
	for i, s := range series {
		values := make([][]any, 0, len(s.points))
		for at, value := range s.points {
			values = append(values, []any{float64(at.UnixMilli()) / 1000, strconv.FormatFloat(value, 'f', -1, 64)})
		}
		result[i] = map[string]any{"metric": s.labels, "values": values}
	}
	return successBody(t, map[string]any{"resultType": "matrix", "result": result})
}

func labelValuesBody(t *testing.T, values ...string) string {
	return successBody(t, append([]string{}, values...))
}

func successBody(t *testing.T, data any) string {
	body, encodeErr := json.Marshal(map[string]any{"status": "success", "data": data})
	require.NoError(t, encodeErr)
	return string(body)
}

// decodeExploreURL returns the Explore URL's base and its single pane, decoded without the production types so
// the assertions check the keys Grafana reads.
func decodeExploreURL(t *testing.T, raw string) (string, map[string]any) {
	parsed, parseErr := url.Parse(raw)
	require.NoError(t, parseErr)
	require.Equal(t, "1", parsed.Query().Get("schemaVersion"))

	var panes map[string]map[string]any
	require.NoError(t, json.Unmarshal([]byte(parsed.Query().Get("panes")), &panes))
	require.Len(t, panes, 1)

	base := parsed.Scheme + "://" + parsed.Host + parsed.Path
	for _, pane := range panes {
		return base, pane
	}
	return base, nil
}
