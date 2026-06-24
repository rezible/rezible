package eino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentruncitation"
	"github.com/rezible/rezible/ent/agentrunfinding"
	"github.com/rezible/rezible/ent/agentrunresult"
	"github.com/rezible/rezible/ent/agentruntoolcall"
	"github.com/rezible/rezible/ent/agenttask"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

type (
	AgentWorkflowResult struct {
		Content   string
		Data      map[string]any
		Citations []rez.AgentRunCitationInput
		Findings  []rez.AgentRunFindingInput
	}

	agentWorkflow interface {
		Kind() rez.AgentWorkflowKind
		Validate(context.Context, *ent.AgentTask, *ent.AgentRun) error
		Run(context.Context, *WorkflowRecorder, *ent.AgentTask, *ent.AgentRun) (*AgentWorkflowResult, error)
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
	workflows map[rez.AgentWorkflowKind]agentWorkflow
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
		workflows: make(map[rez.AgentWorkflowKind]agentWorkflow),
	}

	s.registerWorkflow(&incidentContextPackWorkflow{incidents: incidents, modelFactory: s.modelProv})
	s.registerWorkflow(&alertInvestigationWorkflow{alerts: alerts, modelFactory: s.modelProv, aiEnabled: cfg.AI.Enabled})

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

func (s *AgentService) GetTask(ctx context.Context, id uuid.UUID) (*ent.AgentTask, error) {
	return s.db.Client(ctx).AgentTask.Get(ctx, id)
}

func (s *AgentService) ListTasks(ctx context.Context, params rez.ListAgentTasksParams) (*ent.ListResult[ent.AgentTask], error) {
	query := s.db.Client(ctx).AgentTask.Query().
		Order(agenttask.ByUpdatedAt(sql.OrderDesc()))
	if params.WorkflowKind != "" {
		query.Where(agenttask.WorkflowKind(string(params.WorkflowKind)))
	}
	if params.TriggerKind != "" {
		query.Where(agenttask.TriggerKind(params.TriggerKind))
	}
	if params.SubjectType != "" && params.SubjectID != uuid.Nil {
		subjectFilter := fmt.Sprintf(`{"subjects":[{"type":%q,"id":%q}]}`, params.SubjectType, params.SubjectID.String())
		query.Where(func(sel *sql.Selector) {
			sel.Where(sql.P(func(b *sql.Builder) {
				b.Ident(sel.C(agenttask.FieldWorkflowInput)).WriteString(" @> ").Arg(subjectFilter)
			}))
		})
	}
	return ent.DoListQuery[ent.AgentTask, *ent.AgentTaskQuery](ctx, query, params.ListParams)
}

func (s *AgentService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Query().
		Where(agentrun.ID(id)).
		WithAgentTask().
		Only(ctx)
}

func (s *AgentService) ListRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AgentRun], error) {
	query := s.db.Client(ctx).AgentRun.Query().
		WithAgentTask().
		Order(agentrun.ByUpdatedAt(sql.OrderDesc()))
	if params.AgentTaskID != uuid.Nil {
		query.Where(agentrun.AgentTaskID(params.AgentTaskID))
	}
	if params.Status != "" {
		query.Where(agentrun.StatusEQ(params.Status))
	}
	if params.WorkflowKind != "" {
		query.Where(agentrun.HasAgentTaskWith(agenttask.WorkflowKind(string(params.WorkflowKind))))
	}
	return ent.DoListQuery[ent.AgentRun, *ent.AgentRunQuery](ctx, query, params.ListParams)
}

func (s *AgentService) ListRunCitations(ctx context.Context, runID uuid.UUID) ([]*ent.AgentRunCitation, error) {
	return s.db.Client(ctx).AgentRunCitation.Query().
		Where(agentruncitation.AgentRunID(runID)).
		Order(agentruncitation.ByCreatedAt(), agentruncitation.ByID()).
		All(ctx)
}

func (s *AgentService) ListRunFindings(ctx context.Context, runID uuid.UUID) ([]*ent.AgentRunFinding, error) {
	return s.db.Client(ctx).AgentRunFinding.Query().
		Where(agentrunfinding.AgentRunID(runID)).
		Order(agentrunfinding.BySequence()).
		All(ctx)
}

func (s *AgentService) ListRunToolCalls(ctx context.Context, runID uuid.UUID) ([]*ent.AgentRunToolCall, error) {
	return s.db.Client(ctx).AgentRunToolCall.Query().
		Where(agentruntoolcall.AgentRunID(runID)).
		Order(agentruntoolcall.ByCreatedAt(), agentruntoolcall.ByID()).
		All(ctx)
}

func (s *AgentService) GetRunResult(ctx context.Context, runID uuid.UUID) (*ent.AgentRunResult, error) {
	return s.db.Client(ctx).AgentRunResult.Query().
		Where(agentrunresult.AgentRunID(runID)).
		Only(ctx)
}

var insertAgentRunRequestJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (s *AgentService) CreateTask(ctx context.Context, params rez.CreateAgentTaskRequest) (*ent.AgentTask, error) {
	if params.WorkflowKind == "" {
		return nil, fmt.Errorf("missing agent workflow kind")
	}
	if _, ok := s.workflows[params.WorkflowKind]; !ok {
		return nil, fmt.Errorf("unknown agent workflow kind %q", params.WorkflowKind)
	}
	ownerID := params.OwnerUserID
	if ownerID == uuid.Nil {
		userID, userOK := execution.GetContext(ctx).UserID()
		if !userOK {
			return nil, fmt.Errorf("missing task owner user")
		}
		ownerID = userID
	}
	if params.WorkflowInput == nil {
		params.WorkflowInput = map[string]any{}
	}
	if params.TriggerKind == "" {
		params.TriggerKind = "manual"
	}
	if params.TriggerPayload == nil {
		params.TriggerPayload = map[string]any{}
	}

	var created *ent.AgentTask
	if err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		task, createErr := tx.AgentTask.Create().
			SetOwnerUserID(ownerID).
			SetWorkflowKind(string(params.WorkflowKind)).
			SetWorkflowInput(redactJSON(params.WorkflowInput)).
			SetTriggerKind(params.TriggerKind).
			SetTriggerPayload(redactJSON(params.TriggerPayload)).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		created = task.Unwrap()
		return nil
	}); err != nil {
		return nil, err
	}

	if _, runErr := s.RequestTaskRun(ctx, created.ID); runErr != nil {
		return nil, runErr
	}
	return s.GetTask(ctx, created.ID)
}

func (s *AgentService) RequestTaskRun(ctx context.Context, taskId uuid.UUID) (*ent.AgentRun, error) {
	var run *ent.AgentRun
	err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "agent_task_run_attempt", taskId.String()); lockErr != nil {
			return lockErr
		}
		task, taskErr := tx.AgentTask.Get(ctx, taskId)
		if taskErr != nil {
			return fmt.Errorf("get agent task: %w", taskErr)
		}
		if _, ok := s.workflows[rez.AgentWorkflowKind(task.WorkflowKind)]; !ok {
			return fmt.Errorf("unknown agent workflow kind %q", task.WorkflowKind)
		}
		latest, latestErr := tx.AgentRun.Query().
			Where(agentrun.AgentTaskIDEQ(taskId)).
			Order(agentrun.ByStartedAt(sql.OrderDesc())).
			First(ctx)
		if latestErr != nil && !ent.IsNotFound(latestErr) {
			return fmt.Errorf("query latest agent run: %w", latestErr)
		}
		attempt := 1
		if latest != nil {
			attempt = latest.Attempt + 1
		}
		created, createErr := tx.AgentRun.Create().
			SetAgentTaskID(taskId).
			SetAttempt(attempt).
			SetStatus(agentrun.StatusQueued).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent run: %w", createErr)
		}
		run = created.Unwrap()
		if _, jobErr := s.jobs.Insert(ctx, jobs.RunAgentWorkflow{AgentRunID: run.ID}, insertAgentRunRequestJobOpts); jobErr != nil {
			return fmt.Errorf("enqueue agent workflow: %w", jobErr)
		}
		s.publishEvent(ctx, rez.EventOnAgentRunQueued{
			AgentRunID:   run.ID,
			AgentTaskID:  task.ID,
			WorkflowKind: rez.AgentWorkflowKind(task.WorkflowKind),
		})
		return nil
	})
	return run, err
}

func (s *AgentService) RunWorkflow(ctx context.Context, id uuid.UUID) error {
	run, getErr := s.GetRun(ctx, id)
	if getErr != nil {
		return fmt.Errorf("get agent run: %w", getErr)
	}
	task := run.Edges.AgentTask
	if task == nil {
		return fmt.Errorf("agent run missing task")
	}
	runCtx := contextForTaskOwner(ctx, task)
	return s.runWorkflow(runCtx, task, run)
}

func (s *AgentService) runWorkflow(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) error {
	if run.Status == agentrun.StatusSucceeded || run.Status == agentrun.StatusRunning {
		return nil
	}
	w, ok := s.workflows[rez.AgentWorkflowKind(task.WorkflowKind)]
	if !ok {
		return s.failRun(ctx, task, run, fmt.Errorf("unknown agent workflow kind %q", task.WorkflowKind))
	}
	if validationErr := w.Validate(ctx, task, run); validationErr != nil {
		return s.failRun(ctx, task, run, validationErr)
	}
	started, startErr := s.setRunStarted(ctx, task, run)
	if startErr != nil {
		return s.failRun(ctx, task, run, startErr)
	}
	recorder := NewWorkflowRecorder(s.db, started.ID)
	result, runErr := w.Run(ctx, recorder, task, started)
	if runErr != nil {
		return s.failRun(ctx, task, started, runErr)
	}
	if completeErr := s.completeRun(ctx, task, started, result); completeErr != nil {
		return completeErr
	}
	return nil
}

func (s *AgentService) setRunStarted(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) (*ent.AgentRun, error) {
	started, err := s.db.Client(ctx).AgentRun.UpdateOneID(run.ID).
		SetStatus(agentrun.StatusRunning).
		SetStartedAt(time.Now().UTC()).
		ClearFinishedAt().
		ClearErrorMessage().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark agent run started: %w", err)
	}
	s.publishEvent(ctx, rez.EventOnAgentRunStarted{
		AgentRunID:   started.ID,
		AgentTaskID:  task.ID,
		WorkflowKind: rez.AgentWorkflowKind(task.WorkflowKind),
	})
	return started, nil
}

func (s *AgentService) completeRun(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun, result *AgentWorkflowResult) error {
	if result == nil {
		result = &AgentWorkflowResult{}
	}
	now := time.Now().UTC()
	if err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		citations, citationErr := createRunCitations(ctx, tx, run.ID, result.Citations)
		if citationErr != nil {
			return citationErr
		}
		if err := createRunFindings(ctx, tx, run.ID, result.Findings, citations); err != nil {
			return err
		}
		if strings.TrimSpace(result.Content) != "" {
			if _, resultErr := tx.AgentRunResult.Create().
				SetAgentRunID(run.ID).
				SetContent(result.Content).
				SetData(redactJSON(result.Data)).
				Save(ctx); resultErr != nil {
				return fmt.Errorf("create agent run result: %w", resultErr)
			}
		}
		if _, updateErr := tx.AgentRun.UpdateOneID(run.ID).
			SetStatus(agentrun.StatusSucceeded).
			SetFinishedAt(now).
			ClearErrorMessage().
			Save(ctx); updateErr != nil {
			return fmt.Errorf("mark agent run succeeded: %w", updateErr)
		}
		return nil
	}); err != nil {
		return err
	}
	s.publishEvent(ctx, rez.EventOnAgentRunCompleted{
		AgentRunID:   run.ID,
		AgentTaskID:  task.ID,
		WorkflowKind: rez.AgentWorkflowKind(task.WorkflowKind),
	})
	return nil
}

func (s *AgentService) failRun(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun, err error) error {
	errMsg := err.Error()
	_, updateErr := s.db.Client(ctx).AgentRun.UpdateOneID(run.ID).
		SetStatus(agentrun.StatusFailed).
		SetFinishedAt(time.Now().UTC()).
		SetErrorMessage(errMsg).
		Save(ctx)
	if updateErr != nil {
		err = errors.Join(err, fmt.Errorf("mark agent run failed: %w", updateErr))
	}
	s.publishEvent(ctx, rez.EventOnAgentRunFailed{
		AgentRunID:   run.ID,
		AgentTaskID:  task.ID,
		WorkflowKind: rez.AgentWorkflowKind(task.WorkflowKind),
		ErrorMessage: errMsg,
	})
	return err
}

func createRunCitations(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []rez.AgentRunCitationInput) ([]*ent.AgentRunCitation, error) {
	res := make([]*ent.AgentRunCitation, 0, len(inputs))
	for _, input := range inputs {
		if strings.TrimSpace(input.Summary) == "" {
			continue
		}
		if err := validateCitationInput(runID, input); err != nil {
			return nil, err
		}
		create := tx.AgentRunCitation.Create().
			SetAgentRunID(runID).
			SetCitationKind(input.CitationKind).
			SetSummary(input.Summary).
			SetSnapshot(redactJSON(input.Snapshot))
		if input.DomainEntityType != "" {
			create.SetDomainEntityType(input.DomainEntityType)
		}
		if input.DomainEntityID != uuid.Nil {
			create.SetDomainEntityID(input.DomainEntityID)
		}
		if input.KnowledgeEntityID != uuid.Nil {
			create.SetKnowledgeEntityID(input.KnowledgeEntityID)
		}
		if input.KnowledgeRelationshipID != uuid.Nil {
			create.SetKnowledgeRelationshipID(input.KnowledgeRelationshipID)
		}
		if input.KnowledgeEvidenceID != uuid.Nil {
			create.SetKnowledgeEvidenceID(input.KnowledgeEvidenceID)
		}
		if input.AgentTaskID != uuid.Nil {
			create.SetAgentTaskID(input.AgentTaskID)
		}
		if input.AgentRunToolCallID != uuid.Nil {
			create.SetAgentRunToolCallID(input.AgentRunToolCallID)
		}
		citation, createErr := create.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create agent run citation: %w", createErr)
		}
		res = append(res, citation)
	}
	return res, nil
}

func validateCitationInput(runID uuid.UUID, input rez.AgentRunCitationInput) error {
	_ = runID
	targets := 0
	if input.DomainEntityType != "" || input.DomainEntityID != uuid.Nil {
		if input.DomainEntityType == "" || input.DomainEntityID == uuid.Nil {
			return fmt.Errorf("domain citation requires type and id")
		}
		targets++
	}
	for _, id := range []uuid.UUID{input.KnowledgeEntityID, input.KnowledgeRelationshipID, input.KnowledgeEvidenceID, input.AgentTaskID, input.AgentRunToolCallID} {
		if id != uuid.Nil {
			targets++
		}
	}
	if targets != 1 {
		return fmt.Errorf("agent citation requires exactly one target")
	}
	if input.CitationKind == "" {
		return fmt.Errorf("agent citation kind is required")
	}
	return nil
}

func createRunFindings(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []rez.AgentRunFindingInput, citations []*ent.AgentRunCitation) error {
	for i, input := range inputs {
		if strings.TrimSpace(input.Content) == "" {
			continue
		}
		finding, createErr := tx.AgentRunFinding.Create().
			SetAgentRunID(runID).
			SetSequence(i + 1).
			SetFindingKind(input.FindingKind).
			SetContent(input.Content).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent run finding: %w", createErr)
		}
		for _, citation := range input.Citations {
			if citation.CitationIndex <= 0 || citation.CitationIndex > len(citations) {
				continue
			}
			if citation.SupportKind == "" {
				citation.SupportKind = "supports"
			}
			if _, linkErr := tx.AgentRunFindingCitation.Create().
				SetAgentRunFindingID(finding.ID).
				SetAgentRunCitationID(citations[citation.CitationIndex-1].ID).
				SetSupportKind(citation.SupportKind).
				Save(ctx); linkErr != nil {
				return fmt.Errorf("create agent run finding citation: %w", linkErr)
			}
		}
	}
	return nil
}

func contextForTaskOwner(ctx context.Context, task *ent.AgentTask) context.Context {
	exec := execution.GetContext(ctx)
	tenantID := task.TenantID
	exec.ActorKind = execution.KindUser
	exec.Auth.TenantID = &tenantID
	exec.Auth.UserID = &task.OwnerUserID
	return execution.SetContext(ctx, exec)
}

func redactJSON(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "cookie") || strings.Contains(lower, "authorization") || strings.Contains(lower, "password") {
			out[key] = "redacted"
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			out[key] = redactJSON(typed)
		default:
			out[key] = value
		}
	}
	return out
}

func workflowInputSubject(input map[string]any, subjectType string) uuid.UUID {
	subjects, _ := input["subjects"].([]any)
	for _, item := range subjects {
		obj, _ := item.(map[string]any)
		if obj["type"] != subjectType {
			continue
		}
		idText, _ := obj["id"].(string)
		id, err := uuid.Parse(idText)
		if err == nil {
			return id
		}
	}
	return uuid.Nil
}

func encodePromptContext(ctx *rez.AgentWorkflowContext) (string, error) {
	payload := map[string]any{
		"schema":      ctx.PromptSchema,
		"generatedAt": ctx.GeneratedAt,
		"context":     ctx.Context,
		"items":       ctx.Items,
	}
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *AgentService) publishEvent(ctx context.Context, event any) {
	if err := s.msgs.PublishEvent(ctx, event); err != nil {
		s.logger.WarnContext(ctx, "failed to publish agent event", "error", err)
	}
}
