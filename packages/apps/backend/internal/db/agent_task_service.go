package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/agents"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunresult"
	"github.com/rezible/rezible/ent/agenttask"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

type AgentTaskService struct {
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
	msgs   rez.MessageService
	reg    *agents.WorkflowRegistry
}

func NewAgentTaskService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, reg *agents.WorkflowRegistry) (*AgentTaskService, error) {
	s := &AgentTaskService{
		logger: tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:     db,
		jobs:   jobSvc,
		msgs:   msgSvc,
		reg:    reg,
	}
	jobs.RegisterWorkerFunc(s.handleRequestAgentTaskRun)
	jobs.RegisterWorkerFunc(s.handleRunAgentWorkflow)
	return s, nil
}

func (s *AgentTaskService) GetTask(ctx context.Context, id uuid.UUID) (*ent.AgentTask, error) {
	return s.db.Client(ctx).AgentTask.Get(ctx, id)
}

func (s *AgentTaskService) ListTasks(ctx context.Context, params rez.ListAgentTasksParams) (*ent.ListResult[ent.AgentTask], error) {
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

func (s *AgentTaskService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Query().
		Where(agentrun.ID(id)).
		WithAgentTask().
		Only(ctx)
}

func (s *AgentTaskService) ListRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AgentRun], error) {
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

func (s *AgentTaskService) GetRunResult(ctx context.Context, runID uuid.UUID) (*ent.AgentRunResult, error) {
	return s.db.Client(ctx).AgentRunResult.Query().
		Where(agentrunresult.AgentRunID(runID)).
		Only(ctx)
}

var requestAgentTaskRunJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

var runAgentWorkflowJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (s *AgentTaskService) CreateTask(ctx context.Context, params rez.CreateAgentTaskRequest) (*ent.AgentTask, error) {
	if params.WorkflowInput == nil {
		params.WorkflowInput = map[string]any{}
	}
	if validationErr := agents.ValidateWorkflowInput(params.WorkflowKind, params.WorkflowInput); validationErr != nil {
		return nil, validationErr
	}
	ownerID := params.OwnerUserID
	if ownerID == uuid.Nil {
		userID, userOK := execution.GetContext(ctx).UserID()
		if !userOK {
			return nil, fmt.Errorf("missing task owner user")
		}
		ownerID = userID
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
			SetWorkflowInput(agents.RedactPayload(params.WorkflowInput)).
			SetTriggerKind(params.TriggerKind).
			SetTriggerPayload(agents.RedactPayload(params.TriggerPayload)).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		created = task.Unwrap()
		if _, jobErr := s.jobs.Insert(ctx, jobs.RequestAgentTaskRun{AgentTaskID: created.ID}, requestAgentTaskRunJobOpts); jobErr != nil {
			return fmt.Errorf("enqueue agent task run request: %w", jobErr)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetTask(ctx, created.ID)
}

func (s *AgentTaskService) RequestNewTaskRun(ctx context.Context, taskID uuid.UUID) (*ent.AgentRun, error) {
	task, taskErr := s.GetTask(ctx, taskID)
	if taskErr != nil {
		return nil, fmt.Errorf("get agent task: %w", taskErr)
	}
	kind := agents.WorkflowKind(task.WorkflowKind)
	if validationErr := agents.ValidateWorkflowInput(kind, task.WorkflowInput); validationErr != nil {
		return nil, validationErr
	}
	run, runErr := s.CreateRequestedRun(ctx, task.ID)
	if runErr != nil {
		return nil, fmt.Errorf("create agent task run request: %w", runErr)
	}
	return run, nil
}

func (s *AgentTaskService) handleRequestAgentTaskRun(ctx context.Context, args jobs.RequestAgentTaskRun) error {
	_, err := s.CreateRequestedRun(ctx, args.AgentTaskID)
	return err
}

func (s *AgentTaskService) CreateRequestedRun(ctx context.Context, taskID uuid.UUID) (*ent.AgentRun, error) {
	var run *ent.AgentRun
	var task *ent.AgentTask
	err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "agent_task_run_attempt", taskID.String()); lockErr != nil {
			return lockErr
		}
		loadedTask, taskErr := tx.AgentTask.Get(ctx, taskID)
		if taskErr != nil {
			return fmt.Errorf("get agent task: %w", taskErr)
		}
		if validationErr := agents.ValidateWorkflowInput(agents.WorkflowKind(loadedTask.WorkflowKind), loadedTask.WorkflowInput); validationErr != nil {
			return validationErr
		}
		latest, latestErr := tx.AgentRun.Query().
			Where(agentrun.AgentTaskIDEQ(taskID)).
			Order(agentrun.ByAttempt(sql.OrderDesc())).
			First(ctx)
		if latestErr != nil && !ent.IsNotFound(latestErr) {
			return fmt.Errorf("query latest agent run: %w", latestErr)
		}
		attempt := 1
		if latest != nil {
			attempt = latest.Attempt + 1
		}
		created, createErr := tx.AgentRun.Create().
			SetAgentTaskID(taskID).
			SetAttempt(attempt).
			SetStatus(agentrun.StatusQueued).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent run: %w", createErr)
		}
		run = created.Unwrap()
		task = loadedTask.Unwrap()
		if _, jobErr := s.jobs.Insert(ctx, jobs.RunAgentWorkflow{AgentRunID: run.ID}, runAgentWorkflowJobOpts); jobErr != nil {
			return fmt.Errorf("enqueue agent workflow: %w", jobErr)
		}
		s.publishEvent(ctx, rez.AgentRunQueuedEvent{
			AgentRunID:   run.ID,
			AgentTaskID:  task.ID,
			WorkflowKind: agents.WorkflowKind(task.WorkflowKind),
		})
		return nil
	})
	return run, err
}

func (s *AgentTaskService) handleRunAgentWorkflow(ctx context.Context, args jobs.RunAgentWorkflow) error {
	return s.RunWorkflow(ctx, args.AgentRunID)
}

func (s *AgentTaskService) RunWorkflow(ctx context.Context, id uuid.UUID) error {
	run, getErr := s.GetRun(ctx, id)
	if getErr != nil {
		return fmt.Errorf("get agent run: %w", getErr)
	}
	task := run.Edges.AgentTask
	if task == nil {
		return fmt.Errorf("agent run missing task")
	}
	runCtx := s.contextForTask(ctx, task)
	return s.runWorkflow(runCtx, task, run)
}

func (s *AgentTaskService) runWorkflow(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) error {
	if run.Status == agentrun.StatusSucceeded || run.Status == agentrun.StatusRunning {
		return nil
	}
	handler, ok := s.reg.Get(agents.WorkflowKind(task.WorkflowKind))
	if !ok {
		return s.failRun(ctx, task, run, fmt.Errorf("unknown agent workflow runner %q", task.WorkflowKind))
	}
	started, startErr := s.UpdateRun(ctx, task, run, UpdateAgentRunParams{Status: agentrun.StatusRunning})
	if startErr != nil {
		return s.failRun(ctx, task, run, startErr)
	}
	result, runErr := handler(ctx, task, started)
	if runErr != nil {
		return s.failRun(ctx, task, started, runErr)
	}
	if _, completeErr := s.UpdateRun(ctx, task, started, UpdateAgentRunParams{Status: agentrun.StatusSucceeded, Result: result}); completeErr != nil {
		return s.failRun(ctx, task, started, completeErr)
	}
	return nil
}

type UpdateAgentRunParams struct {
	Status       agentrun.Status
	Result       *agents.RunResult
	ErrorMessage string
}

func (s *AgentTaskService) UpdateRun(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun, params UpdateAgentRunParams) (*ent.AgentRun, error) {
	now := time.Now().UTC()
	var updated *ent.AgentRun
	if err := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		switch params.Status {
		case agentrun.StatusRunning:
			saved, updateErr := tx.AgentRun.UpdateOneID(run.ID).
				SetStatus(agentrun.StatusRunning).
				SetStartedAt(now).
				ClearFinishedAt().
				ClearErrorMessage().
				Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("mark agent run started: %w", updateErr)
			}
			updated = saved.Unwrap()
		case agentrun.StatusSucceeded:
			result := params.Result
			if result == nil {
				result = &agents.RunResult{}
			}
			if strings.TrimSpace(result.Content) == "" {
				return fmt.Errorf("agent run result content is required")
			}
			citations, citationErr := s.createRunCitations(ctx, tx, run.ID, result.Citations)
			if citationErr != nil {
				return citationErr
			}
			if err := s.createRunFindings(ctx, tx, run.ID, result.Findings, citations); err != nil {
				return err
			}
			if _, resultErr := tx.AgentRunResult.Create().
				SetAgentRunID(run.ID).
				SetContent(result.Content).
				SetData(agents.RedactPayload(result.Data)).
				Save(ctx); resultErr != nil {
				return fmt.Errorf("create agent run result: %w", resultErr)
			}
			saved, updateErr := tx.AgentRun.UpdateOneID(run.ID).
				SetStatus(agentrun.StatusSucceeded).
				SetFinishedAt(now).
				ClearErrorMessage().
				Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("mark agent run succeeded: %w", updateErr)
			}
			updated = saved.Unwrap()
		case agentrun.StatusFailed:
			saved, updateErr := tx.AgentRun.UpdateOneID(run.ID).
				SetStatus(agentrun.StatusFailed).
				SetFinishedAt(now).
				SetErrorMessage(params.ErrorMessage).
				Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("mark agent run failed: %w", updateErr)
			}
			updated = saved.Unwrap()
		default:
			return fmt.Errorf("unsupported agent run status update %q", params.Status)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	s.publishRunUpdated(ctx, task, updated, params)
	return updated, nil
}

func (s *AgentTaskService) failRun(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun, err error) error {
	errMsg := err.Error()
	if _, updateErr := s.UpdateRun(ctx, task, run, UpdateAgentRunParams{Status: agentrun.StatusFailed, ErrorMessage: errMsg}); updateErr != nil {
		return fmt.Errorf("%w: %w", err, updateErr)
	}
	return err
}

func (s *AgentTaskService) createRunCitations(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []agents.RunCitationInput) ([]*ent.AgentRunCitation, error) {
	res := make([]*ent.AgentRunCitation, 0, len(inputs))
	for _, input := range inputs {
		if err := agents.ValidateRunCitationInput(input); err != nil {
			return nil, err
		}
		create := tx.AgentRunCitation.Create().
			SetAgentRunID(runID).
			SetCitationKind(input.CitationKind).
			SetSummary(input.Summary).
			SetSnapshot(agents.RedactPayload(input.Snapshot))
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

func (s *AgentTaskService) createRunFindings(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []agents.RunFindingInput, citations []*ent.AgentRunCitation) error {
	for i, input := range inputs {
		if strings.TrimSpace(input.Content) == "" {
			return fmt.Errorf("agent run finding content is required")
		}
		if strings.TrimSpace(input.FindingKind) == "" {
			return fmt.Errorf("agent run finding kind is required")
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
			if citation.CitationIndex <= 0 {
				return fmt.Errorf("agent run finding citation index must be greater than zero")
			}
			if citation.CitationIndex > len(citations) {
				return fmt.Errorf("agent run finding citation index %d out of range", citation.CitationIndex)
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

func (s *AgentTaskService) publishEvent(ctx context.Context, event any) {
	if err := s.msgs.PublishEvent(ctx, event); err != nil {
		s.logger.WarnContext(ctx, "failed to publish agent event", "error", err)
	}
}

func (s *AgentTaskService) publishRunUpdated(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun, params UpdateAgentRunParams) {
	switch params.Status {
	case agentrun.StatusRunning:
		s.publishEvent(ctx, rez.AgentRunStartedEvent{
			AgentRunID:   run.ID,
			AgentTaskID:  task.ID,
			WorkflowKind: agents.WorkflowKind(task.WorkflowKind),
		})
	case agentrun.StatusSucceeded:
		s.publishEvent(ctx, rez.AgentRunCompletedEvent{
			AgentRunID:   run.ID,
			AgentTaskID:  task.ID,
			WorkflowKind: agents.WorkflowKind(task.WorkflowKind),
		})
	case agentrun.StatusFailed:
		s.publishEvent(ctx, rez.AgentRunFailedEvent{
			AgentRunID:   run.ID,
			AgentTaskID:  task.ID,
			WorkflowKind: agents.WorkflowKind(task.WorkflowKind),
			ErrorMessage: params.ErrorMessage,
		})
	}
}

func (s *AgentTaskService) contextForTask(ctx context.Context, task *ent.AgentTask) context.Context {
	exec := execution.GetContext(ctx)
	exec.ActorKind = execution.KindUser
	exec.Auth.TenantID = new(task.TenantID)
	exec.Auth.UserID = &task.OwnerUserID
	return execution.SetContext(ctx, exec)
}
