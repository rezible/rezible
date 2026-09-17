package db

import (
	"context"
	"log/slog"

	rez "github.com/rezible/rezible"
)

type AiWorkflowRunner struct {
	logger *slog.Logger
}

func NewAiWorkflowRunner(tel rez.TelemetryService) *AiWorkflowRunner {
	return &AiWorkflowRunner{
		logger: tel.NewLogger(rez.NewLoggerOptions{Name: "ai_workflow_runner"}),
	}
}

func (r *AiWorkflowRunner) ExecuteWorkflow(ctx context.Context, name string, run func(context.Context) error) error {
	r.logger.InfoContext(ctx, "running AI workflow", "workflow", name)
	return run(ctx)
}
