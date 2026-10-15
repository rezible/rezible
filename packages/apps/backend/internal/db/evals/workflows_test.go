package evals

import (
	"context"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func TestMatchSubset(t *testing.T) {
	cases := []struct {
		name      string
		reference string
		output    string
		matches   bool
		why       string
	}{
		{
			name:      "keys the reference leaves out are not compared",
			reference: `{"decision": "hold"}`,
			output:    `{"decision": "hold", "explanation": "Weak evidence."}`,
			matches:   true,
		},
		{
			name:      "a different value",
			reference: `{"decision": "raise"}`,
			output:    `{"decision": "hold"}`,
			why:       `decision: expected "raise", got "hold"`,
		},
		{
			name:      "a missing key",
			reference: `{"decision": "hold"}`,
			output:    `{"explanation": "Weak evidence."}`,
			why:       "decision: missing",
		},
		{
			name:      "nested objects compare by the reference's keys",
			reference: `{"facts": {"signals": 2}}`,
			output:    `{"facts": {"signals": 2, "entities": 3}}`,
			matches:   true,
		},
		{
			name:      "the first difference in key order is reported",
			reference: `{"b": 1, "a": 1}`,
			output:    `{"b": 2, "a": 2}`,
			why:       "a: expected 1, got 2",
		},
		{
			name:      "arrays match element by element",
			reference: `{"reasons": [{"name": "breadth"}, {"name": "novelty"}]}`,
			output:    `{"reasons": [{"name": "breadth", "met": true}, {"name": "novelty", "met": true}]}`,
			matches:   true,
		},
		{
			name:      "arrays compare in order",
			reference: `{"reasons": ["breadth", "novelty"]}`,
			output:    `{"reasons": ["novelty", "breadth"]}`,
			why:       `reasons[0]: expected "breadth", got "novelty"`,
		},
		{
			name:      "arrays must have the same length",
			reference: `{"reasons": ["breadth"]}`,
			output:    `{"reasons": ["breadth", "novelty"]}`,
			why:       `reasons: expected ["breadth"], got ["breadth","novelty"]`,
		},
		{
			name:      "an object where the reference has a value",
			reference: `{"decision": "hold"}`,
			output:    `{"decision": {"value": "hold"}}`,
			why:       `decision: expected "hold", got {"value":"hold"}`,
		},
		{
			name:      "a whole output that is not an object",
			reference: `{"decision": "hold"}`,
			output:    `"hold"`,
			why:       `output: expected {"decision":"hold"}, got "hold"`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reference, output any
			require.NoError(t, json.Unmarshal([]byte(tc.reference), &reference))
			require.NoError(t, json.Unmarshal([]byte(tc.output), &output))
			matches, why := matchSubset(reference, output)
			require.Equal(t, tc.matches, matches)
			require.Equal(t, tc.why, why)
		})
	}
}

type stubWorkflow[I, O any] func(context.Context, I) (O, error)

func (f stubWorkflow[I, O]) Run(ctx context.Context, input I) (O, error) {
	return f(ctx, input)
}

// judgeAnswering returns a judge that gives answer(reference) for each dataset case, finding the case by its
// input, and adds usage to the caller's tally as a model call would.
func judgeAnswering(t *testing.T, answer func(reference rezai.SituationJudgeOutput) (rezai.SituationJudgeOutput, error)) rezai.JudgeSituationCandidateWorkflow {
	t.Helper()
	cases, loadErr := loadDataset(rezai.JudgeSituationCandidateDefinition.Name)
	require.NoError(t, loadErr)
	references := make(map[string]rezai.SituationJudgeOutput)
	for _, datasetCase := range cases {
		decoded, decodeErr := decodeCase[rezai.SituationJudgeInput, rezai.SituationJudgeOutput](datasetCase)
		require.NoError(t, decodeErr)
		key, _ := json.Marshal(decoded.input)
		references[string(key)] = decoded.reference
	}
	return stubWorkflow[rezai.SituationJudgeInput, rezai.SituationJudgeOutput](
		func(ctx context.Context, input rezai.SituationJudgeInput) (rezai.SituationJudgeOutput, error) {
			key, _ := json.Marshal(input)
			reference, found := references[string(key)]
			require.True(t, found, "the judge is asked about a dataset case")
			rezai.AddModelUsage(ctx, "test/model", 100, 60, 10)
			return answer(reference)
		})
}

func judgeTarget(t *testing.T, judge rezai.JudgeSituationCandidateWorkflow) WorkflowTarget {
	t.Helper()
	targets, targetsErr := Workflows{JudgeSituationCandidate: judge}.Targets()
	require.NoError(t, targetsErr)
	require.Len(t, targets, 1)
	return targets[0]
}

func TestEveryEmbeddedDatasetHasATarget(t *testing.T) {
	judge := judgeAnswering(t, func(reference rezai.SituationJudgeOutput) (rezai.SituationJudgeOutput, error) {
		return reference, nil
	})
	targets, targetsErr := Workflows{JudgeSituationCandidate: judge}.Targets()
	require.NoError(t, targetsErr, "every dataset decodes into its workflow's types and validates")

	var targetNames []string
	for _, target := range targets {
		targetNames = append(targetNames, target.Target().Name)
		require.Equal(t, TargetKindWorkflow, target.Target().Kind)
	}
	files, globErr := fs.Glob(datasetFiles, "datasets/*.json")
	require.NoError(t, globErr)
	var datasetNames []string
	for _, file := range files {
		datasetNames = append(datasetNames, strings.TrimSuffix(strings.TrimPrefix(file, "datasets/"), ".json"))
	}
	require.ElementsMatch(t, datasetNames, targetNames)
}

func TestDecodeCaseRejectsCasesThatNoLongerFit(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		reference string
		want      string
	}{
		{"an unknown input field", `{"facts": {}, "reasons": [{"reason": "breadth"}], "removed": 1}`, `{}`, "decode input"},
		{"an invalid input", `{"facts": {}, "reasons": []}`, `{}`, "validate input"},
		{"an unknown reference field", `{"facts": {}, "reasons": [{"reason": "breadth"}]}`, `{"verdict": "hold"}`, "decode reference"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			datasetCase := DatasetCase{TestCaseID: "case", Input: json.RawMessage(tc.input), Reference: json.RawMessage(tc.reference)}
			_, decodeErr := decodeCase[rezai.SituationJudgeInput, rezai.SituationJudgeOutput](datasetCase)
			require.ErrorContains(t, decodeErr, tc.want)
		})
	}
}

func TestCompareJudgeDecision(t *testing.T) {
	breadth := schematypes.SituationRaiseReasonBreadth
	novelty := schematypes.SituationRaiseReasonNovelty
	pastIncident := schematypes.SituationRaiseReasonPastIncident
	raise := func(reasons ...schematypes.SituationRaiseReason) rezai.SituationJudgeOutput {
		return rezai.SituationJudgeOutput{Decision: "raise", Reasons: reasons}
	}
	hold := rezai.SituationJudgeOutput{Decision: "hold"}

	cases := []struct {
		name      string
		output    rezai.SituationJudgeOutput
		reference rezai.SituationJudgeOutput
		passes    bool
	}{
		{"a matching hold", hold, hold, true},
		{"a hold where a raise is expected", hold, raise(breadth), false},
		{"a raise where a hold is expected", raise(breadth), hold, false},
		{"a raise citing every expected reason", raise(breadth, pastIncident), raise(breadth, pastIncident), true},
		{"a raise citing some of the expected reasons", raise(pastIncident), raise(breadth, pastIncident), true},
		{"a raise citing an unexpected reason", raise(breadth, novelty), raise(breadth, pastIncident), false},
		{"a raise citing nothing", raise(), raise(breadth), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			passes, why := compareJudgeDecision(tc.output, tc.reference)
			require.Equal(t, tc.passes, passes)
			require.Equal(t, tc.passes, why == "", "a failure says why")
		})
	}
}

func TestRunWorkflow(t *testing.T) {
	ctx := t.Context()

	t.Run("answers matching every reference pass", func(t *testing.T) {
		judge := judgeAnswering(t, func(reference rezai.SituationJudgeOutput) (rezai.SituationJudgeOutput, error) {
			reference.Explanation = "Free text, which is not compared."
			return reference, nil
		})
		target := judgeTarget(t, judge)
		result := RunWorkflow(ctx, target, 2)

		require.Equal(t, target.Target(), result.Target)
		require.Len(t, result.Repetitions, 2)
		caseCount := len(target.Cases())
		for index, repetition := range result.Repetitions {
			require.Equal(t, index+1, repetition.Index)
			require.Equal(t, RunStatusPassed, repetition.Status)
			require.Nil(t, repetition.Error)
			require.Len(t, repetition.Checks, caseCount)
			require.Len(t, repetition.Outputs, caseCount)
			require.Equal(t, []ModelUsage{
				{Model: "test/model", InputTokens: 100 * caseCount, CachedInputTokens: 60 * caseCount, OutputTokens: 10 * caseCount},
			}, repetition.Usage)
		}
	})

	t.Run("holding every case fails the raises", func(t *testing.T) {
		judge := judgeAnswering(t, func(rezai.SituationJudgeOutput) (rezai.SituationJudgeOutput, error) {
			return rezai.SituationJudgeOutput{Decision: "hold", Explanation: "Unsure."}, nil
		})
		target := judgeTarget(t, judge)
		repetition := RunWorkflow(ctx, target, 1).Repetitions[0]

		require.Equal(t, RunStatusFailed, repetition.Status)
		for index, datasetCase := range target.Cases() {
			var reference rezai.SituationJudgeOutput
			require.NoError(t, json.Unmarshal(datasetCase.Reference, &reference))
			check := repetition.Checks[index]
			require.Equal(t, datasetCase.TestCaseID, check.ID)
			require.Equal(t, reference.Decision == "hold", check.Passed, "case %s", check.ID)
		}
	})

	t.Run("a failing case is an error and the other cases still run", func(t *testing.T) {
		judge := judgeAnswering(t, func(reference rezai.SituationJudgeOutput) (rezai.SituationJudgeOutput, error) {
			if reference.Decision == "raise" {
				return rezai.SituationJudgeOutput{}, rezai.ErrWorkflowInvalidOutput
			}
			return reference, nil
		})
		target := judgeTarget(t, judge)
		repetition := RunWorkflow(ctx, target, 1).Repetitions[0]

		require.Equal(t, RunStatusError, repetition.Status)
		require.NotNil(t, repetition.Error)
		require.Len(t, repetition.Checks, len(target.Cases()))
		firstRaise := ""
		for _, datasetCase := range target.Cases() {
			var reference rezai.SituationJudgeOutput
			require.NoError(t, json.Unmarshal(datasetCase.Reference, &reference))
			if reference.Decision == "raise" {
				firstRaise = datasetCase.TestCaseID
				break
			}
		}
		require.Equal(t, firstRaise, repetition.Error.Step)
		for _, check := range repetition.Checks {
			_, hasOutput := repetition.Outputs.(map[string]any)[check.ID]
			require.Equal(t, check.Passed, hasOutput, "only the cases that ran have outputs: %s", check.ID)
		}
	})
}
