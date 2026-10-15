package evals

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/integrations"
)

// maxScenarioLogLines bounds the lines a scenario's logs expand to.
const maxScenarioLogLines = 10_000

// ScenarioLog is one log line of a service, or the same line repeated from From, every Every, before To.
type ScenarioLog struct {
	// Service is the resource ref of a service entity the scenario ingests.
	Service string            `json:"service"`
	Line    string            `json:"line"`
	Labels  map[string]string `json:"labels,omitempty"`
	At      *time.Time        `json:"at,omitempty"`
	From    *time.Time        `json:"from,omitempty"`
	To      *time.Time        `json:"to,omitempty"`
	Every   string            `json:"every,omitempty"`
}

func (l ScenarioLog) validate() error {
	switch {
	case strings.TrimSpace(l.Service) == "":
		return fmt.Errorf("no service")
	case strings.TrimSpace(l.Line) == "":
		return fmt.Errorf("no line")
	}
	series := l.From != nil || l.To != nil || l.Every != ""
	switch {
	case l.At != nil && series:
		return fmt.Errorf("both at and from, to and every")
	case l.At != nil:
		return nil
	case l.From == nil || l.To == nil || l.Every == "":
		return fmt.Errorf("neither at nor all of from, to and every")
	case !l.To.After(*l.From):
		return fmt.Errorf("to is not after from")
	}
	every, everyErr := time.ParseDuration(l.Every)
	if everyErr != nil {
		return fmt.Errorf("every: %w", everyErr)
	}
	if every < time.Second {
		return fmt.Errorf("every is under one second")
	}
	return nil
}

// times returns when the entry's lines are logged, oldest first. The entry must be valid.
func (l ScenarioLog) times() []time.Time {
	if l.At != nil {
		return []time.Time{*l.At}
	}
	every, _ := time.ParseDuration(l.Every)
	var times []time.Time
	for at := *l.From; at.Before(*l.To); at = at.Add(every) {
		times = append(times, at)
	}
	return times
}

// lineCount is how many lines the entry expands to, without expanding it. The entry must be valid.
func (l ScenarioLog) lineCount() int {
	if l.At != nil {
		return 1
	}
	every, _ := time.ParseDuration(l.Every)
	span := l.To.Sub(*l.From)
	return int((span + every - 1) / every)
}

func validateLogs(logs []ScenarioLog) error {
	lines := 0
	for index, entry := range logs {
		if entryErr := entry.validate(); entryErr != nil {
			return fmt.Errorf("entry %d: %w", index+1, entryErr)
		}
		lines += entry.lineCount()
	}
	if lines > maxScenarioLogLines {
		return fmt.Errorf("%d lines, more than the %d allowed", lines, maxScenarioLogLines)
	}
	return nil
}

// installedLogs is an evaluation installation whose scenario has logs: it reads them as the organization's
// log source.
type installedLogs struct {
	*installedIntegration
	clock rez.Clock
	logs  []ScenarioLog
}

var _ integrations.LogQuerier = (*installedLogs)(nil)

func (ii *installedLogs) CheckLogsAvailable() error {
	return nil
}

func (ii *installedLogs) SearchLogs(_ context.Context, params integrations.LogSearchParams) (*integrations.LogSearchResult, error) {
	service, lines := ii.matchingLines(params.Service, params.Start, params.End, params.Contains)
	slices.SortStableFunc(lines, func(a, b integrations.LogLine) int {
		return b.Time.Compare(a.Time)
	})
	read := ii.readRecord(service, params.Contains, params.Start, params.End, 0)
	if limit := params.EffectiveLimit(); len(lines) > limit {
		read.Limits = append(read.Limits, fmt.Sprintf("reached the limit of %d lines; %d more matched", limit, len(lines)-limit))
		lines = lines[:limit]
	}
	return &integrations.LogSearchResult{Service: service, Lines: lines, Read: read}, nil
}

func (ii *installedLogs) CountLogs(_ context.Context, params integrations.LogCountParams) (*integrations.LogCountResult, error) {
	service, lines := ii.matchingLines(params.Service, params.Start, params.End, params.Contains)
	step, stepLimit := integrations.ReadStep(params.Start, params.End, params.Step)
	read := ii.readRecord(service, params.Contains, params.Start, params.End, step)
	if stepLimit != "" {
		read.Limits = append(read.Limits, stepLimit)
	}
	var points []integrations.MetricPoint
	for at := params.Start; at.Before(params.End); at = at.Add(step) {
		points = append(points, integrations.MetricPoint{Time: at})
	}
	for _, line := range lines {
		points[int(line.Time.Sub(params.Start)/step)].Value++
	}
	return &integrations.LogCountResult{Service: service, Points: points, Read: read}, nil
}

// matchingLines returns the value the service is read by, and its lines in [start, end) containing the text.
// The service is read by the first of its aliases a log entry names; with none, by its display name, which no
// entry names.
func (ii *installedLogs) matchingLines(service integrations.LogService, start, end time.Time, contains string) (integrations.LogServiceName, []integrations.LogLine) {
	name := integrations.LogServiceName{Value: service.DisplayName}
	for _, alias := range service.Aliases {
		if slices.ContainsFunc(ii.logs, func(entry ScenarioLog) bool { return entry.Service == alias }) {
			name = integrations.LogServiceName{Value: alias, FromAlias: true}
			break
		}
	}
	lines := make([]integrations.LogLine, 0)
	if !name.FromAlias {
		return name, lines
	}
	for _, entry := range ii.logs {
		if entry.Service != name.Value || !strings.Contains(entry.Line, contains) {
			continue
		}
		for _, at := range entry.times() {
			if !at.Before(start) && at.Before(end) {
				lines = append(lines, integrations.LogLine{Time: at, Line: entry.Line, Labels: entry.Labels})
			}
		}
	}
	return name, lines
}

func (ii *installedLogs) readRecord(service integrations.LogServiceName, contains string, start, end time.Time, step time.Duration) integrations.ReadRecord {
	query := "service=" + strconv.Quote(service.Value)
	if contains != "" {
		query += " contains=" + strconv.Quote(contains)
	}
	return integrations.ReadRecord{
		Query: query,
		Start: start,
		End:   end,
		Step:  step,
		RanAt: ii.clock.Now(),
	}
}
