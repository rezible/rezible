package eino

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
	"github.com/rezible/rezible/ent/agentcase"
	"github.com/rezible/rezible/ent/agentcaseartifact"
	"github.com/rezible/rezible/ent/agentcaseconclusion"
	"github.com/rezible/rezible/ent/agentcasestep"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/jobs"
)

type (
	AgentCaseArtifact struct {
		Kind     agentcaseartifact.Kind
		Role     string
		Name     string
		Payload  map[string]any
		Redacted bool
	}

	AgentCaseConclusion struct {
		Kind               string
		Summary            string
		Confidence         string
		RecommendedActions []string
		Limitations        []string
		Payload            map[string]any
	}

	AgentCaseStep struct {
		Kind        agentcasestep.Kind
		Title       string
		Summary     string
		Input       map[string]any
		Output      map[string]any
		Artifacts   []AgentCaseArtifact
		Conclusions []AgentCaseConclusion
	}

	AgentWorkflowResult struct {
		Summary string
		Steps   []AgentCaseStep
	}

	agentWorkflow interface {
		Kind() agentrun.WorkflowKind
		Validate(context.Context, *ent.AgentRun) error
		Run(context.Context, *ent.AgentRun) (*AgentWorkflowResult, error)
	}
)

type AgentService struct {
	cfg       rez.AiConfig
	logger    *slog.Logger
	db        rez.Database
	jobs      rez.JobService
	msgs      rez.MessageService
	incidents rez.IncidentService
	alerts    rez.AlertService
	modelProv ModelProvider
	workflows map[agentrun.WorkflowKind]agentWorkflow
}

func NewAgentService(cfg rez.Config, tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, incidents rez.IncidentService, alerts rez.AlertService) (*AgentService, error) {
	s := &AgentService{
		cfg:       cfg.AI,
		logger:    tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:        db,
		jobs:      jobSvc,
		msgs:      msgSvc,
		incidents: incidents,
		alerts:    alerts,
		modelProv: newChatModelProvider(cfg.AI),
		workflows: make(map[agentrun.WorkflowKind]agentWorkflow),
	}

	s.registerWorkflow(&incidentContextPackWorkflow{incidents: incidents, modelFactory: s.modelProv})
	s.registerWorkflow(&alertInvestigationWorkflow{alerts: alerts, modelFactory: s.modelProv, aiEnabled: cfg.AI.Enabled})
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
		rez.NewEventHandler("eino.AgentService.OnIncidentUpdated", s.onIncidentUpdated),
		rez.NewEventHandler("eino.AgentService.OnIncidentImpactsUpdated", s.onIncidentImpactsUpdated),
	)
	return errors.Join(eventsErr)
}

func (s *AgentService) registerWorkflow(w agentWorkflow) {
	s.workflows[w.Kind()] = w
}

func (s *AgentService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Get(ctx, id)
}

func (s *AgentService) GetCase(ctx context.Context, id uuid.UUID) (*ent.AgentCase, error) {
	return s.db.Client(ctx).AgentCase.Get(ctx, id)
}

func (s *AgentService) ListCases(ctx context.Context, params rez.ListAgentCasesParams) (*ent.ListResult[ent.AgentCase], error) {
	query := s.db.Client(ctx).AgentCase.Query().
		Order(agentcase.ByUpdatedAt(sql.OrderDesc()))
	if params.Status != "" {
		query.Where(agentcase.StatusEQ(params.Status))
	}
	if params.WorkflowKind != "" {
		query.Where(agentcase.WorkflowKindEQ(agentcase.WorkflowKind(params.WorkflowKind)))
	}
	if params.SubjectKind != "" {
		query.Where(agentcase.SubjectKind(params.SubjectKind))
	}
	if params.SubjectID != uuid.Nil {
		query.Where(agentcase.SubjectID(params.SubjectID))
	}
	return ent.DoListQuery[ent.AgentCase, *ent.AgentCaseQuery](ctx, query, params.ListParams)
}

func (s *AgentService) ListCaseSteps(ctx context.Context, id uuid.UUID) ([]*ent.AgentCaseStep, error) {
	return s.db.Client(ctx).AgentCaseStep.Query().
		Where(agentcasestep.AgentCaseID(id)).
		Order(agentcasestep.BySequence(sql.OrderAsc())).
		All(ctx)
}

func (s *AgentService) ListCaseArtifacts(ctx context.Context, id uuid.UUID) ([]*ent.AgentCaseArtifact, error) {
	return s.db.Client(ctx).AgentCaseArtifact.Query().
		Where(agentcaseartifact.AgentCaseID(id)).
		Order(agentcaseartifact.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
}

func (s *AgentService) ListCaseConclusions(ctx context.Context, id uuid.UUID) ([]*ent.AgentCaseConclusion, error) {
	return s.db.Client(ctx).AgentCaseConclusion.Query().
		Where(agentcaseconclusion.AgentCaseID(id)).
		Order(agentcaseconclusion.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
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
	if params.AgentCaseID != uuid.Nil {
		query.Where(agentrun.AgentCaseID(params.AgentCaseID))
	}
	if params.SubjectKind != "" {
		query.Where(agentrun.SubjectKind(params.SubjectKind))
	}
	if params.SubjectID != uuid.Nil {
		query.Where(agentrun.SubjectID(params.SubjectID))
	}
	return ent.DoListQuery[ent.AgentRun, *ent.AgentRunQuery](ctx, query, params.ListParams)
}

var insertAgentRunRequestJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (s *AgentService) CreateCase(ctx context.Context, params rez.AgentCaseRequest) (*ent.AgentCase, error) {
	if params.WorkflowKind == "" {
		return nil, fmt.Errorf("missing agent workflow kind")
	}
	if _, ok := s.workflows[params.WorkflowKind]; !ok {
		return nil, fmt.Errorf("unknown agent workflow kind %q", params.WorkflowKind)
	}
	if params.TriggerMetadata == nil {
		params.TriggerMetadata = map[string]any{}
	}
	title := params.Title
	if title == "" {
		title = string(params.WorkflowKind)
		if params.SubjectKind != "" {
			title += " for " + params.SubjectKind
		}
	}

	var created *ent.AgentCase
	if err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		create := tx.AgentCase.Create().
			SetStatus(agentcase.StatusOpen).
			SetTitle(title).
			SetWorkflowKind(agentcase.WorkflowKind(params.WorkflowKind)).
			SetTriggerMetadata(params.TriggerMetadata)
		if params.Query != "" {
			create.SetQuery(params.Query)
		}
		if params.SubjectKind != "" {
			create.SetSubjectKind(params.SubjectKind)
		}
		if params.SubjectID != uuid.Nil {
			create.SetSubjectID(params.SubjectID)
		}
		res, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent case: %w", createErr)
		}
		created = res.Unwrap()
		return nil
	}); err != nil {
		return nil, err
	}

	if _, runErr := s.RequestCaseRun(ctx, rez.AgentCaseRunRequest{
		AgentCaseID: created.ID,
		Metadata:    params.TriggerMetadata,
	}); runErr != nil {
		return nil, runErr
	}
	return s.GetCase(ctx, created.ID)
}

func (s *AgentService) RequestCaseRun(ctx context.Context, params rez.AgentCaseRunRequest) (*ent.AgentRun, error) {
	if params.AgentCaseID == uuid.Nil {
		return nil, fmt.Errorf("missing agent case id")
	}
	c, caseErr := s.GetCase(ctx, params.AgentCaseID)
	if caseErr != nil {
		return nil, fmt.Errorf("get agent case: %w", caseErr)
	}
	if c.WorkflowKind == "" {
		return nil, fmt.Errorf("agent case has no workflow kind")
	}
	subjectID := uuid.Nil
	if c.SubjectID != nil {
		subjectID = *c.SubjectID
	}
	return s.RequestRun(ctx, rez.AgentRunRequest{
		AgentCaseID:    c.ID,
		WorkflowKind:   agentrun.WorkflowKind(c.WorkflowKind),
		IdempotencyKey: params.IdempotencyKey,
		SubjectKind:    c.SubjectKind,
		SubjectID:      subjectID,
		Metadata:       params.Metadata,
	})
}

func (s *AgentService) RequestRun(ctx context.Context, params rez.AgentRunRequest) (*ent.AgentRun, error) {
	if params.AgentCaseID == uuid.Nil {
		return nil, fmt.Errorf("missing agent case id")
	}
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
			SetAgentCaseID(params.AgentCaseID).
			SetWorkflowKind(params.WorkflowKind).
			SetStatus(agentrun.StatusQueued).
			SetIdempotencyKey(runKey).
			SetTriggerMetadata(params.Metadata).
			SetModelMetadata(s.modelProv.ModelMetadata()).
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
	})
}

func (s *AgentService) getRunRequestIdempotencyKey(params rez.AgentRunRequest) (string, error) {
	if params.WorkflowKind == "" {
		return "", fmt.Errorf("missing agent workflow kind")
	}
	if params.IdempotencyKey != "" {
		return params.IdempotencyKey, nil
	}
	if params.AgentCaseID != uuid.Nil {
		return fmt.Sprintf("%s:%d", params.AgentCaseID, time.Now().UTC().UnixNano()), nil
	}
	if params.SubjectID == uuid.Nil {
		return params.SubjectKind, nil
	}
	return fmt.Sprintf("%s:%s:%s", params.WorkflowKind, params.SubjectKind, params.SubjectID), nil
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
	var started *ent.AgentRun
	if err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		update := tx.AgentRun.UpdateOneID(run.ID).
			SetStatus(agentrun.StatusRunning).
			SetStartedAt(time.Now().UTC()).
			ClearCompletedAt().
			ClearFailedAt().
			ClearErrorCode().
			ClearErrorMessage()
		res, startErr := update.Save(ctx)
		if startErr != nil {
			return fmt.Errorf("mark agent run started: %w", startErr)
		}
		started = res.Unwrap()
		if run.AgentCaseID != nil {
			if _, caseErr := tx.AgentCase.UpdateOneID(*run.AgentCaseID).
				SetStatus(agentcase.StatusRunning).
				ClearErrorCode().
				ClearErrorMessage().
				Save(ctx); caseErr != nil {
				return fmt.Errorf("mark agent case running: %w", caseErr)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, rez.EventOnAgentRunStarted{
		AgentRunID:   started.ID,
		WorkflowKind: run.WorkflowKind,
	})

	return started, nil
}

func (s *AgentService) completeRun(ctx context.Context, run *ent.AgentRun, result *AgentWorkflowResult) error {
	if result == nil {
		result = &AgentWorkflowResult{}
	}

	now := time.Now().UTC()
	saveTx := func(txCtx context.Context, tx *ent.Client) error {
		if run.AgentCaseID == nil {
			return fmt.Errorf("agent run has no case id")
		}
		nextSequence, seqErr := nextCaseStepSequence(txCtx, tx, *run.AgentCaseID)
		if seqErr != nil {
			return seqErr
		}
		steps := append([]AgentCaseStep{{
			Kind:      agentcasestep.KindSystem,
			Title:     "Model configuration",
			Summary:   "Captured model configuration for this run.",
			Artifacts: []AgentCaseArtifact{redactedModelCaseArtifact("model", s.cfg)},
		}}, result.Steps...)
		for _, step := range steps {
			createdStep, stepErr := createCaseStep(txCtx, tx, *run.AgentCaseID, run.ID, nextSequence, now, step)
			if stepErr != nil {
				return stepErr
			}
			if err := createCaseStepArtifacts(txCtx, tx, *run.AgentCaseID, run.ID, createdStep.ID, step.Artifacts); err != nil {
				return err
			}
			if err := createCaseStepConclusions(txCtx, tx, *run.AgentCaseID, run.ID, createdStep.ID, step.Conclusions); err != nil {
				return err
			}
			nextSequence++
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
		caseUpdate := tx.AgentCase.UpdateOneID(*run.AgentCaseID).
			SetStatus(agentcase.StatusCompleted)
		if result.Summary != "" {
			caseUpdate.SetSummary(result.Summary)
		}
		if _, caseErr := caseUpdate.Save(txCtx); caseErr != nil {
			return fmt.Errorf("mark agent case completed: %w", caseErr)
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
	updateErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		update := tx.AgentRun.UpdateOneID(run.ID).
			SetStatus(agentrun.StatusFailed).
			SetFailedAt(time.Now().UTC()).
			SetErrorCode(code).
			SetErrorMessage(errMsg)
		if _, saveErr := update.Save(ctx); saveErr != nil {
			return fmt.Errorf("mark agent run failed: %w", saveErr)
		}
		if run.AgentCaseID != nil {
			if _, caseErr := tx.AgentCase.UpdateOneID(*run.AgentCaseID).
				SetStatus(agentcase.StatusFailed).
				SetErrorCode(code).
				SetErrorMessage(errMsg).
				Save(ctx); caseErr != nil {
				return fmt.Errorf("mark agent case failed: %w", caseErr)
			}
		}
		return nil
	})
	if updateErr != nil {
		return errors.Join(err, updateErr)
	}
	s.publishEvent(ctx, rez.EventOnAgentRunFailed{
		AgentRunID:   run.ID,
		WorkflowKind: run.WorkflowKind,
		ErrorCode:    code,
		ErrorMessage: errMsg,
	})
	return err
}

func nextCaseStepSequence(ctx context.Context, tx *ent.Client, caseID uuid.UUID) (int, error) {
	latest, latestErr := tx.AgentCaseStep.Query().
		Where(agentcasestep.AgentCaseID(caseID)).
		Order(agentcasestep.BySequence(sql.OrderDesc())).
		First(ctx)
	if latestErr != nil {
		if ent.IsNotFound(latestErr) {
			return 1, nil
		}
		return 0, fmt.Errorf("query latest agent case step: %w", latestErr)
	}
	return latest.Sequence + 1, nil
}

func createCaseStep(ctx context.Context, tx *ent.Client, caseID, runID uuid.UUID, sequence int, now time.Time, step AgentCaseStep) (*ent.AgentCaseStep, error) {
	create := tx.AgentCaseStep.Create().
		SetAgentCaseID(caseID).
		SetAgentRunID(runID).
		SetSequence(sequence).
		SetKind(step.Kind).
		SetTitle(step.Title).
		SetStartedAt(now).
		SetCompletedAt(now)
	if step.Summary != "" {
		create.SetSummary(step.Summary)
	}
	if step.Input != nil {
		create.SetInput(step.Input)
	}
	if step.Output != nil {
		create.SetOutput(step.Output)
	}
	created, stepErr := create.Save(ctx)
	if stepErr != nil {
		return nil, fmt.Errorf("create agent case step: %w", stepErr)
	}
	return created, nil
}

func createCaseStepArtifacts(ctx context.Context, tx *ent.Client, caseID, runID, stepID uuid.UUID, artifacts []AgentCaseArtifact) error {
	for _, artifact := range artifacts {
		create := tx.AgentCaseArtifact.Create().
			SetAgentCaseID(caseID).
			SetAgentCaseStepID(stepID).
			SetAgentRunID(runID).
			SetKind(artifact.Kind).
			SetName(artifact.Name).
			SetPayload(artifact.Payload).
			SetRedacted(artifact.Redacted)
		if artifact.Role != "" {
			create.SetRole(artifact.Role)
		}
		if _, artifactErr := create.Save(ctx); artifactErr != nil {
			return fmt.Errorf("create agent case artifact: %w", artifactErr)
		}
	}
	return nil
}

func createCaseStepConclusions(ctx context.Context, tx *ent.Client, caseID, runID, stepID uuid.UUID, conclusions []AgentCaseConclusion) error {
	for _, conclusion := range conclusions {
		create := tx.AgentCaseConclusion.Create().
			SetAgentCaseID(caseID).
			SetAgentCaseStepID(stepID).
			SetAgentRunID(runID).
			SetKind(conclusion.Kind)
		if conclusion.Summary != "" {
			create.SetSummary(conclusion.Summary)
		}
		if conclusion.Confidence != "" {
			create.SetConfidence(conclusion.Confidence)
		}
		if len(conclusion.RecommendedActions) > 0 {
			create.SetRecommendedActions(conclusion.RecommendedActions)
		}
		if len(conclusion.Limitations) > 0 {
			create.SetLimitations(conclusion.Limitations)
		}
		if conclusion.Payload != nil {
			create.SetPayload(conclusion.Payload)
		}
		if _, conclusionErr := create.Save(ctx); conclusionErr != nil {
			return fmt.Errorf("create agent case conclusion: %w", conclusionErr)
		}
	}
	return nil
}

func (s *AgentService) publishEvent(ctx context.Context, event any) {
	if err := s.msgs.PublishEvent(ctx, event); err != nil {
		s.logger.WarnContext(ctx, "failed to publish agent event", "error", err)
	}
}
