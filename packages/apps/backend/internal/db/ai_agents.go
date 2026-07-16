package db

import (
	"context"
	"encoding/json"
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
	db            rez.Database
	msgs          rez.MessageService
	notifications rez.DatabaseNotificationService

	statusSubs   map[uuid.UUID]*snapshotStatusSub
	statusSubsMu sync.Mutex
}

type snapshotStatusSub struct {
	tenantID    int
	lastStatus  aars.Status
	lastUpdated time.Time
	chans       []chan aix.SnapshotStatus
}

const aiAgentSnapshotStatusChannel = "rezible_ai_agent_snapshot_status"

func NewAiAgentSnapshotService(db rez.Database, msgs rez.MessageService, notifications rez.DatabaseNotificationService) (*AiAgentSnapshotService, error) {
	s := &AiAgentSnapshotService{
		db:            db,
		msgs:          msgs,
		notifications: notifications,
		statusSubs:    make(map[uuid.UUID]*snapshotStatusSub),
	}
	return s, nil
}

func (s *AiAgentSnapshotService) GetLatestSnapshotForRun(ctx context.Context, runId uuid.UUID) (*ent.AiAgentRunSnapshot, error) {
	query := s.db.Client(ctx).AiAgentRunSnapshot.Query().
		Where(aars.AiAgentRunID(runId)).
		Order(aars.ByCreatedAt(sql.OrderDesc()), aars.ByID(sql.OrderDesc())).
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
	if id == uuid.Nil {
		id = uuid.New()
	}

	var snapshot *ent.AiAgentRunSnapshot
	var changeEvent *rez.EventOnAiAgentRunSnapshotChange
	updateTxFn := func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "ai_agent_snapshot", id.String()); lockErr != nil {
			return fmt.Errorf("lock snapshot: %w", lockErr)
		}

		var existing *ent.AiAgentRunSnapshot
		var getErr error
		if existing, getErr = tx.AiAgentRunSnapshot.Get(ctx, id); getErr != nil && !ent.IsNotFound(getErr) {
			return fmt.Errorf("failed to lookup existing (%s): %w", id, getErr)
		}

		var mutator ent.EntityMutator[*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation]
		if existing != nil {
			mutator = existing.Update()
		} else {
			mutator = tx.AiAgentRunSnapshot.Create()
		}

		m := mutator.Mutation()
		if existing == nil {
			m.SetID(id)
		}

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

		if existing == nil {
			if _, ok := m.AiAgentRunID(); !ok {
				return fmt.Errorf("snapshot session ID is required")
			}
		} else {
			m.SetAiAgentRunID(existing.AiAgentRunID)
		}

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("failed to save: %w", saveErr)
		}

		run, runErr := tx.AiAgentRun.Get(ctx, updated.AiAgentRunID)
		if runErr != nil {
			return fmt.Errorf("query agent run: %w", runErr)
		}
		changeEvent = &rez.EventOnAiAgentRunSnapshotChange{
			AgentName:          run.AgentName,
			AgentRunMetadata:   run.Metadata,
			AgentRunId:         run.ID,
			AgentRunSnapshotId: updated.ID,
			Delta:              *delta,
		}

		snapshot = updated.Unwrap()
		return nil
	}

	if txErr := s.db.WithTx(ctx, updateTxFn); txErr != nil {
		return nil, fmt.Errorf("tx error: %w", txErr)
	}

	if changeEvent != nil {
		if eventErr := s.msgs.PublishEvent(ctx, *changeEvent); eventErr != nil {
			slog.ErrorContext(ctx, "failed to publish ai agent snapshot change event",
				"snapshotId", changeEvent.AgentRunSnapshotId.String(),
				"error", eventErr,
			)
		}
	}

	return snapshot, nil
}

func (s *AiAgentSnapshotService) Start(ctx context.Context) error {
	return s.notifications.Listen(ctx, "rezible_ai_agent_snapshot_status", s.refreshSubscribedStatuses, s.handleSnapshotStatusNotification)
}

type aiAgentSnapshotStatusNotification struct {
	TenantID   int         `json:"tenant_id"`
	SnapshotID uuid.UUID   `json:"snapshot_id"`
	Status     aars.Status `json:"status"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (s *AiAgentSnapshotService) refreshSubscribedStatuses(ctx context.Context) error {
	tenantIds := make(map[int][]uuid.UUID)
	s.statusSubsMu.Lock()
	for id, sub := range s.statusSubs {
		tenantIds[sub.tenantID] = append(tenantIds[sub.tenantID], id)
	}
	s.statusSubsMu.Unlock()

	for tenantID, ids := range tenantIds {
		tenantCtx := execution.NewTenantContext(ctx, tenantID)

		var statuses []struct {
			ID        uuid.UUID   `json:"id"`
			Status    aars.Status `json:"status"`
			UpdatedAt time.Time   `json:"updated_at"`
		}
		queryStatus := s.db.Client(tenantCtx).AiAgentRunSnapshot.Query().
			Where(aars.IDIn(ids...)).
			Select(aars.FieldID, aars.FieldStatus, aars.FieldUpdatedAt)

		if queryErr := queryStatus.Scan(tenantCtx, &statuses); queryErr != nil {
			return fmt.Errorf("query ai agent snapshot statuses: %w", queryErr)
		}
		for _, status := range statuses {
			s.notifyStatusUpdate(aiAgentSnapshotStatusNotification{
				TenantID:   tenantID,
				SnapshotID: status.ID,
				Status:     status.Status,
				UpdatedAt:  status.UpdatedAt,
			})
		}
	}

	return nil
}

func (s *AiAgentSnapshotService) handleSnapshotStatusNotification(ctx context.Context, payload []byte) error {
	var notification aiAgentSnapshotStatusNotification
	if jsonErr := json.Unmarshal(payload, &notification); jsonErr != nil {
		slog.WarnContext(ctx, "failed to decode ai agent snapshot status notification", "error", jsonErr)
		return nil
	}
	s.notifyStatusUpdate(notification)
	return nil
}

func (s *AiAgentSnapshotService) notifyStatusUpdate(notif aiAgentSnapshotStatusNotification) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()

	sub := s.statusSubs[notif.SnapshotID]
	if sub == nil || notif.UpdatedAt.Before(sub.lastUpdated) {
		return
	}
	if sub.lastStatus == notif.Status {
		if notif.UpdatedAt.After(sub.lastUpdated) {
			sub.lastUpdated = notif.UpdatedAt
		}
		return
	}

	sub.lastStatus = notif.Status
	sub.lastUpdated = notif.UpdatedAt
	for _, ch := range sub.chans {
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- aix.SnapshotStatus(notif.Status):
		default:
		}
	}
}

func (s *AiAgentSnapshotService) OnSnapshotStatusChange(ctx context.Context, id uuid.UUID) <-chan aix.SnapshotStatus {
	ch := make(chan aix.SnapshotStatus, 1)

	snap, snapErr := s.GetAgentRunSnapshot(ctx, id)
	if snapErr != nil || snap == nil {
		close(ch)
		return ch
	}

	s.addStatusSub(id, ch, snap)

	context.AfterFunc(ctx, func() {
		s.removeStatusSub(id, ch)
	})

	return ch
}

func (s *AiAgentSnapshotService) addStatusSub(id uuid.UUID, ch chan aix.SnapshotStatus, snap *ent.AiAgentRunSnapshot) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()

	sub := s.statusSubs[id]
	status := snap.Status
	if sub == nil {
		sub = &snapshotStatusSub{
			tenantID:    snap.TenantID,
			lastStatus:  snap.Status,
			lastUpdated: snap.UpdatedAt,
		}
		s.statusSubs[id] = sub
	} else {
		status = sub.lastStatus
	}
	ch <- aix.SnapshotStatus(status)
	sub.chans = append(sub.chans, ch)
}

func (s *AiAgentSnapshotService) removeStatusSub(id uuid.UUID, ch chan aix.SnapshotStatus) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()
	if sub, subExists := s.statusSubs[id]; subExists && sub != nil {
		if i := slices.Index(sub.chans, ch); i >= 0 {
			sub.chans = slices.Delete(sub.chans, i, i+1)
			if len(sub.chans) == 0 {
				delete(s.statusSubs, id)
			}
			close(ch)
		}
	}
}
