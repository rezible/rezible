package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aar "github.com/rezible/rezible/ent/aiagentrun"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type AiAgentService struct {
	rez.AiAgentSnapshotService
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
	msgs   rez.MessageService
	ai     rez.AiService
}

func NewAiAgentService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgSvc rez.MessageService, sessions rez.AiAgentSnapshotService, aiSvc rez.AiService) (*AiAgentService, error) {
	s := &AiAgentService{
		AiAgentSnapshotService: sessions,
		logger:                 tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_service"}),
		db:                     db,
		jobs:                   jobSvc,
		msgs:                   msgSvc,
		ai:                     aiSvc,
	}
	jobs.RegisterWorkerFunc(s.handleInvokeAgentRun)
	return s, nil
}

func (s *AiAgentService) GetAgentRun(ctx context.Context, id uuid.UUID) (*ent.AiAgentRun, error) {
	return s.db.Client(ctx).AiAgentRun.Query().
		Where(aar.ID(id)).
		Only(ctx)
}

func (s *AiAgentService) ListAgentRuns(ctx context.Context, params rez.ListAgentRunsParams) (*ent.ListResult[ent.AiAgentRun], error) {
	query := s.db.Client(ctx).AiAgentRun.Query().
		Order(aar.ByCreatedAt(sql.OrderDesc()))
	predicates := params.Predicates
	for key, val := range params.Metadata {
		predicates = append(predicates, predicate.AiAgentRun(func(s *sql.Selector) {
			s.Where(sqljson.ValueEQ(aar.FieldMetadata, val, sqljson.DotPath(key)))
		}))
	}
	if len(predicates) > 0 {
		query.Where(predicates...)
	}
	return ent.DoListQuery[ent.AiAgentRun, *ent.AiAgentRunQuery](ctx, query, params.ListParams)
}

func (s *AiAgentService) SetRun(ctx context.Context, id uuid.UUID, setFn func(*ent.AiAgentRunMutation)) (*ent.AiAgentRun, error) {
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

func (s *AiAgentService) CreateAgentRun(ctx context.Context, name string, params rez.CreateAgentRunParams) (*ent.AiAgentRun, error) {
	jsonInput, inputOk := params.Input.([]byte)
	if !inputOk {
		var jsonErr error
		if jsonInput, jsonErr = json.Marshal(params.Input); jsonErr != nil {
			return nil, jsonErr
		}
	}

	inputErr := s.ai.ValidateAgentRunInput(name, jsonInput)
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
			SetAgentName(name).
			SetScopes(params.PermissionScopes).
			SetInput(jsonInput).
			SetMetadata(params.Metadata)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent task: %w", createErr)
		}
		run = created.Unwrap()

		if invokeErr := s.InvokeAgentRun(ctx, run.ID, rez.InvokeAgentRunParams{}); invokeErr != nil {
			return fmt.Errorf("invoke run: %w", invokeErr)
		}

		return nil
	})
}

func (s *AiAgentService) InvokeAgentRun(ctx context.Context, id uuid.UUID, params rez.InvokeAgentRunParams) error {
	args := jobs.InvokeAgentRun{
		AgentRunID: id,
		Message:    params.Message,
		Resume:     params.Resume,
	}
	if params.ParentSnapshotID != uuid.Nil {
		args.ParentSnapshotID = &params.ParentSnapshotID
	}
	jobOpts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
	_, jobErr := s.jobs.Insert(ctx, args, jobOpts)
	if jobErr != nil {
		return fmt.Errorf("insert start agent run job: %w", jobErr)
	}
	return nil
}

/*
func (s *AiAgentService) getAndStartAgentRun(ctx context.Context, id uuid.UUID) (*ent.AiAgentRun, rez.AiAgentInvoker, error) {
	var run *ent.AiAgentRun
	var agent rez.AiAgentInvoker
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
		agent, agentErr = s.ai.GetAgentRunner(txRun)
		if agentErr != nil {
			return fmt.Errorf("get agent run invoker: %w", agentErr)
		}

		return nil
	})
}

func (s *AiAgentService) handleStartAgentRun(ctx context.Context, args jobs.StartAgentRun) error {
	run, agent, runErr := s.getAndStartAgentRun(ctx, args.AgentRunID)
	if run == nil || runErr != nil {
		return runErr
	}

	ctx = execution.NewAiAgentRunContext(ctx, run)

	snapshotId, startErr := agent.Start(ctx)
	if startErr != nil {
		return fmt.Errorf("start agent run: %w", startErr)
	}

	//event := rez.EventOnAiAgentRunSnapshot{
	//	AgentName:       run.AgentName,
	//	AgentRunId:      run.ID,
	//	AgentSnapshotId: snapshotId,
	//	RunMetadata:     run.Metadata,
	//}
	//if eventErr := s.msgs.PublishEvent(ctx, event); eventErr != nil {
	//	slog.Error("failed to publish agent run finished event",
	//		"error", eventErr.Error(),
	//	)
	//}

	return nil
}
*/

func (s *AiAgentService) LookupAgentRunsByMetadata(ctx context.Context, metadata map[string]any) (ent.AiAgentRuns, error) {
	query := s.db.Client(ctx).AiAgentRun.Query()
	for key, val := range metadata {
		query = query.Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueEQ(aar.FieldMetadata, val, sqljson.DotPath(key)))
		})
	}
	return query.All(ctx)
}

func (s *AiAgentService) lookupAgentRunInvoker(ctx context.Context, id uuid.UUID, parentSnapshotId *uuid.UUID) (*ent.AiAgentRun, rez.AiAgentInvoker, error) {
	var run *ent.AiAgentRun
	var agent rez.AiAgentInvoker
	return run, agent, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var txRun *ent.AiAgentRun
		var runErr error
		if parentSnapshotId != nil {
			txRun, runErr = tx.AiAgentRun.Query().
				Where(aar.And(
					aar.ID(id),
					aar.HasSnapshotsWith(aars.ID(*parentSnapshotId)),
					aar.StartedAtNotNil(),
				)).Only(ctx)
		} else {
			txRun, runErr = tx.AiAgentRun.UpdateOneID(id).
				SetStartedAt(time.Now()).
				Save(ctx)
		}
		if runErr != nil {
			return fmt.Errorf("get agent run: %w", runErr)
		}
		run = txRun.Unwrap()

		var agentErr error
		if agent, agentErr = s.ai.GetAgentRunner(run); agentErr != nil {
			return fmt.Errorf("get agent runner: %w", agentErr)
		}

		return nil
	})
}

func (s *AiAgentService) handleInvokeAgentRun(ctx context.Context, args jobs.InvokeAgentRun) error {
	run, agent, runErr := s.lookupAgentRunInvoker(ctx, args.AgentRunID, args.ParentSnapshotID)
	if run == nil || runErr != nil {
		return runErr
	}

	snapshotId, sendErr := agent.Invoke(ctx, args.ParentSnapshotID, args.Message, args.Resume)
	if sendErr != nil {
		return fmt.Errorf("continue agent run: %w", sendErr)
	}

	slog.InfoContext(ctx, "continued agent run",
		"name", run.AgentName,
		"snapshot", snapshotId.String())

	return nil
}

func (s *AiAgentService) GetAgentRunOutput(ctx context.Context, id uuid.UUID) (*ent.AiAgentRunOutput, error) {
	return s.db.Client(ctx).AiAgentRunOutput.Get(ctx, id)
}

func (s *AiAgentService) ClaimAgentRunOutput(ctx context.Context, id uuid.UUID, fn func(context.Context, []byte) (map[string]any, error)) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// claim transactional lock?
		output, queryErr := tx.AiAgentRunOutput.Get(ctx, id)
		if queryErr != nil {
			return fmt.Errorf("get agent run output: %w", queryErr)
		}
		md, mdErr := fn(ctx, output.Data)
		if mdErr != nil {
			return fmt.Errorf("get agent run output metadata: %w", mdErr)
		}
		return output.Update().SetMetadata(md).Exec(ctx)
	})
}

type AiAgentSnapshotService struct {
	db   rez.Database
	msgs rez.MessageService
}

func NewAiAgentSnapshotService(db rez.Database, msgs rez.MessageService) (*AiAgentSnapshotService, error) {
	s := &AiAgentSnapshotService{db: db, msgs: msgs}
	return s, nil
}

func (s *AiAgentSnapshotService) GetLatestSnapshotForRun(ctx context.Context, runId uuid.UUID) (*ent.AiAgentRunSnapshot, error) {
	query := s.db.Client(ctx).AiAgentRunSnapshot.Query().
		Where(aars.AiAgentRunID(runId)).
		Order(aars.ByCreatedAt(sql.OrderDesc())).
		Limit(1)
	res, resErr := query.First(ctx)
	if resErr != nil {
		if ent.IsNotFound(resErr) {
			return nil, nil
		}
		return nil, resErr
	}
	return res, nil
}

func (s *AiAgentSnapshotService) GetAgentRunSnapshot(ctx context.Context, id uuid.UUID) (*ent.AiAgentRunSnapshot, error) {
	res, resErr := s.db.Client(ctx).AiAgentRunSnapshot.Get(ctx, id)
	if resErr != nil && !ent.IsNotFound(resErr) {
		return nil, resErr
	}
	return res, nil
}

func (s *AiAgentSnapshotService) SetAgentRunSnapshot(ctx context.Context, id uuid.UUID, setFn rez.AiAgentSnapshotSetFunc) (*ent.AiAgentRunSnapshot, error) {
	var snapshot *ent.AiAgentRunSnapshot
	return snapshot, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var curr *ent.AiAgentRunSnapshot
		var mutator ent.EntityMutator[*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation]
		if id != uuid.Nil {
			var getErr error
			if curr, getErr = tx.AiAgentRunSnapshot.Get(ctx, id); getErr != nil && !ent.IsNotFound(getErr) {
				return fmt.Errorf("failed to lookup existing (%s): %w", id, getErr)
			}
		}
		if curr != nil {
			mutator = curr.Update()
		} else {
			mutator = tx.AiAgentRunSnapshot.Create()
		}

		m := mutator.Mutation()
		changedOutputs, setErr := setFn(curr, m)
		if setErr != nil {
			return fmt.Errorf("update snapshot: %w", setErr)
		}

		if len(m.Fields()) == 0 {
			if curr == nil {
				return fmt.Errorf("no fields changed, no existing snapshot")
			}
			snapshot = curr.Unwrap()
			return nil
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("failed to save: %w", saveErr)
		}

		if len(changedOutputs) > 0 {
			run, runErr := tx.AiAgentRun.Get(ctx, saved.AiAgentRunID)
			if runErr != nil {
				return fmt.Errorf("query run for outputs: %w", runErr)
			}
			event := rez.EventOnAiAgentRunOutput{
				AgentName:          run.AgentName,
				AgentRunMetadata:   run.Metadata,
				AgentRunSnapshotId: saved.ID,
				Parts:              changedOutputs,
			}
			if eventErr := s.msgs.PublishEvent(ctx, event); eventErr != nil {
				return fmt.Errorf("failed to publish event: %w", eventErr)
			}
		}

		snapshot = saved.Unwrap()
		return nil
	})
}
