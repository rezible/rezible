package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/agents"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunresult"
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
	jobs.RegisterWorkerFunc(s.handleStartAgentRun)
	return s, nil
}

func (s *AgentService) GetRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, error) {
	return s.db.Client(ctx).AgentRun.Query().
		Where(agentrun.ID(id)).
		Only(ctx)
}

func (s *AgentService) ListRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AgentRun], error) {
	query := s.db.Client(ctx).AgentRun.Query().
		Order(agentrun.ByCreatedAt(sql.OrderDesc()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.AgentRun, *ent.AgentRunQuery](ctx, query, params.ListParams)
}

func (s *AgentService) SetRun(ctx context.Context, id uuid.UUID, setFn func(*ent.AgentRunMutation)) (*ent.AgentRun, error) {
	var run *ent.AgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.AgentRun, *ent.AgentRunMutation]
		if id == uuid.Nil {
			mutator = tx.AgentRun.Create()
		} else {
			mutator = tx.AgentRun.UpdateOneID(id)
		}
		setFn(mutator.Mutation())
		txRun, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}
		run = txRun.Unwrap()
		return nil
	})
}

func (s *AgentService) GetRunResult(ctx context.Context, runID uuid.UUID) (*ent.AgentRunResult, error) {
	return s.db.Client(ctx).AgentRunResult.Query().
		Where(agentrunresult.AgentRunID(runID)).
		Only(ctx)
}

var runAgentWorkflowJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (s *AgentService) CreateRun(ctx context.Context, params rez.CreateAgentRunParams) (*ent.AgentRun, error) {
	input, inputErr := json.Marshal(params.Input)
	if inputErr != nil {
		return nil, oapi.Error(ctx, "invalid input", inputErr)
	}
	if validationErr := agents.ValidateInput(params.Workflow, input); validationErr != nil {
		return nil, fmt.Errorf("input validation: %w", validationErr)
	}

	ownerID := params.OwnerUserID
	if ownerID == uuid.Nil {
		userID, userOK := execution.GetContext(ctx).UserID()
		if !userOK {
			return nil, fmt.Errorf("missing task owner user")
		}
		ownerID = userID
	}

	var run *ent.AgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		create := tx.AgentRun.Create().
			SetOwnerUserID(ownerID).
			SetWorkflow(params.Workflow).
			SetInput(input).
			SetMetadata(params.Metadata)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}

		_, jobErr := s.jobs.Insert(ctx, jobs.StartAgentRun{AgentRunID: created.ID}, runAgentWorkflowJobOpts)
		if jobErr != nil {
			return fmt.Errorf("enqueue agent workflow: %w", jobErr)
		}

		run = created.Unwrap()

		return nil
	})
}

func (s *AgentService) getAndStartAgentRun(ctx context.Context, id uuid.UUID) (*ent.AgentRun, rez.Agent, error) {
	var run *ent.AgentRun
	var agent rez.Agent
	return run, agent, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		getStartedRun := tx.AgentRun.UpdateOneID(id).
			Where(agentrun.ID(id), agentrun.StartedAtIsNil()).
			SetStartedAt(time.Now().UTC())
		txRun, queryErr := getStartedRun.Save(ctx)
		if queryErr != nil {
			if ent.IsNotFound(queryErr) {
				return nil
			}
			return fmt.Errorf("get agent run: %w", queryErr)
		}

		txAgent, agentOk := s.agents.Get(txRun.Workflow)
		if !agentOk {
			return fmt.Errorf("agent not found for workflow '%s'", txRun.Workflow)
		}

		run = txRun.Unwrap()
		agent = txAgent

		return nil
	})
}

func (s *AgentService) handleStartAgentRun(ctx context.Context, args jobs.StartAgentRun) error {
	run, agent, runErr := s.getAndStartAgentRun(ctx, args.AgentRunID)
	if run == nil || runErr != nil {
		return runErr
	}

	ctx = execution.NewAgentContext(ctx, run)
	snapshotId, startErr := agent.Start(ctx, run, nil)
	if startErr != nil {
		return fmt.Errorf("start agent run: %w", startErr)
	}

	slog.InfoContext(ctx, "started agent run",
		"workflow", run.Workflow,
		"snapshot", snapshotId.String())

	return nil
}

func (s *AgentService) handleContinueAgentRun(ctx context.Context, args jobs.ContinueAgentRun) error {
	run, runErr := s.GetRun(ctx, args.AgentRunID)
	if run == nil || runErr != nil {
		return fmt.Errorf("get agent run: %w", runErr)
	}
	ctx = execution.NewAgentContext(ctx, run)

	agent, agentOk := s.agents.Get(run.Workflow)
	if !agentOk {
		return fmt.Errorf("agent not found for workflow '%s'", run.Workflow)
	}

	var snapshotId uuid.UUID
	var sendErr error
	if args.Message != nil {
		params := rez.SendAgentRunMessageParams{
			ParentSnapshotID: args.ParentSnapshotID,
			Message:          args.Message,
		}
		snapshotId, sendErr = agent.SendMessage(ctx, run, params)
	} else if args.Resume != nil {
		params := rez.ResumeAgentRunParams{
			ParentSnapshotID: args.ParentSnapshotID,
			Respond:          args.Resume.Respond,
			Restart:          args.Resume.Restart,
		}
		snapshotId, sendErr = agent.Resume(ctx, run, params)
	} else {
		return fmt.Errorf("invalid args")
	}
	if sendErr != nil {
		return fmt.Errorf("continue agent run: %w", sendErr)
	}

	slog.InfoContext(ctx, "continued agent run",
		"workflow", run.Workflow,
		"snapshot", snapshotId.String())

	return nil
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
