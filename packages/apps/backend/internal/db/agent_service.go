package db

import (
	"context"
	"fmt"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunresult"
	"github.com/rezible/rezible/ent/agenttask"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type AgentService struct {
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
	msgs   rez.MessageService
	agents rez.AgentRegistry
}

func NewAgentService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, agents rez.AgentRegistry) (*AgentService, error) {
	s := &AgentService{
		logger: tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:     db,
		jobs:   jobSvc,
		msgs:   msgSvc,
		agents: agents,
	}
	jobs.RegisterWorkerFunc(s.handleRunAgentWorkflow)
	return s, nil
}

func (s *AgentService) GetTask(ctx context.Context, id uuid.UUID) (*ent.AgentTask, error) {
	return s.db.Client(ctx).AgentTask.Get(ctx, id)
}

func (s *AgentService) ListTasks(ctx context.Context, params rez.ListAgentTasksParams) (*ent.ListResult[ent.AgentTask], error) {
	query := s.db.Client(ctx).AgentTask.Query().
		Order(agenttask.ByUpdatedAt(sql.OrderDesc()))
	if params.Workflow != "" {
		query.Where(agenttask.Workflow(params.Workflow))
	}
	if params.TriggerKind != "" {
		query.Where(agenttask.TriggerKind(params.TriggerKind))
	}
	if len(params.SubjectPredicates) > 0 {
		query.Where(agenttask.HasSubjectsWith(params.SubjectPredicates...))
	}
	return ent.DoListQuery[ent.AgentTask, *ent.AgentTaskQuery](ctx, query, params.ListParams)
}

func (s *AgentService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Query().
		Where(agentrun.ID(id)).
		WithTask().
		Only(ctx)
}

func (s *AgentService) ListRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AgentRun], error) {
	query := s.db.Client(ctx).AgentRun.Query().
		WithTask(func(q *ent.AgentTaskQuery) {
			if params.Workflow != "" {
				q.Where(agenttask.Workflow(params.Workflow))
			}
		}).
		Order(agentrun.ByUpdatedAt(sql.OrderDesc()))
	if params.AgentTaskID != uuid.Nil {
		query.Where(agentrun.AgentTaskID(params.AgentTaskID))
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.AgentRun, *ent.AgentRunQuery](ctx, query, params.ListParams)
}

func (s *AgentService) GetRunResult(ctx context.Context, runID uuid.UUID) (*ent.AgentRunResult, error) {
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

func (s *AgentService) CreateTask(ctx context.Context, params rez.CreateAgentTaskParams) (*ent.AgentTask, error) {
	//input, validationErr := s.reg.ValidateAndEncodeTaskInput(params.Workflow, params.WorkflowInput)
	//if validationErr != nil {
	//	return nil, fmt.Errorf("validation: %w", validationErr)
	//}
	input := []byte("{}")
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

	var created *ent.AgentTask
	return created, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		create := tx.AgentTask.Create().
			SetOwnerUserID(ownerID).
			SetWorkflow(params.Workflow).
			SetInput(input).
			SetTriggerKind(params.TriggerKind).
			SetTriggerMetadata(params.TriggerPayload)
		task, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		created = task.Unwrap()
		return nil
	})
}

func (s *AgentService) RequestNewTaskRun(ctx context.Context, taskID uuid.UUID) (*ent.AgentRun, error) {
	task, taskErr := s.GetTask(ctx, taskID)
	if taskErr != nil {
		return nil, fmt.Errorf("get agent task: %w", taskErr)
	}
	run, runErr := s.createTaskRun(ctx, task)
	if runErr != nil {
		return nil, fmt.Errorf("create agent task run request: %w", runErr)
	}
	return run, nil
}

func (s *AgentService) publishRunStatusEvent(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) {
	//s.publishEvent(ctx, rez.AgentRunStatusChangeEvent{
	//	AgentRunID:  run.ID,
	//	AgentTaskID: task.ID,
	//	Workflow:    task.Workflow,
	//	Status:      run.Status,
	//})
}

func (s *AgentService) createTaskRun(ctx context.Context, task *ent.AgentTask) (*ent.AgentRun, error) {
	var run *ent.AgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "agent_task_run_attempt", task.ID.String()); lockErr != nil {
			return lockErr
		}
		count, countErr := tx.AgentRun.Query().
			Where(agentrun.AgentTaskIDEQ(task.ID)).
			Count(ctx)
		if countErr != nil && !ent.IsNotFound(countErr) {
			return fmt.Errorf("query latest agent run: %w", countErr)
		}
		create := tx.AgentRun.Create().
			SetAgentTaskID(task.ID).
			SetAttempt(count + 1)

		created, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent run: %w", createErr)
		}

		if _, jobErr := s.jobs.Insert(ctx, jobs.RunAgentWorkflow{AgentRunID: created.ID}, runAgentWorkflowJobOpts); jobErr != nil {
			return fmt.Errorf("enqueue agent workflow: %w", jobErr)
		}
		s.publishRunStatusEvent(ctx, task, created)

		run = created.Unwrap()
		return nil
	})
}

func (s *AgentService) handleRunAgentWorkflow(ctx context.Context, args jobs.RunAgentWorkflow) error {
	return s.RunWorkflow(ctx, args.AgentRunID)
}

func (s *AgentService) RunWorkflow(ctx context.Context, id uuid.UUID) error {
	var run *ent.AgentRun
	var task *ent.AgentTask
	lookupRunTxFn := func(ctx context.Context, tx *ent.Client) error {
		queryRun := tx.AgentRun.Query().
			WithTask().
			Where(agentrun.IDEQ(id)).
			Where(agentrun.StartedAtIsNil())
		txRun, runErr := queryRun.Only(ctx)
		if runErr != nil {
			return fmt.Errorf("query agent run: %w", runErr)
		}
		run = txRun.Unwrap()

		txTask, taskErr := txRun.Edges.TaskOrErr()
		if taskErr != nil {
			return fmt.Errorf("edge task: %w", taskErr)
		}
		task = txTask.Unwrap()

		return nil
	}
	if txErr := s.db.WithTx(ctx, lookupRunTxFn); txErr != nil {
		return fmt.Errorf("agent run task: %w", txErr)
	}
	return s.runWorkflow(ctx, task, run)
}

func (s *AgentService) contextForTask(ctx context.Context, task *ent.AgentTask) context.Context {
	exec := execution.GetContext(ctx)
	exec.ActorKind = execution.KindUser
	exec.Auth.TenantID = new(task.TenantID)
	exec.Auth.UserID = &task.OwnerUserID
	return execution.SetContext(ctx, exec)
}

func (s *AgentService) runWorkflow(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) error {
	ctx = s.contextForTask(ctx, task)

	queryResult := s.db.Client(ctx).AgentRunResult.Query().
		Where(agentrunresult.AgentRunIDEQ(run.ID))
	resultExists, lookupResultErr := queryResult.Exist(ctx)
	if lookupResultErr != nil {
		return fmt.Errorf("query result: %w", lookupResultErr)
	} else if resultExists {
		return nil
	}

	/*
		runnerFn, runnerErr := s.reg.GetWorkflowRunner(task.Workflow)
		if runnerErr != nil {
			return fmt.Errorf("get workflow runner: %w", runnerErr)
		}

		startedRun, startErr := s.updateRun(ctx, run.ID, func(m *ent.AgentRunMutation) {
			m.SetStartedAt(time.Now().UTC())
		})
		if startErr != nil {
			return fmt.Errorf("set agent run startedAt: %w", startErr)
		}

		res := runnerFn(ctx, task, startedRun)
		if resErr := s.recordRunResult(ctx, startedRun, res); resErr != nil {
			slog.ErrorContext(ctx, "failed to record run result", "err", resErr)
			// TODO: retry?
		}
	*/
	return nil
}

func (s *AgentService) updateRun(ctx context.Context, id uuid.UUID, setFn func(*ent.AgentRunMutation)) (*ent.AgentRun, error) {
	var run *ent.AgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		update := tx.AgentRun.UpdateOneID(id)

		m := update.Mutation()
		setFn(m)

		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("save update: %w", updateErr)
		}
		run = updated.Unwrap()

		if startedAt, startedAtUpdated := m.StartedAt(); startedAtUpdated {
			//s.publishRunStatusEvent()
			slog.DebugContext(ctx, "run status updated", "id", run.ID, "started", startedAt)
		}
		// s.publishRunUpdated(ctx, task, updated, params)

		return nil
	})
}

/*
func (s *AgentService) createRunCitations(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []agents.TaskRunCitationInput) ([]*ent.AgentRunCitation, error) {
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

func (s *AgentService) createRunFindings(ctx context.Context, tx *ent.Client, runID uuid.UUID, inputs []agents.TaskRunFindingInput, citations []*ent.AgentRunCitation) error {
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

func (s *AgentService) publishEvent(ctx context.Context, event any) {
	if err := s.msgs.PublishEvent(ctx, event); err != nil {
		s.logger.WarnContext(ctx, "failed to publish agent event", "error", err)
	}
}
*/
