package evals

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rezible/rezible/pkg/execution"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	rez "github.com/rezible/rezible"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	rezai "github.com/rezible/rezible/pkg/ai"
)

// Datasets are named after their workflow and use Genkit's dataset shape, so the Developer UI can import one.
//
//go:embed datasets/*.json
var datasetFiles embed.FS

// DatasetCase is one case of a workflow dataset: an input, and the output expected for it.
type DatasetCase struct {
	TestCaseID string          `json:"testCaseId"`
	Input      json.RawMessage `json:"input"`
	Reference  json.RawMessage `json:"reference"`
}

// WorkflowTarget is a workflow with its dataset.
type WorkflowTarget interface {
	Target() Target
	Cases() []DatasetCase
	// RunCase runs the workflow on one case's input and compares the output with the case's reference. It
	// returns the workflow's error, if any, and no check.
	RunCase(ctx context.Context, index int) (*Check, any, error)
}

// WorkflowComparer decides whether a workflow's output matches a reference, and says why it does not.
type WorkflowComparer[O any] func(output, reference O) (bool, string)

// Workflows holds each workflow that has a dataset.
type Workflows struct {
	JudgeSituationCandidate rezai.JudgeSituationCandidateWorkflow
}

// Targets returns the evaluation target of each workflow.
func (w Workflows) Targets() ([]WorkflowTarget, error) {
	judge, judgeErr := NewWorkflowTarget(rezai.JudgeSituationCandidateDefinition, w.JudgeSituationCandidate, compareJudgeDecision)
	if judgeErr != nil {
		return nil, judgeErr
	}
	return []WorkflowTarget{judge}, nil
}

// NewWorkflowTarget loads the workflow's dataset. Every case's input and reference must decode into the
// workflow's types, with no unknown fields, and every input must be valid, so a dataset that no longer fits its
// workflow fails here rather than in a run. A nil compare matches the output against the reference with
// matchSubset.
func NewWorkflowTarget[I rez.ValidatingInput, O any](def rezai.AiPromptWorkflowDefinition[I, O], workflow rez.AiWorkflow[I, O], compare WorkflowComparer[O]) (WorkflowTarget, error) {
	if workflow == nil {
		return nil, fmt.Errorf("workflow %s: no workflow to evaluate", def.Name)
	}
	cases, loadErr := loadDataset(def.Name)
	if loadErr != nil {
		return nil, fmt.Errorf("workflow %s: %w", def.Name, loadErr)
	}
	target := &workflowTarget[I, O]{
		target: Target{
			Name:        def.Name,
			Kind:        TargetKindWorkflow,
			Description: def.Description,
		},
		workflow: workflow,
		compare:  compare,
		cases:    cases,
		decoded:  make([]decodedCase[I, O], len(cases)),
	}
	for index, datasetCase := range cases {
		decoded, decodeErr := decodeCase[I, O](datasetCase)
		if decodeErr != nil {
			return nil, fmt.Errorf("workflow %s: case %s: %w", def.Name, datasetCase.TestCaseID, decodeErr)
		}
		target.decoded[index] = *decoded
	}
	return target, nil
}

func loadDataset(name string) ([]DatasetCase, error) {
	encoded, readErr := datasetFiles.ReadFile("datasets/" + name + ".json")
	if readErr != nil {
		return nil, fmt.Errorf("read dataset: %w", readErr)
	}
	var cases []DatasetCase
	if decodeErr := decodeStrict(encoded, &cases); decodeErr != nil {
		return nil, fmt.Errorf("decode dataset: %w", decodeErr)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("dataset has no cases")
	}
	ids := mapset.NewSet[string]()
	for _, datasetCase := range cases {
		if strings.TrimSpace(datasetCase.TestCaseID) == "" {
			return nil, fmt.Errorf("a case has no testCaseId")
		}
		if !ids.Add(datasetCase.TestCaseID) {
			return nil, fmt.Errorf("case %s appears twice", datasetCase.TestCaseID)
		}
	}
	return cases, nil
}

type decodedCase[I, O any] struct {
	input     I
	reference O
	// referenceJSON is the reference as written, so a subset match compares only the keys it has.
	referenceJSON any
}

func decodeCase[I rez.ValidatingInput, O any](datasetCase DatasetCase) (*decodedCase[I, O], error) {
	var decoded decodedCase[I, O]
	if inputErr := decodeStrict(datasetCase.Input, &decoded.input); inputErr != nil {
		return nil, fmt.Errorf("decode input: %w", inputErr)
	}
	if validErr := decoded.input.Validate(); validErr != nil {
		return nil, fmt.Errorf("validate input: %w", validErr)
	}
	if referenceErr := decodeStrict(datasetCase.Reference, &decoded.reference); referenceErr != nil {
		return nil, fmt.Errorf("decode reference: %w", referenceErr)
	}
	if referenceErr := json.Unmarshal(datasetCase.Reference, &decoded.referenceJSON); referenceErr != nil {
		return nil, fmt.Errorf("decode reference: %w", referenceErr)
	}
	return &decoded, nil
}

func decodeStrict(encoded []byte, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	return decoder.Decode(into)
}

type workflowTarget[I, O any] struct {
	target   Target
	workflow rez.AiWorkflow[I, O]
	compare  WorkflowComparer[O]
	cases    []DatasetCase
	decoded  []decodedCase[I, O]
}

func (t *workflowTarget[I, O]) Target() Target {
	return t.target
}

func (t *workflowTarget[I, O]) Cases() []DatasetCase {
	return t.cases
}

func (t *workflowTarget[I, O]) RunCase(ctx context.Context, index int) (*Check, any, error) {
	decoded := t.decoded[index]
	output, runErr := t.workflow.Run(ctx, decoded.input)
	if runErr != nil {
		return nil, nil, runErr
	}
	// The output came from JSON, so it encodes and decodes.
	encoded, _ := json.Marshal(output)
	var observed any
	_ = json.Unmarshal(encoded, &observed)

	var passed bool
	var why string
	if t.compare != nil {
		passed, why = t.compare(output, decoded.reference)
	} else {
		passed, why = matchSubset(decoded.referenceJSON, observed)
	}
	summary := "The output matches the reference."
	if !passed {
		summary = why
	}
	check := &Check{
		ID:       t.cases[index].TestCaseID,
		Passed:   passed,
		Summary:  summary,
		Expected: decoded.referenceJSON,
		Observed: observed,
	}
	return check, observed, nil
}

// compareJudgeDecision passes when the decision matches and, for a raise, every cited reason is one the
// reference cites. A raise citing fewer reasons than the reference still passes; one citing none does not.
func compareJudgeDecision(output, reference rezai.SituationJudgeOutput) (bool, string) {
	if output.Decision != reference.Decision {
		return false, fmt.Sprintf("decision: expected %q, got %q", reference.Decision, output.Decision)
	}
	if sitjudg.Decision(output.Decision) != sitjudg.DecisionRaise {
		return true, ""
	}
	if len(output.Reasons) == 0 {
		return false, "the raise cites no reason"
	}
	expected := mapset.NewSet(reference.Reasons...)
	for _, reason := range output.Reasons {
		if !expected.Contains(reason) {
			return false, fmt.Sprintf("the raise cites %q, which the reference does not", reason)
		}
	}
	return true, ""
}

// RunWorkflow runs every case of the target's dataset, runs times over, one case at a time.
func RunWorkflow(ctx context.Context, target WorkflowTarget, runs int) RunResult {
	result := RunResult{Target: target.Target()}
	for index := 1; index <= runs; index++ {
		result.Repetitions = append(result.Repetitions, runWorkflowRepetition(ctx, target, index))
	}
	return result
}

// runWorkflowRepetition runs each case once. A case whose workflow fails is recorded as a failed check and
// makes the repetition an error; the remaining cases still run.
func runWorkflowRepetition(ctx context.Context, target WorkflowTarget, index int) Repetition {
	repetition := Repetition{Index: index}
	attrs := []attribute.KeyValue{
		attribute.String("eval_target", target.Target().Name),
		attribute.Int("eval_repetition", index),
	}
	// The repetition's function never fails: its failures are recorded on the repetition.
	_ = execution.Do(ctx, "eval.repetition", func(ctx context.Context) error {
		if spanContext := trace.SpanContextFromContext(ctx); spanContext.HasTraceID() {
			repetition.TraceID = spanContext.TraceID().String()
		}
		started := time.Now()
		var usage UsageTotals
		outputs := make(map[string]any)
		for caseIndex, datasetCase := range target.Cases() {
			caseCtx, totals := rezai.WithTurnUsage(ctx)
			check, output, runErr := target.RunCase(caseCtx, caseIndex)
			caseUsage := totals()
			usage.Add(caseUsage.Model, caseUsage.InputTokens, caseUsage.CachedInputTokens, caseUsage.OutputTokens)
			if runErr != nil {
				check = &Check{
					ID:      datasetCase.TestCaseID,
					Summary: fmt.Sprintf("The workflow failed: %s", runErr),
				}
				if repetition.Error == nil {
					repetition.Error = &RunError{Step: datasetCase.TestCaseID, Message: runErr.Error()}
				}
			} else {
				outputs[datasetCase.TestCaseID] = output
			}
			repetition.Checks = append(repetition.Checks, *check)
		}
		repetition.DurationMilliseconds = time.Since(started).Milliseconds()
		repetition.Usage = usage.List()
		repetition.Outputs = outputs
		return nil
	}, attrs...)
	repetition.Status = repetitionStatus(repetition)
	return repetition
}

func repetitionStatus(repetition Repetition) RunStatus {
	if repetition.Error != nil {
		return RunStatusError
	}
	for _, check := range repetition.Checks {
		if !check.Passed {
			return RunStatusFailed
		}
	}
	return RunStatusPassed
}

// matchSubset reports whether output holds everything in reference. Both are decoded JSON. An object matches
// when each of the reference's keys matches in the output, so keys the reference leaves out are not compared.
// An array matches when it has the same length and each element matches in order. Any other value must be
// equal. When they do not match, the string says where.
func matchSubset(reference, output any) (bool, string) {
	return matchSubsetAt("", reference, output)
}

func matchSubsetAt(path string, reference, output any) (bool, string) {
	switch ref := reference.(type) {
	case map[string]any:
		out, isObject := output.(map[string]any)
		if !isObject {
			return false, mismatch(path, reference, output)
		}
		// Sorted, so the difference reported is the same on every run.
		for _, key := range slices.Sorted(maps.Keys(ref)) {
			outValue, present := out[key]
			if !present {
				return false, fmt.Sprintf("%s: missing", joinPath(path, key))
			}
			if matched, why := matchSubsetAt(joinPath(path, key), ref[key], outValue); !matched {
				return false, why
			}
		}
		return true, ""
	case []any:
		out, isArray := output.([]any)
		if !isArray || len(out) != len(ref) {
			return false, mismatch(path, reference, output)
		}
		for index := range ref {
			if matched, why := matchSubsetAt(fmt.Sprintf("%s[%d]", path, index), ref[index], out[index]); !matched {
				return false, why
			}
		}
		return true, ""
	default:
		if !reflect.DeepEqual(reference, output) {
			return false, mismatch(path, reference, output)
		}
		return true, ""
	}
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func mismatch(path string, expected, observed any) string {
	if path == "" {
		path = "output"
	}
	return fmt.Sprintf("%s: expected %s, got %s", path, encodeForSummary(expected), encodeForSummary(observed))
}
