package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aar "github.com/rezible/rezible/ent/aiagentrun"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
	"github.com/sourcegraph/conc/pool"
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

func (s *AiAgentService) lookupAgentRunInvoker(ctx context.Context, id uuid.UUID, parentSnapshotId *uuid.UUID) (*ent.AiAgentRun, rez.AiAgentRunInvoker, error) {
	var run *ent.AiAgentRun
	var agent rez.AiAgentRunInvoker
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
		if agent, agentErr = s.ai.GetAgentRunInvoker(run); agentErr != nil {
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

type AiAgentSnapshotService struct {
	db   rez.Database
	msgs rez.MessageService

	shutdownFn   func() error
	statusSubs   map[uuid.UUID][]chan aix.SnapshotStatus
	statusSubsMu sync.RWMutex
}

func NewAiAgentSnapshotService(db rez.Database, msgs rez.MessageService) (*AiAgentSnapshotService, error) {
	s := &AiAgentSnapshotService{
		db:         db,
		msgs:       msgs,
		shutdownFn: func() error { return nil },
		statusSubs: make(map[uuid.UUID][]chan aix.SnapshotStatus),
	}
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

func (s *AiAgentSnapshotService) UpdateAgentRunSnapshot(ctx context.Context, id uuid.UUID, setFn rez.AiAgentSnapshotSetFunc) (*ent.AiAgentRunSnapshot, error) {
	var snapshot *ent.AiAgentRunSnapshot
	return snapshot, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var existing *ent.AiAgentRunSnapshot
		if id != uuid.Nil {
			var getErr error
			if existing, getErr = tx.AiAgentRunSnapshot.Get(ctx, id); getErr != nil && !ent.IsNotFound(getErr) {
				return fmt.Errorf("failed to lookup existing (%s): %w", id, getErr)
			}
		}

		var mutator ent.EntityMutator[*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation]
		if existing != nil {
			mutator = existing.Update()
		} else {
			mutator = tx.AiAgentRunSnapshot.Create()
		}

		m := mutator.Mutation()
		delta, setErr := setFn(existing, m)
		if setErr != nil {
			return fmt.Errorf("update snapshot: %w", setErr)
		}

		if delta == nil {
			if existing != nil {
				snapshot = existing.Unwrap()
			}
			return nil
		}

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("failed to save: %w", saveErr)
		}

		run, runErr := tx.AiAgentRun.Get(ctx, updated.AiAgentRunID)
		if runErr != nil {
			return fmt.Errorf("query agent run: %w", runErr)
		}
		event := rez.EventOnAiAgentRunSnapshotChange{
			AgentName:          run.AgentName,
			AgentRunMetadata:   run.Metadata,
			AgentRunId:         run.ID,
			AgentRunSnapshotId: updated.ID,
			Delta:              *delta,
		}
		if eventErr := s.msgs.PublishEvent(ctx, event); eventErr != nil {
			return fmt.Errorf("failed to publish event: %w", eventErr)
		}

		snapshot = updated.Unwrap()
		return nil
	})
}

func (s *AiAgentSnapshotService) Start(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)

	p := pool.New().WithErrors().WithContext(cancelCtx)
	s.shutdownFn = func() error {
		cancel()
		if p != nil {
			if poolErr := p.Wait(); poolErr != nil && !errors.Is(poolErr, context.Canceled) {
				return fmt.Errorf("ai agent snapshot service shutdown: %w", poolErr)
			}
		}
		return nil
	}
	p.Go(s.startSnapshotStatusPoll)

	return nil
}

func (s *AiAgentSnapshotService) startSnapshotStatusPoll(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	statuses := make(map[uuid.UUID]aars.Status)
	for {
		select {
		case <-ticker.C:
			statuses = s.pollForStatusUpdates(ctx, statuses)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *AiAgentSnapshotService) pollForStatusUpdates(ctx context.Context, prevStatuses map[uuid.UUID]aars.Status) map[uuid.UUID]aars.Status {
	slog.Debug("ai agent snapshot service polling for status updates")
	if len(s.statusSubs) == 0 {
		return prevStatuses
	}

	ids := make([]uuid.UUID, 0, len(s.statusSubs))
	for id := range s.statusSubs {
		ids = append(ids, id)
	}

	queryStatuses := s.db.Client(ctx).AiAgentRunSnapshot.Query().Where(aars.IDIn(ids...))

	var statuses []struct {
		ID     uuid.UUID   `json:"id"`
		Status aars.Status `json:"status"`
	}
	if queryErr := queryStatuses.Select(aars.FieldID, aars.FieldStatus).Scan(ctx, &statuses); queryErr != nil {
		slog.Error("failed to query ai agent snapshot statuses", "err", queryErr)
	}
	newStatuses := make(map[uuid.UUID]aars.Status, len(statuses))
	for _, status := range statuses {
		newStatuses[status.ID] = status.Status
	}

	for id, newStatus := range newStatuses {
		if oldStatus, existed := prevStatuses[id]; existed && oldStatus != newStatus {
			s.notifyStatusUpdate(id, newStatus)
		}
	}

	return newStatuses
}

func (s *AiAgentSnapshotService) notifyStatusUpdate(id uuid.UUID, status aars.Status) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()
	for _, ch := range s.statusSubs[id] {
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- aix.SnapshotStatus(status):
		default:
		}
	}
}

func (s *AiAgentSnapshotService) Shutdown() error {
	slog.Info("Stopping ai agent status poll")
	return s.shutdownFn()
}

func (s *AiAgentSnapshotService) OnSnapshotStatusChange(ctx context.Context, id uuid.UUID) <-chan aix.SnapshotStatus {
	ch := make(chan aix.SnapshotStatus, 1)

	s.statusSubsMu.Lock()
	snap, snapErr := s.GetAgentRunSnapshot(ctx, id)
	if snapErr != nil {
		s.statusSubsMu.Unlock()
		close(ch)
		return ch
	}
	ch <- aix.SnapshotStatus(snap.Status)
	s.statusSubs[id] = append(s.statusSubs[id], ch)
	s.statusSubsMu.Unlock()

	context.AfterFunc(ctx, func() {
		s.removeStatusSub(id, ch)
	})

	return ch
}

func (s *AiAgentSnapshotService) removeStatusSub(id uuid.UUID, ch chan aix.SnapshotStatus) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()
	subs := s.statusSubs[id]
	i := slices.Index(subs, ch)
	if i < 0 {
		return
	}
	subs = slices.Delete(subs, i, i+1)
	if len(subs) == 0 {
		delete(s.statusSubs, id)
	} else {
		s.statusSubs[id] = subs
	}
	close(ch)
}
