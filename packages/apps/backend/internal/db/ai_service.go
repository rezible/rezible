package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aar "github.com/rezible/rezible/ent/aiagentrun"
	aarr "github.com/rezible/rezible/ent/aiagentrunresult"
	ars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	"github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type AiSessionService struct {
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
	msgs   rez.MessageService
	ai     rez.AiService
}

func NewAiSessionService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, aiSvc rez.AiService) (*AiSessionService, error) {
	s := &AiSessionService{
		logger: tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:     db,
		jobs:   jobSvc,
		msgs:   msgSvc,
		ai:     aiSvc,
	}
	jobs.RegisterWorkerFunc(s.handleStartAgentRun)
	jobs.RegisterWorkerFunc(s.handleContinueAgentRun)
	return s, nil
}

func (s *AiSessionService) GetAgentRun(ctx context.Context, id uuid.UUID) (*ent.AiAgentRun, error) {
	return s.db.Client(ctx).AiAgentRun.Query().
		Where(aar.ID(id)).
		Only(ctx)
}

func (s *AiSessionService) ListAgentRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AiAgentRun], error) {
	query := s.db.Client(ctx).AiAgentRun.Query().
		Order(aar.ByCreatedAt(sql.OrderDesc()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.AiAgentRun, *ent.AiAgentRunQuery](ctx, query, params.ListParams)
}

func (s *AiSessionService) SetRun(ctx context.Context, id uuid.UUID, setFn func(*ent.AiAgentRunMutation)) (*ent.AiAgentRun, error) {
	var run *ent.AiAgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.AiAgentRun, *ent.AiAgentRunMutation]
		if id == uuid.Nil {
			mutator = tx.AiAgentRun.Create()
		} else {
			mutator = tx.AiAgentRun.UpdateOneID(id)
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

func (s *AiSessionService) GetRunResult(ctx context.Context, runID uuid.UUID) (*ent.AiAgentRunResult, error) {
	return s.db.Client(ctx).AiAgentRunResult.Query().
		Where(aarr.AiAgentRunID(runID)).
		Only(ctx)
}

func (s *AiSessionService) CreateAgentRun(ctx context.Context, params rez.CreateAgentRunParams) (*ent.AiAgentRun, error) {
	jsonInput, inputErr := ai.ValidateAndEncodeAgentInput(params.AgentName, params.Input)
	if inputErr != nil {
		return nil, fmt.Errorf("invalid input: %w", inputErr)
	}

	ownerID := params.OwnerUserID
	if ownerID == uuid.Nil {
		userID, userOK := execution.GetContext(ctx).UserID()
		if !userOK {
			return nil, fmt.Errorf("missing task owner user")
		}
		ownerID = userID
	}

	var run *ent.AiAgentRun
	return run, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		create := tx.AiAgentRun.Create().
			SetOwnerUserID(ownerID).
			SetAgentName(params.AgentName).
			SetScopes(params.PermissionScopes).
			SetInput(jsonInput).
			SetMetadata(params.Metadata)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		run = created.Unwrap()

		startJobOpts := &river.InsertOpts{
			UniqueOpts: river.UniqueOpts{
				ByArgs:  true,
				ByState: jobs.UniqueStateNonCompleted,
			},
		}
		_, jobErr := s.jobs.Insert(ctx, jobs.StartAgentRun{AgentRunID: created.ID}, startJobOpts)
		if jobErr != nil {
			return fmt.Errorf("insert start agent run job: %w", jobErr)
		}

		return nil
	})
}

func (s *AiSessionService) getAndStartAgentRun(ctx context.Context, id uuid.UUID) (*ent.AiAgentRun, rez.AiAgentRunInvoker, error) {
	var run *ent.AiAgentRun
	var agent rez.AiAgentRunInvoker
	return run, agent, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		getStartedRun := tx.AiAgentRun.UpdateOneID(id).
			Where(aar.ID(id), aar.StartedAtIsNil()).
			SetStartedAt(time.Now().UTC())
		txRun, queryErr := getStartedRun.Save(ctx)
		if queryErr != nil {
			if ent.IsNotFound(queryErr) {
				return nil
			}
			return fmt.Errorf("get agent run: %w", queryErr)
		}
		run = txRun.Unwrap()

		var agentErr error
		agent, agentErr = s.ai.GetAgentRunInvoker(txRun)
		if agentErr != nil {
			return fmt.Errorf("get agent run invoker: %w", agentErr)
		}

		return nil
	})
}

func (s *AiSessionService) handleStartAgentRun(ctx context.Context, args jobs.StartAgentRun) error {
	run, agent, runErr := s.getAndStartAgentRun(ctx, args.AgentRunID)
	if run == nil || runErr != nil {
		return runErr
	}

	ctx = execution.NewAiAgentRunContext(ctx, run)

	snapshotId, startErr := agent.Start(ctx)
	if startErr != nil {
		return fmt.Errorf("start agent run: %w", startErr)
	}

	slog.InfoContext(ctx, "started agent run",
		"name", run.AgentName,
		"snapshot", snapshotId.String())

	return nil
}

func (s *AiSessionService) getAndResumeAgentRun(ctx context.Context, id uuid.UUID) (*ent.AiAgentRun, rez.AiAgentRunInvoker, error) {
	var run *ent.AiAgentRun
	var agent rez.AiAgentRunInvoker
	return run, agent, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryRun := tx.AiAgentRun.Query().
			Where(aar.ID(id)).
			WithSnapshots(func(q *ent.AiAgentRunSnapshotQuery) {
				q.Order(ars.ByCreatedAt()).Limit(1)
			})
		txRun, queryErr := queryRun.Only(ctx)
		if queryErr != nil {
			if ent.IsNotFound(queryErr) {
				return nil
			}
			return fmt.Errorf("get agent run: %w", queryErr)
		}
		run = txRun.Unwrap()

		var agentErr error
		agent, agentErr = s.ai.GetAgentRunInvoker(txRun)
		if agentErr != nil {
			return fmt.Errorf("get agent run invoker: %w", agentErr)
		}

		return nil
	})
}

func (s *AiSessionService) handleContinueAgentRun(ctx context.Context, args jobs.ContinueAgentRun) error {
	if args.Message == nil && args.Resume == nil {
		return fmt.Errorf("invalid args")
	}
	run, agent, runErr := s.getAndResumeAgentRun(ctx, args.AgentRunID)
	if run == nil || runErr != nil {
		return runErr
	}
	ctx = execution.NewAiAgentRunContext(ctx, run)

	var snapshotId uuid.UUID
	var sendErr error
	if args.Message != nil {
		snapshotId, sendErr = agent.SendMessage(ctx, rez.SendAgentRunMessageParams{
			ParentSnapshotID: args.ParentSnapshotID,
			Message:          args.Message,
		})
	} else if args.Resume != nil {
		snapshotId, sendErr = agent.Resume(ctx, rez.ResumeAgentRunParams{
			ParentSnapshotID: args.ParentSnapshotID,
			Respond:          args.Resume.Respond,
			Restart:          args.Resume.Restart,
		})
	}
	if sendErr != nil {
		return fmt.Errorf("continue agent run: %w", sendErr)
	}

	slog.InfoContext(ctx, "continued agent run",
		"name", run.AgentName,
		"snapshot", snapshotId.String())

	return nil
}
