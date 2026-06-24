package eino

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentruntoolcall"
)

type WorkflowRecorder struct {
	db    rez.Database
	runID uuid.UUID
}

func NewWorkflowRecorder(db rez.Database, runID uuid.UUID) *WorkflowRecorder {
	return &WorkflowRecorder{db: db, runID: runID}
}

func (r *WorkflowRecorder) RecordToolCall(ctx context.Context, toolName string, params map[string]any, result map[string]any, err error) (*ent.AgentRunToolCall, error) {
	now := time.Now().UTC()
	status := agentruntoolcall.StatusSucceeded
	errMsg := ""
	if err != nil {
		status = agentruntoolcall.StatusFailed
		errMsg = err.Error()
	}
	create := r.db.Client(ctx).AgentRunToolCall.Create().
		SetAgentRunID(r.runID).
		SetToolName(toolName).
		SetStatus(status).
		SetToolParams(redactJSON(params)).
		SetResult(redactJSON(result)).
		SetStartedAt(now).
		SetFinishedAt(now)
	if errMsg != "" {
		create.SetErrorMessage(errMsg)
	}
	toolCall, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("record agent tool call: %w", createErr)
	}
	return toolCall, nil
}
