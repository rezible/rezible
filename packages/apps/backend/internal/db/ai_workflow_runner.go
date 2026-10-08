package db

import (
	"context"
	"log/slog"
)

type AiWorkflowRunner struct{}

func NewAiWorkflowRunner() *AiWorkflowRunner {
	return &AiWorkflowRunner{}
}

func (r *AiWorkflowRunner) ExecuteWorkflow(ctx context.Context, name string, run func(context.Context) error) error {
	slog.InfoContext(ctx, "running AI workflow", "workflow", name)
	return run(ctx)
}
