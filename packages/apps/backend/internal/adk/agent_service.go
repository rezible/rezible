package adk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunartifact"
	"github.com/rezible/rezible/jobs"
)

type (
	AgentRunArtifact struct {
		Kind     agentrunartifact.Kind
		Name     string
		Payload  map[string]any
		Redacted bool
	}

	AgentWorkflowResult struct {
		Artifacts []AgentRunArtifact
	}

	agentWorkflow interface {
		Kind() agentrun.WorkflowKind
		Validate(context.Context, *ent.AgentRun) error
		Run(context.Context, *ent.AgentRun) (*AgentWorkflowResult, error)
	}
)

type AgentService struct {
	cfg          rez.AiConfig
	logger       *slog.Logger
	db           rez.Database
	jobs         rez.JobService
	msgs         rez.MessageService
	incidents    rez.IncidentService
	modelFactory LanguageModelFactory
	workflows    map[agentrun.WorkflowKind]agentWorkflow
}

func NewAgentService(cfg rez.Config, tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, incidents rez.IncidentService) (*AgentService, error) {
	s := &AgentService{
		cfg:          cfg.AI,
		logger:       tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:           db,
		jobs:         jobSvc,
		msgs:         msgSvc,
		incidents:    incidents,
		modelFactory: newLanguageModelFactory(cfg.AI),
		workflows:    make(map[agentrun.WorkflowKind]agentWorkflow),
	}

	s.registerWorkflow(&incidentContextPackWorkflow{incidents: incidents, modelFactory: s.modelFactory})
	s.registerWorkflow(stubWorkflow{kind: agentrun.WorkflowKindAlertInvestigation})
	s.registerWorkflow(stubWorkflow{kind: agentrun.WorkflowKindRetrospectiveAnalysis})

	jobs.RegisterWorkerFunc(func(ctx context.Context, args jobs.RunAgentWorkflow) error {
		return s.RunWorkflow(ctx, args.AgentRunID)
	})

	if handlersErr := s.registerMessageHandlers(); handlersErr != nil {
		return nil, handlersErr
	}

	return s, nil
}

func (s *AgentService) registerMessageHandlers() error {
	eventsErr := s.msgs.AddEventHandlers(
		rez.NewEventHandler("adk.AgentService.OnIncidentUpdated", s.onIncidentUpdated),
		rez.NewEventHandler("adk.AgentService.OnIncidentImpactsUpdated", s.onIncidentImpactsUpdated),
	)
	return errors.Join(eventsErr)
}

func (s *AgentService) registerWorkflow(w agentWorkflow) {
	s.workflows[w.Kind()] = w
}

func (s *AgentService) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return s.requestIncidentContextPackRun(ctx, ev.IncidentId, "incident_updated")
}

func (s *AgentService) onIncidentImpactsUpdated(ctx context.Context, ev *rez.EventOnIncidentImpactsUpdated) error {
	return s.requestIncidentContextPackRun(ctx, ev.IncidentId, "incident_impacts_updated")
}

func (s *AgentService) requestIncidentContextPackRun(ctx context.Context, incidentID uuid.UUID, trigger string) error {
	if !s.cfg.Enabled || incidentID == uuid.Nil {
		return nil
	}
	bucket := time.Now().UTC().Truncate(5 * time.Minute).Format(time.RFC3339)
	_, reqErr := s.RequestRun(ctx, rez.AgentRunRequest{
		WorkflowKind:   agentrun.WorkflowKindIncidentContextPack,
		IdempotencyKey: fmt.Sprintf("incident-context-pack:auto:%s:%s", incidentID, bucket),
		SubjectKind:    "incident",
		SubjectID:      incidentID,
		Metadata: map[string]any{
			"trigger": trigger,
			"bucket":  bucket,
		},
	})
	if reqErr != nil {
		return fmt.Errorf("request incident context pack run: %w", reqErr)
	}
	return nil
}

func (s *AgentService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Get(ctx, id)
}

func (s *AgentService) ListRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AgentRun], error) {
	query := s.db.Client(ctx).AgentRun.Query().
		Order(agentrun.ByUpdatedAt(sql.OrderDesc()))
	if params.WorkflowKind != "" {
		query.Where(agentrun.WorkflowKindEQ(params.WorkflowKind))
	}
	if params.Status != "" {
		query.Where(agentrun.StatusEQ(params.Status))
	}
	if params.SubjectKind != "" {
		query.Where(agentrun.SubjectKind(params.SubjectKind))
	}
	if params.SubjectID != uuid.Nil {
		query.Where(agentrun.SubjectID(params.SubjectID))
	}
	return ent.DoListQuery[ent.AgentRun, *ent.AgentRunQuery](ctx, query, params.ListParams)
}

func (s *AgentService) ListRunArtifacts(ctx context.Context, id uuid.UUID) ([]*ent.AgentRunArtifact, error) {
	if _, getErr := s.GetRun(ctx, id); getErr != nil {
		return nil, fmt.Errorf("get agent run: %w", getErr)
	}
	artifacts, queryErr := s.db.Client(ctx).AgentRunArtifact.Query().
		Where(agentrunartifact.AgentRunID(id)).
		Order(agentrunartifact.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query agent run artifacts: %w", queryErr)
	}
	return artifacts, nil
}

func (s *AgentService) RequestRun(ctx context.Context, params rez.AgentRunRequest) (*ent.AgentRun, error) {
	if _, ok := s.workflows[params.WorkflowKind]; !ok {
		return nil, fmt.Errorf("unknown agent workflow kind %q", params.WorkflowKind)
	}

	runKey, keyErr := s.getRunRequestIdempotencyKey(params)
	if keyErr != nil {
		return nil, fmt.Errorf("invalid run request: %w", keyErr)
	}

	getOrCreateRunTx := func(ctx context.Context, tx *ent.Client) (*ent.AgentRun, error) {
		queryExisting := tx.AgentRun.Query().
			Where(agentrun.WorkflowKindEQ(params.WorkflowKind), agentrun.IdempotencyKey(runKey))
		existing, queryErr := queryExisting.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("query existing agent run: %w", queryErr)
		}
		if existing != nil {
			return existing, nil
		}

		create := tx.AgentRun.Create().
			SetWorkflowKind(params.WorkflowKind).
			SetStatus(agentrun.StatusQueued).
			SetIdempotencyKey(runKey).
			SetTriggerMetadata(params.Metadata).
			SetModelMetadata(s.modelFactory.ModelMetadata()).
			SetQueuedAt(time.Now().UTC())
		if params.SubjectKind != "" {
			create.SetSubjectKind(params.SubjectKind)
		}
		if params.SubjectID != uuid.Nil {
			create.SetSubjectID(params.SubjectID)
		}
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create agent run: %w", createErr)
		}

		return created, nil
	}

	var run *ent.AgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		res, runErr := getOrCreateRunTx(ctx, client)
		if runErr != nil {
			return runErr
		}
		run = res.Unwrap()
		if reqErr := s.enqueueAgentRunRequest(ctx, run); reqErr != nil {
			return fmt.Errorf("enqueue request: %w", reqErr)
		}
		return nil
	})
}

func (s *AgentService) getRunRequestIdempotencyKey(params rez.AgentRunRequest) (string, error) {
	if params.WorkflowKind == "" {
		return "", fmt.Errorf("missing agent workflow kind")
	}
	if params.IdempotencyKey != "" {
		return params.IdempotencyKey, nil
	}
	if params.SubjectID == uuid.Nil {
		return params.SubjectKind, nil
	}
	return fmt.Sprintf("%s:%s:%s", params.WorkflowKind, params.SubjectKind, params.SubjectID), nil
}

var insertAgentRunRequestJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (s *AgentService) enqueueAgentRunRequest(ctx context.Context, run *ent.AgentRun) error {
	_, jobErr := s.jobs.Insert(ctx, jobs.RunAgentWorkflow{AgentRunID: run.ID}, insertAgentRunRequestJobOpts)
	if jobErr != nil {
		return fmt.Errorf("enqueue agent workflow: %w", jobErr)
	}

	s.publishEvent(ctx, rez.EventOnAgentRunQueued{
		AgentRunID:   run.ID,
		WorkflowKind: run.WorkflowKind,
		SubjectKind:  run.SubjectKind,
	})

	return nil
}

func (s *AgentService) RunWorkflow(ctx context.Context, id uuid.UUID) error {
	run, getErr := s.GetRun(ctx, id)
	if getErr != nil {
		return fmt.Errorf("get agent run: %w", getErr)
	}
	return s.runWorkflow(ctx, run)
}

func (s *AgentService) runWorkflow(ctx context.Context, run *ent.AgentRun) error {
	if run.Status == agentrun.StatusSucceeded || run.Status == agentrun.StatusRunning {
		return nil
	}
	w, ok := s.workflows[run.WorkflowKind]
	if !ok {
		return s.failRun(ctx, run, "unknown_workflow", fmt.Errorf("unknown agent workflow kind %q", run.WorkflowKind))
	}
	if validationErr := w.Validate(ctx, run); validationErr != nil {
		return s.failRun(ctx, run, "invalid_run", validationErr)
	}
	started, startErr := s.setRunStarted(ctx, run)
	if startErr != nil {
		return s.failRun(ctx, run, "set_run_started", startErr)
	}
	result, runErr := w.Run(ctx, started)
	if runErr != nil {
		return s.failRun(ctx, started, "workflow_error", runErr)
	}
	if completeErr := s.completeRun(ctx, started, result); completeErr != nil {
		return completeErr
	}

	return nil
}

func (s *AgentService) setRunStarted(ctx context.Context, run *ent.AgentRun) (*ent.AgentRun, error) {
	update := s.db.Client(ctx).AgentRun.UpdateOneID(run.ID).
		SetStatus(agentrun.StatusRunning).
		SetStartedAt(time.Now().UTC()).
		ClearCompletedAt().
		ClearFailedAt().
		ClearErrorCode().
		ClearErrorMessage()
	started, startErr := update.Save(ctx)
	if startErr != nil {
		return nil, fmt.Errorf("mark agent run started: %w", startErr)
	}

	s.publishEvent(ctx, rez.EventOnAgentRunStarted{
		AgentRunID:   started.ID,
		WorkflowKind: run.WorkflowKind,
	})

	return started, nil
}

func (s *AgentService) completeRun(ctx context.Context, run *ent.AgentRun, result *AgentWorkflowResult) error {
	artifacts := []AgentRunArtifact{redactedModelArtifact("model", s.cfg)}
	if result == nil {
		result = &AgentWorkflowResult{}
	}
	artifacts = append(artifacts, result.Artifacts...)

	now := time.Now().UTC()
	saveTx := func(txCtx context.Context, tx *ent.Client) error {
		if len(artifacts) > 0 {
			builders := make([]*ent.AgentRunArtifactCreate, len(artifacts))
			for i, artifact := range artifacts {
				builders[i] = tx.AgentRunArtifact.Create().
					SetAgentRunID(run.ID).
					SetKind(artifact.Kind).
					SetName(artifact.Name).
					SetPayload(artifact.Payload).
					SetRedacted(artifact.Redacted)
			}
			createErr := tx.AgentRunArtifact.CreateBulk(builders...).Exec(txCtx)
			if createErr != nil {
				return fmt.Errorf("create agent run artifact: %w", createErr)
			}
		}

		update := tx.AgentRun.UpdateOneID(run.ID).
			SetStatus(agentrun.StatusSucceeded).
			SetCompletedAt(now).
			ClearFailedAt().
			ClearErrorCode().
			ClearErrorMessage()
		if _, updateErr := update.Save(txCtx); updateErr != nil {
			return fmt.Errorf("mark agent run completed: %w", updateErr)
		}
		return nil
	}
	if txErr := s.db.WithTx(ctx, saveTx); txErr != nil {
		return txErr
	}
	s.publishEvent(ctx, rez.EventOnAgentRunCompleted{
		AgentRunID:   run.ID,
		WorkflowKind: run.WorkflowKind,
	})
	return nil
}

func (s *AgentService) failRun(ctx context.Context, run *ent.AgentRun, code string, err error) error {
	errMsg := err.Error()
	update := s.db.Client(ctx).AgentRun.UpdateOneID(run.ID).
		SetStatus(agentrun.StatusFailed).
		SetFailedAt(time.Now().UTC()).
		SetErrorCode(code).
		SetErrorMessage(errMsg)
	if _, updateErr := update.Save(ctx); updateErr != nil {
		return errors.Join(err, fmt.Errorf("mark agent run failed: %w", updateErr))
	}
	s.publishEvent(ctx, rez.EventOnAgentRunFailed{
		AgentRunID:   run.ID,
		WorkflowKind: run.WorkflowKind,
		ErrorCode:    code,
		ErrorMessage: errMsg,
	})
	return err
}

func (s *AgentService) publishEvent(ctx context.Context, event any) {
	if err := s.msgs.PublishEvent(ctx, event); err != nil {
		s.logger.WarnContext(ctx, "failed to publish agent event", "error", err)
	}
}
