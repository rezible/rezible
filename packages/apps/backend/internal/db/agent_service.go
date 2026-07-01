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
	jobs.RegisterWorkerFunc(s.handleRunAgent)
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
	if agent, agentOk := s.agents.Get(params.Workflow); !agentOk {
		return nil, fmt.Errorf("agent not found for workflow %q", params.Workflow)
	} else if validationErr := agent.ValidateInput(params.Input); validationErr != nil {
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

	triggerKind := agentrun.TriggerKindManual
	if tk := agentrun.TriggerKind(params.TriggerKind); agentrun.TriggerKindValidator(tk) == nil {
		triggerKind = tk
	}

	var created *ent.AgentRun
	return created, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		create := tx.AgentRun.Create().
			SetOwnerUserID(ownerID).
			SetWorkflow(params.Workflow).
			SetInput(params.Input).
			SetTriggerKind(triggerKind).
			SetTriggerMetadata(params.TriggerPayload)
		task, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		created = task.Unwrap()

		_, jobErr := s.jobs.Insert(ctx, jobs.RunAgent{AgentRunID: created.ID}, runAgentWorkflowJobOpts)
		if jobErr != nil {
			return fmt.Errorf("enqueue agent workflow: %w", jobErr)
		}

		return nil
	})
}

func (s *AgentService) handleRunAgent(ctx context.Context, args jobs.RunAgent) error {
	queryRun := s.db.Client(ctx).AgentRun.Query().
		Where(agentrun.ID(args.AgentRunID)).
		WithResult()
	run, queryErr := queryRun.Only(ctx)
	if queryErr != nil {
		return fmt.Errorf("get agent run: %w", queryErr)
	}
	ctx = execution.NewAgentContext(ctx, run)

	agent, agentOk := s.agents.Get(run.Workflow)
	if !agentOk {
		return fmt.Errorf("agent not found for workflow '%s'", run.Workflow)
	}

	snapshotId, runErr := agent.Run(ctx, run, nil)
	if runErr != nil {
		return fmt.Errorf("run agent: %w", runErr)
	}
	slog.InfoContext(ctx, "invoked agent",
		"workflow", run.Workflow,
		"snapshot", snapshotId)

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
