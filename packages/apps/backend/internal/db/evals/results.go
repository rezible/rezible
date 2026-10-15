package evals

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TargetKind says how a target is evaluated: a workflow over a dataset, or an agent through a scenario.
type TargetKind string

const (
	TargetKindWorkflow TargetKind = "workflow"
	TargetKindAgent    TargetKind = "agent"
)

// RunStatus is a repetition's outcome. passed: every check passed. failed: a check failed. error: the
// target or the framework failed, so the repetition is incomplete.
type RunStatus string

const (
	RunStatusPassed RunStatus = "passed"
	RunStatusFailed RunStatus = "failed"
	RunStatusError  RunStatus = "error"
)

type (
	// Target is what an evaluation runs.
	Target struct {
		Name        string     `json:"name"`
		Kind        TargetKind `json:"kind"`
		Description string     `json:"description"`
	}

	// RunResult is every repetition of one target.
	RunResult struct {
		Target      Target       `json:"target"`
		Repetitions []Repetition `json:"repetitions"`
	}

	// Repetition is one run of a target.
	Repetition struct {
		Index                int          `json:"index"`
		Status               RunStatus    `json:"status"`
		Error                *RunError    `json:"error,omitempty"`
		Checks               []Check      `json:"checks"`
		Usage                []ModelUsage `json:"usage"`
		DurationMilliseconds int64        `json:"duration_ms"`
		// Turns are an agent's turns, in sequence.
		Turns []Turn `json:"turns,omitempty"`
		// Outputs is what the target produced, for people to read.
		Outputs any    `json:"outputs,omitempty"`
		TraceID string `json:"trace_id,omitempty"`
	}

	// Check is one named judgement over a repetition.
	Check struct {
		ID       string `json:"id"`
		Passed   bool   `json:"passed"`
		Summary  string `json:"summary"`
		Expected any    `json:"expected,omitempty"`
		Observed any    `json:"observed,omitempty"`
	}

	ModelUsage struct {
		Model        string `json:"model"`
		InputTokens  int    `json:"input_tokens"`
		OutputTokens int    `json:"output_tokens"`
	}

	Turn struct {
		Sequence             int    `json:"sequence"`
		Status               string `json:"status"`
		FinishReason         string `json:"finish_reason"`
		Model                string `json:"model"`
		InputTokens          int    `json:"input_tokens"`
		OutputTokens         int    `json:"output_tokens"`
		Reads                int    `json:"reads"`
		DurationMilliseconds int64  `json:"duration_ms"`
	}

	RunError struct {
		// Step is where the repetition failed: a workflow's case ID, or an agent scenario's phase.
		Step    string `json:"step"`
		Message string `json:"message"`
	}
)

// Report writes repetition-<n>.json for each repetition and summary.json under dir, prints the summary, and
// returns it.
func Report(result RunResult, dir string, writer io.Writer) (Summary, error) {
	summary := Summarize(result)
	if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
		return summary, fmt.Errorf("create %s: %w", dir, mkdirErr)
	}
	files := map[string]any{"summary.json": summary}
	for _, repetition := range result.Repetitions {
		files[fmt.Sprintf("repetition-%d.json", repetition.Index)] = RunResult{
			Target:      result.Target,
			Repetitions: []Repetition{repetition},
		}
	}
	for name, content := range files {
		encoded, encodeErr := json.MarshalIndent(content, "", "  ")
		if encodeErr != nil {
			return summary, fmt.Errorf("encode %s: %w", name, encodeErr)
		}
		if writeErr := os.WriteFile(filepath.Join(dir, name), encoded, 0o644); writeErr != nil {
			return summary, fmt.Errorf("write %s: %w", name, writeErr)
		}
	}
	if printErr := summary.Write(writer); printErr != nil {
		return summary, fmt.Errorf("print summary: %w", printErr)
	}
	if _, printErr := fmt.Fprintf(writer, "results in %s\n", dir); printErr != nil {
		return summary, fmt.Errorf("print summary: %w", printErr)
	}
	return summary, nil
}
