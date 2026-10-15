package evals

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	check := func(id string, passed bool) Check {
		return Check{ID: id, Passed: passed}
	}
	turn := func(sequence, input, output, reads int, duration int64) Turn {
		return Turn{Sequence: sequence, InputTokens: input, OutputTokens: output, Reads: reads, DurationMilliseconds: duration}
	}
	workflow := Target{Name: "judge", Kind: TargetKindWorkflow}
	agent := Target{Name: "checkout", Kind: TargetKindAgent}
	usage := []ModelUsage{{Model: "m", InputTokens: 10, OutputTokens: 1}}
	stepError := &RunError{Step: "phase 2", Message: "turn failed"}

	cases := []struct {
		name   string
		result RunResult
		want   Summary
	}{
		{
			name: "workflow cases counted across repetitions",
			result: RunResult{Target: workflow, Repetitions: []Repetition{
				{Index: 1, Status: RunStatusPassed, Checks: []Check{check("a", true), check("b", true)}, Usage: usage, DurationMilliseconds: 100},
				{Index: 2, Status: RunStatusFailed, Checks: []Check{check("a", true), check("b", false)}, Usage: usage, DurationMilliseconds: 200},
			}},
			want: Summary{
				Target: workflow, Runs: 2, Passed: 1, Failed: 1,
				Checks: []CheckSummary{{ID: "a", Passed: 2, Recorded: 2}, {ID: "b", Passed: 1, Recorded: 2}},
				Repetitions: []RepetitionSummary{
					{Index: 1, Status: RunStatusPassed, Usage: usage, DurationMilliseconds: 100},
					{Index: 2, Status: RunStatusFailed, Usage: usage, DurationMilliseconds: 200},
				},
			},
		},
		{
			name: "an errored agent repetition records only the checks it reached, and turns become ranges",
			result: RunResult{Target: agent, Repetitions: []Repetition{
				{
					Index: 1, Status: RunStatusPassed,
					Checks: []Check{check("report_published", true), check("answer_published", true)},
					Turns:  []Turn{turn(1, 40_000, 3_000, 6, 48_000), turn(2, 45_000, 1_000, 1, 25_000)},
				},
				{
					Index: 2, Status: RunStatusError, Error: stepError,
					Checks: []Check{check("report_published", false)},
					Turns:  []Turn{turn(1, 52_000, 5_000, 11, 71_000)},
				},
			}},
			want: Summary{
				Target: agent, Runs: 2, Passed: 1, Errors: 1,
				Checks: []CheckSummary{
					{ID: "report_published", Passed: 1, Recorded: 2},
					{ID: "answer_published", Passed: 1, Recorded: 1},
				},
				Repetitions: []RepetitionSummary{
					{Index: 1, Status: RunStatusPassed},
					{Index: 2, Status: RunStatusError, Error: stepError},
				},
				Turns: []TurnSummary{
					{
						Sequence: 1, Repetitions: 2,
						InputTokens:          IntRange{Min: 40_000, Max: 52_000},
						OutputTokens:         IntRange{Min: 3_000, Max: 5_000},
						Reads:                IntRange{Min: 6, Max: 11},
						DurationMilliseconds: IntRange{Min: 48_000, Max: 71_000},
					},
					{
						Sequence: 2, Repetitions: 1,
						InputTokens:          IntRange{Min: 45_000, Max: 45_000},
						OutputTokens:         IntRange{Min: 1_000, Max: 1_000},
						Reads:                IntRange{Min: 1, Max: 1},
						DurationMilliseconds: IntRange{Min: 25_000, Max: 25_000},
					},
				},
			},
		},
		{
			name:   "no repetitions",
			result: RunResult{Target: workflow},
			want:   Summary{Target: workflow, Checks: []CheckSummary{}, Repetitions: []RepetitionSummary{}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, Summarize(tc.result))
		})
	}
}
