package grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rezible/rezible/pkg/integrations"
)

var (
	errNoLogsDataSource    = errors.New("no logs data source configured")
	errNoMetricsDataSource = errors.New("no metrics data source configured")
)

var (
	_ integrations.LogQuerier    = (*InstalledIntegration)(nil)
	_ integrations.MetricQuerier = (*InstalledIntegration)(nil)
)

func (ii *InstalledIntegration) logsDataSource() (dataSource, error) {
	if ii.settings.LogsDataSourceUID == "" {
		return dataSource{}, errNoLogsDataSource
	}
	return dataSource{kind: "loki", uid: ii.settings.LogsDataSourceUID}, nil
}

func (ii *InstalledIntegration) metricsDataSource() (dataSource, error) {
	if ii.settings.MetricsDataSourceUID == "" {
		return dataSource{}, errNoMetricsDataSource
	}
	return dataSource{kind: "prometheus", uid: ii.settings.MetricsDataSourceUID}, nil
}

func (ii *InstalledIntegration) SearchLogs(ctx context.Context, params integrations.LogSearchParams) (*integrations.LogSearchResult, error) {
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}
	ds, dsErr := ii.logsDataSource()
	if dsErr != nil {
		return nil, dsErr
	}
	limit := params.EffectiveLimit()
	query := ii.settings.logSelector(params.Service, params.Contains)
	values := url.Values{
		"query":     {query},
		"start":     {formatTime(params.Start)},
		"end":       {formatTime(params.End)},
		"limit":     {strconv.Itoa(limit)},
		"direction": {"backward"},
	}
	read := integrations.ReadRecord{
		DataSourceUID: ds.uid,
		Query:         query,
		Start:         params.Start,
		End:           params.End,
		RanAt:         ii.clock.Now(),
		ExploreURL:    ii.client.exploreURL(ds, query, params.Start, params.End),
	}

	var data queryData
	if getErr := ii.client.get(ctx, ds, "/loki/api/v1/query_range", values, &data); getErr != nil {
		return nil, fmt.Errorf("search logs: %w", getErr)
	}
	var streams []lokiStream
	if decodeErr := data.decode("streams", &streams); decodeErr != nil {
		return nil, fmt.Errorf("search logs: %w", decodeErr)
	}

	lines, limits := mergeLogStreams(streams, limit)
	read.Limits = limits
	return &integrations.LogSearchResult{Lines: lines, Read: read}, nil
}

// mergeLogStreams merges lines across streams, newest first, keeping at most limit lines.
func mergeLogStreams(streams []lokiStream, limit int) ([]integrations.LogLine, []string) {
	lines := make([]integrations.LogLine, 0)
	for _, stream := range streams {
		for _, entry := range stream.Values {
			line := integrations.LogLine{
				Time:   entry.time,
				Line:   entry.line,
				Labels: stream.Stream,
			}
			lines = append(lines, line)
		}
	}
	sort.SliceStable(lines, func(a, b int) bool {
		return lines[a].Time.After(lines[b].Time)
	})

	var limits []string
	if len(lines) >= limit {
		lines = lines[:limit]
		limits = append(limits, fmt.Sprintf("reached the limit of %d lines; more may exist", limit))
	}
	truncated := 0
	for i := range lines {
		runes := []rune(lines[i].Line)
		if len(runes) > integrations.MaxLogLineLength {
			lines[i].Line = string(runes[:integrations.MaxLogLineLength])
			truncated++
		}
	}
	if truncated > 0 {
		limits = append(limits, fmt.Sprintf("truncated %d lines longer than %d characters", truncated, integrations.MaxLogLineLength))
	}
	return lines, limits
}

func (ii *InstalledIntegration) CountLogs(ctx context.Context, params integrations.LogCountParams) (*integrations.LogCountResult, error) {
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}
	ds, dsErr := ii.logsDataSource()
	if dsErr != nil {
		return nil, dsErr
	}
	step, stepLimit := integrations.ReadStep(params.Start, params.End, params.Step)
	selector := ii.settings.logSelector(params.Service, params.Contains)
	query := fmt.Sprintf("sum(count_over_time(%s [%s]))", selector, formatDuration(step))
	values := url.Values{
		"query": {query},
		"start": {formatTime(params.Start)},
		"end":   {formatTime(params.End)},
		"step":  {formatStep(step)},
	}
	read := integrations.ReadRecord{
		DataSourceUID: ds.uid,
		Query:         query,
		Start:         params.Start,
		End:           params.End,
		Step:          step,
		RanAt:         ii.clock.Now(),
		ExploreURL:    ii.client.exploreURL(ds, query, params.Start, params.End),
	}
	if stepLimit != "" {
		read.Limits = append(read.Limits, stepLimit)
	}

	var data queryData
	if getErr := ii.client.get(ctx, ds, "/loki/api/v1/query_range", values, &data); getErr != nil {
		return nil, fmt.Errorf("count logs: %w", getErr)
	}
	var series []matrixSeries
	if decodeErr := data.decode("matrix", &series); decodeErr != nil {
		return nil, fmt.Errorf("count logs: %w", decodeErr)
	}

	// The sum is a single series, absent when no line matched.
	points := make([]integrations.MetricPoint, 0)
	if len(series) > 0 {
		points = series[0].points()
	}
	return &integrations.LogCountResult{Points: points, Read: read}, nil
}

func (ii *InstalledIntegration) ListMetricNames(ctx context.Context, params integrations.MetricNamesParams) (*integrations.MetricNamesResult, error) {
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}
	ds, dsErr := ii.metricsDataSource()
	if dsErr != nil {
		return nil, dsErr
	}
	selector := ii.settings.metricSelector(params.Service)
	values := url.Values{
		"match[]": {selector},
		"start":   {formatTime(params.Start)},
		"end":     {formatTime(params.End)},
	}
	read := integrations.ReadRecord{
		DataSourceUID: ds.uid,
		Query:         selector,
		Start:         params.Start,
		End:           params.End,
		RanAt:         ii.clock.Now(),
	}

	var names []string
	if getErr := ii.client.get(ctx, ds, "/api/v1/label/__name__/values", values, &names); getErr != nil {
		return nil, fmt.Errorf("list metric names: %w", getErr)
	}

	slices.Sort(names)
	if len(names) > integrations.MaxMetricNames {
		read.Limits = append(read.Limits, fmt.Sprintf("kept the first %d of %d metric names, sorted by name", integrations.MaxMetricNames, len(names)))
		names = names[:integrations.MaxMetricNames]
	}
	return &integrations.MetricNamesResult{Names: names, Read: read}, nil
}

func (ii *InstalledIntegration) QueryMetricRange(ctx context.Context, params integrations.MetricRangeParams) (*integrations.MetricRangeResult, error) {
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}
	ds, dsErr := ii.metricsDataSource()
	if dsErr != nil {
		return nil, dsErr
	}
	step, stepLimit := integrations.ReadStep(params.Start, params.End, params.Step)
	values := url.Values{
		"query": {params.Query},
		"start": {formatTime(params.Start)},
		"end":   {formatTime(params.End)},
		"step":  {formatStep(step)},
	}
	read := integrations.ReadRecord{
		DataSourceUID: ds.uid,
		Query:         params.Query,
		Start:         params.Start,
		End:           params.End,
		Step:          step,
		RanAt:         ii.clock.Now(),
		ExploreURL:    ii.client.exploreURL(ds, params.Query, params.Start, params.End),
	}
	if stepLimit != "" {
		read.Limits = append(read.Limits, stepLimit)
	}

	var data queryData
	if getErr := ii.client.get(ctx, ds, "/api/v1/query_range", values, &data); getErr != nil {
		return nil, fmt.Errorf("query metric range: %w", getErr)
	}
	var matrix []matrixSeries
	if decodeErr := data.decode("matrix", &matrix); decodeErr != nil {
		return nil, fmt.Errorf("query metric range: %w", decodeErr)
	}

	slices.SortFunc(matrix, func(a, b matrixSeries) int {
		return strings.Compare(a.labelSet(), b.labelSet())
	})
	if len(matrix) > integrations.MaxMetricSeries {
		read.Limits = append(read.Limits, fmt.Sprintf("kept the first %d of %d series, ordered by labels", integrations.MaxMetricSeries, len(matrix)))
		matrix = matrix[:integrations.MaxMetricSeries]
	}
	series := make([]integrations.MetricSeries, len(matrix))
	for i, s := range matrix {
		series[i] = integrations.MetricSeries{Labels: s.Metric, Points: s.points()}
	}
	return &integrations.MetricRangeResult{Series: series, Read: read}, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// formatStep formats a whole-second step as the seconds both APIs accept.
func formatStep(step time.Duration) string {
	return strconv.FormatInt(int64(step/time.Second), 10)
}

// formatDuration formats a whole-second step as a LogQL range duration.
func formatDuration(step time.Duration) string {
	return formatStep(step) + "s"
}

// queryData is the data of a Loki or Prometheus query response.
type queryData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

func (d queryData) decode(resultType string, result any) error {
	if d.ResultType != resultType {
		return fmt.Errorf("expected a %s result, got %q", resultType, d.ResultType)
	}
	if decodeErr := json.Unmarshal(d.Result, result); decodeErr != nil {
		return fmt.Errorf("decode %s result: %w", resultType, decodeErr)
	}
	return nil
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values []lokiEntry       `json:"values"`
}

// lokiEntry is a `[nanoseconds, line]` pair. Later elements, such as structured metadata, are ignored.
type lokiEntry struct {
	time time.Time
	line string
}

func (e *lokiEntry) UnmarshalJSON(raw []byte) error {
	var parts []json.RawMessage
	if decodeErr := json.Unmarshal(raw, &parts); decodeErr != nil {
		return decodeErr
	}
	if len(parts) < 2 {
		return fmt.Errorf("log entry has %d elements", len(parts))
	}
	var timestamp string
	if decodeErr := json.Unmarshal(parts[0], &timestamp); decodeErr != nil {
		return fmt.Errorf("log entry time: %w", decodeErr)
	}
	nanos, parseErr := strconv.ParseInt(timestamp, 10, 64)
	if parseErr != nil {
		return fmt.Errorf("log entry time: %w", parseErr)
	}
	if decodeErr := json.Unmarshal(parts[1], &e.line); decodeErr != nil {
		return fmt.Errorf("log entry line: %w", decodeErr)
	}
	e.time = time.Unix(0, nanos).UTC()
	return nil
}

type matrixSeries struct {
	Metric map[string]string `json:"metric"`
	Values []matrixSample    `json:"values"`
}

// labelSet is the series' labels in a stable order, for sorting.
func (s matrixSeries) labelSet() string {
	names := make([]string, 0, len(s.Metric))
	for name := range s.Metric {
		names = append(names, name)
	}
	slices.Sort(names)
	pairs := make([]string, len(names))
	for i, name := range names {
		pairs[i] = name + "=" + strconv.Quote(s.Metric[name])
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (s matrixSeries) points() []integrations.MetricPoint {
	points := make([]integrations.MetricPoint, len(s.Values))
	for i, sample := range s.Values {
		points[i] = integrations.MetricPoint(sample)
	}
	return points
}

// matrixSample is a `[unix seconds, "value"]` pair.
type matrixSample integrations.MetricPoint

func (p *matrixSample) UnmarshalJSON(raw []byte) error {
	var parts []json.RawMessage
	if decodeErr := json.Unmarshal(raw, &parts); decodeErr != nil {
		return decodeErr
	}
	if len(parts) != 2 {
		return fmt.Errorf("sample has %d elements", len(parts))
	}
	var seconds float64
	if decodeErr := json.Unmarshal(parts[0], &seconds); decodeErr != nil {
		return fmt.Errorf("sample time: %w", decodeErr)
	}
	var value string
	if decodeErr := json.Unmarshal(parts[1], &value); decodeErr != nil {
		return fmt.Errorf("sample value: %w", decodeErr)
	}
	parsed, parseErr := strconv.ParseFloat(value, 64)
	if parseErr != nil {
		return fmt.Errorf("sample value: %w", parseErr)
	}
	p.Time = time.UnixMilli(int64(math.Round(seconds * 1000))).UTC()
	p.Value = parsed
	return nil
}
