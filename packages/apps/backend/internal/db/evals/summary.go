package evals

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Summary is what one target's repetitions add up to.
type Summary struct {
	Target      Target              `json:"target"`
	Runs        int                 `json:"runs"`
	Passed      int                 `json:"passed"`
	Failed      int                 `json:"failed"`
	Errors      int                 `json:"errors"`
	Checks      []CheckSummary      `json:"checks"`
	Repetitions []RepetitionSummary `json:"repetitions"`
	Turns       []TurnSummary       `json:"turns,omitempty"`
}

// CheckSummary counts a check's passes over the repetitions that recorded it. A repetition that errored
// before reaching a check does not record it.
type CheckSummary struct {
	ID       string `json:"id"`
	Passed   int    `json:"passed"`
	Recorded int    `json:"recorded"`
}

type RepetitionSummary struct {
	Index                int          `json:"index"`
	Status               RunStatus    `json:"status"`
	Error                *RunError    `json:"error,omitempty"`
	Usage                []ModelUsage `json:"usage"`
	DurationMilliseconds int64        `json:"duration_ms"`
	TraceID              string       `json:"trace_id,omitempty"`
}

// TurnSummary is the range of each measure over the repetitions that had a turn with this sequence number.
type TurnSummary struct {
	Sequence             int      `json:"sequence"`
	Repetitions          int      `json:"repetitions"`
	InputTokens          IntRange `json:"input_tokens"`
	CachedInputTokens    IntRange `json:"cached_input_tokens"`
	OutputTokens         IntRange `json:"output_tokens"`
	Reads                IntRange `json:"reads"`
	DurationMilliseconds IntRange `json:"duration_ms"`
}

type IntRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

func rangeOf(value int64) IntRange {
	return IntRange{Min: value, Max: value}
}

func (r IntRange) widen(value int64) IntRange {
	return IntRange{Min: min(r.Min, value), Max: max(r.Max, value)}
}

// UsageTotals totals tokens by model, in the order models were first used.
type UsageTotals struct {
	byModel []ModelUsage
}

func (u *UsageTotals) Add(model string, inputTokens, cachedInputTokens, outputTokens int) {
	if inputTokens == 0 && cachedInputTokens == 0 && outputTokens == 0 {
		return
	}
	for index := range u.byModel {
		if u.byModel[index].Model == model {
			u.byModel[index].InputTokens += inputTokens
			u.byModel[index].CachedInputTokens += cachedInputTokens
			u.byModel[index].OutputTokens += outputTokens
			return
		}
	}
	u.byModel = append(u.byModel, ModelUsage{Model: model, InputTokens: inputTokens, CachedInputTokens: cachedInputTokens, OutputTokens: outputTokens})
}

func (u *UsageTotals) List() []ModelUsage {
	if u.byModel == nil {
		return []ModelUsage{}
	}
	return u.byModel
}

func encodeForSummary(value any) string {
	// Decoded JSON always encodes.
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// Summarize counts the repetitions by status, each check's passes, and the range of each turn's measures.
// Checks and turns keep the order in which they first appear.
func Summarize(result RunResult) Summary {
	summary := Summary{
		Target:      result.Target,
		Runs:        len(result.Repetitions),
		Checks:      []CheckSummary{},
		Repetitions: []RepetitionSummary{},
	}
	checkIndex := make(map[string]int)
	turnIndex := make(map[int]int)
	for _, repetition := range result.Repetitions {
		switch repetition.Status {
		case RunStatusPassed:
			summary.Passed++
		case RunStatusFailed:
			summary.Failed++
		default:
			summary.Errors++
		}
		summary.Repetitions = append(summary.Repetitions, RepetitionSummary{
			Index:                repetition.Index,
			Status:               repetition.Status,
			Error:                repetition.Error,
			Usage:                repetition.Usage,
			DurationMilliseconds: repetition.DurationMilliseconds,
			TraceID:              repetition.TraceID,
		})
		for _, check := range repetition.Checks {
			index, seen := checkIndex[check.ID]
			if !seen {
				index = len(summary.Checks)
				checkIndex[check.ID] = index
				summary.Checks = append(summary.Checks, CheckSummary{ID: check.ID})
			}
			summary.Checks[index].Recorded++
			if check.Passed {
				summary.Checks[index].Passed++
			}
		}
		for _, turn := range repetition.Turns {
			index, seen := turnIndex[turn.Sequence]
			if !seen {
				turnIndex[turn.Sequence] = len(summary.Turns)
				summary.Turns = append(summary.Turns, TurnSummary{
					Sequence:             turn.Sequence,
					Repetitions:          1,
					InputTokens:          rangeOf(int64(turn.InputTokens)),
					CachedInputTokens:    rangeOf(int64(turn.CachedInputTokens)),
					OutputTokens:         rangeOf(int64(turn.OutputTokens)),
					Reads:                rangeOf(int64(turn.Reads)),
					DurationMilliseconds: rangeOf(turn.DurationMilliseconds),
				})
				continue
			}
			turnSummary := &summary.Turns[index]
			turnSummary.Repetitions++
			turnSummary.InputTokens = turnSummary.InputTokens.widen(int64(turn.InputTokens))
			turnSummary.CachedInputTokens = turnSummary.CachedInputTokens.widen(int64(turn.CachedInputTokens))
			turnSummary.OutputTokens = turnSummary.OutputTokens.widen(int64(turn.OutputTokens))
			turnSummary.Reads = turnSummary.Reads.widen(int64(turn.Reads))
			turnSummary.DurationMilliseconds = turnSummary.DurationMilliseconds.widen(turn.DurationMilliseconds)
		}
	}
	return summary
}

// Write prints the summary as a table for people.
func (s Summary) Write(w io.Writer) error {
	table := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(table, "%s (%s): %d runs, %d passed, %d failed, %d errors\n",
		s.Target.Name, s.Target.Kind, s.Runs, s.Passed, s.Failed, s.Errors)
	for _, check := range s.Checks {
		fmt.Fprintf(table, "%s\t%d/%d\n", check.ID, check.Passed, check.Recorded)
	}
	for _, repetition := range s.Repetitions {
		fmt.Fprintf(table, "run %d\t%s\t%s\t%s\n",
			repetition.Index, repetition.Status, formatMilliseconds(repetition.DurationMilliseconds), formatUsage(repetition.Usage))
		if repetition.Error != nil {
			fmt.Fprintf(table, "\terror at %s: %s\n", repetition.Error.Step, repetition.Error.Message)
		}
	}
	for _, turn := range s.Turns {
		fmt.Fprintf(table, "turn %d\tinput %s (cached %s)\toutput %s\treads %s\t%s\n",
			turn.Sequence,
			formatRange(turn.InputTokens, formatTokens),
			formatRange(turn.CachedInputTokens, formatTokens),
			formatRange(turn.OutputTokens, formatTokens),
			formatRange(turn.Reads, func(value int64) string { return fmt.Sprint(value) }),
			formatRange(turn.DurationMilliseconds, formatMilliseconds))
	}
	return table.Flush()
}

func formatUsage(usage []ModelUsage) string {
	parts := make([]string, 0, len(usage))
	for _, model := range usage {
		parts = append(parts, fmt.Sprintf("%s input %s (cached %s) output %s",
			model.Model, formatTokens(int64(model.InputTokens)), formatTokens(int64(model.CachedInputTokens)), formatTokens(int64(model.OutputTokens))))
	}
	if len(parts) == 0 {
		return "no model usage"
	}
	return strings.Join(parts, ", ")
}

func formatRange(r IntRange, format func(int64) string) string {
	if r.Min == r.Max {
		return format(r.Min)
	}
	return format(r.Min) + "–" + format(r.Max)
}

func formatTokens(tokens int64) string {
	switch {
	case tokens < 1000:
		return fmt.Sprint(tokens)
	case tokens < 10_000:
		return fmt.Sprintf("%.1fk", float64(tokens)/1000)
	default:
		return fmt.Sprintf("%dk", (tokens+500)/1000)
	}
}

func formatMilliseconds(milliseconds int64) string {
	return fmt.Sprintf("%.1fs", float64(milliseconds)/1000)
}
