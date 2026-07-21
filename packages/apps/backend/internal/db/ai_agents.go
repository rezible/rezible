package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	as "github.com/rezible/rezible/ent/agentsession"
	at "github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

type AgentSessionService struct {
	logger     *slog.Logger
	db         rez.Database
	jobs       rez.JobService
	ai         rez.AiService
	turnWorker *agentTurnWorker
}

func NewAgentSessionService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, aiSvc rez.AiService, cfg rez.AiConfig) (*AgentSessionService, error) {
	s := &AgentSessionService{
		logger: tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_session_service"}),
		db:     db,
		jobs:   jobSvc,
		ai:     aiSvc,
		turnWorker: &agentTurnWorker{
			db:      db,
			ai:      aiSvc,
			logger:  tel.NewLogger(rez.NewLoggerOptions{PackageName: "agent_turn_worker"}),
			timeout: cfg.Agents.WorkerTimeout,
		},
	}

	jobs.RegisterWorker(s.turnWorker)

	return s, nil
}

func (s *AgentSessionService) GetAgentSession(ctx context.Context, id uuid.UUID) (*ent.AgentSession, error) {
	query := s.db.Client(ctx).AgentSession.Query().
		Where(as.ID(id))
	if userID, userOK := execution.GetContext(ctx).UserID(); userOK {
		query.Where(as.OwnerUserID(userID))
	}
	return query.Only(ctx)
}

func (s *AgentSessionService) ListAgentSessions(ctx context.Context, params rez.ListAgentSessionsParams) (*ent.ListResult[ent.AgentSession], error) {
	query := s.db.Client(ctx).AgentSession.Query().
		Order(as.ByCreatedAt(sql.OrderDesc()), as.ByID(sql.OrderDesc())).
		Where(params.Predicates...)

	if userID, userOK := execution.GetContext(ctx).UserID(); userOK {
		query.Where(as.OwnerUserID(userID))
	}

	for key, val := range params.Metadata {
		query.Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueEQ(as.FieldMetadata, val, sqljson.DotPath(key)))
		})
	}

	return ent.DoListQuery[ent.AgentSession, *ent.AgentSessionQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) insertInvokeAgentTurnJob(ctx context.Context, sessId uuid.UUID, turnId uuid.UUID) (int64, error) {
	result, insertErr := s.jobs.Insert(ctx, jobs.InvokeAgentTurn{AgentSessionID: sessId, AgentTurnID: turnId}, nil)
	if insertErr != nil {
		return 0, fmt.Errorf("insert turn job: %w", insertErr)
	}
	if result == nil || result.Job == nil {
		return 0, fmt.Errorf("no inserted job returned")
	}
	if result.UniqueSkippedAsDuplicate {
		return 0, fmt.Errorf("%w: river skipped duplicate agent turn job", rez.ErrConflict)
	}
	return result.Job.ID, nil
}

func (s *AgentSessionService) CreateAgentSession(ctx context.Context, params rez.CreateAgentSessionParams) (*ent.AgentSession, error) {
	name := strings.TrimSpace(params.AgentName)
	if name == "" {
		return nil, fmt.Errorf("%w: agent name is required", rez.ErrInvalidInput)
	}
	if params.OwnerUserID == uuid.Nil {
		return nil, fmt.Errorf("owner user ID is required")
	}

	initialInput, initialErr := s.ai.MakeInitialAgentTurnInput(ctx, name, params.Input)
	if initialErr != nil {
		return nil, fmt.Errorf("make initial turn input: %w", initialErr)
	}

	var session *ent.AgentSession
	return session, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		createSession := tx.AgentSession.Create().
			SetAgentName(name).
			SetOwnerUserID(params.OwnerUserID)
		if len(params.PermissionScopes) > 0 {
			createSession.SetDefaultScopes(params.PermissionScopes)
		}
		if params.Metadata != nil {
			createSession.SetMetadata(params.Metadata)
		}

		createdSession, createErr := createSession.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent session: %w", createErr)
		}

		turnParams := &rez.RequestAgentTurnParams{Input: initialInput}
		createdTurn, turnErr := s.createAndRequestTurn(ctx, createdSession.ID, turnParams)
		if turnErr != nil {
			return fmt.Errorf("create turn: %w", turnErr)
		}

		session = createdSession.Unwrap()
		session.Edges.Turns = ent.AgentTurns{createdTurn.Unwrap()}
		return nil
	})
}

func (s *AgentSessionService) queryAgentTurns(ctx context.Context) *ent.AgentTurnQuery {
	q := s.db.Client(ctx).AgentTurn.Query()
	if userID, isUserContext := execution.GetContext(ctx).UserID(); isUserContext {
		q.Where(at.HasAgentSessionWith(as.OwnerUserID(userID)))
	}
	return q
}

func (s *AgentSessionService) RequestAgentTurn(ctx context.Context, sessionID uuid.UUID, params *rez.RequestAgentTurnParams) (*ent.AgentTurn, error) {
	var turn *ent.AgentTurn
	return turn, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryInitialTurn := s.queryAgentTurns(ctx).
			Where(at.AgentSessionID(sessionID), at.ParentIDIsNil(), at.StatusEQ(at.StatusCompleted))

		initialTurnExists, queryInitialTurnErr := queryInitialTurn.Exist(ctx)
		if queryInitialTurnErr != nil {
			return fmt.Errorf("query completed initial turn: %w", queryInitialTurnErr)
		}
		if !initialTurnExists {
			return fmt.Errorf("%w: the initial turn must complete before requesting another turn", rez.ErrConflict)
		}

		createdTurn, turnErr := s.createAndRequestTurn(ctx, sessionID, params)
		if turnErr != nil {
			return fmt.Errorf("create turn: %w", turnErr)
		}
		turn = createdTurn.Unwrap()
		return nil
	})
}

func (s *AgentSessionService) createAndRequestTurn(ctx context.Context, sessionID uuid.UUID, params *rez.RequestAgentTurnParams) (*ent.AgentTurn, error) {
	normalized := *params.Input
	if normalized.Resume != nil &&
		len(normalized.Resume.Respond)+len(normalized.Resume.Restart) == 0 {
		normalized.Resume = nil
	}
	if normalized.Message == nil && normalized.Resume == nil {
		return nil, rez.ErrInvalidInput
	}

	encodedInput, encodeErr := json.Marshal(normalized)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode turn input: %w", encodeErr)
	}

	turnID := uuid.New()

	jobID, jobErr := s.insertInvokeAgentTurnJob(ctx, sessionID, turnID)
	if jobErr != nil {
		return nil, fmt.Errorf("enqueue agent turn job: %w", jobErr)
	}

	createTurn := s.db.Client(ctx).AgentTurn.Create().
		SetID(turnID).
		SetAgentSessionID(sessionID).
		SetNillableParentID(params.ParentTurnID).
		SetRiverJobID(jobID).
		SetInput(encodedInput).
		SetStatus(at.StatusQueued)

	return createTurn.Save(ctx)
}

func (s *AgentSessionService) GetAgentTurn(ctx context.Context, id uuid.UUID) (*ent.AgentTurn, error) {
	return s.queryAgentTurns(ctx).Where(at.ID(id)).Only(ctx)
}

func (s *AgentSessionService) ListAgentTurns(ctx context.Context, params rez.ListAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error) {
	query := s.queryAgentTurns(ctx).
		Where(params.Predicates...).
		Order(at.ByCreatedAt(sql.OrderDesc()), at.ByID(sql.OrderDesc()))
	return ent.DoListQuery[ent.AgentTurn, *ent.AgentTurnQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) GetLatestTurnForSession(ctx context.Context, sessionId uuid.UUID) (*ent.AgentTurn, error) {
	queryByYoungest := s.queryAgentTurns(ctx).
		Where(at.AgentSessionID(sessionId)).
		Order(at.ByCreatedAt(sql.OrderDesc()), at.ByID(sql.OrderDesc()))
	return queryByYoungest.First(ctx)
}

func (s *AgentSessionService) GetLastSuccessfulAgentTurn(ctx context.Context, sessionID uuid.UUID) (*ent.AgentTurn, error) {
	queryTurn := s.queryAgentTurns(ctx).
		Where(at.AgentSessionID(sessionID), at.StatusEQ(at.StatusCompleted),
			at.Not(at.HasChildrenWith(at.StatusEQ(at.StatusCompleted))))
	turn, turnErr := queryTurn.Only(ctx)
	if turnErr != nil && !ent.IsNotFound(turnErr) {
		return nil, turnErr
	}
	return turn, nil
}

func acquireAgentSessionTurnLock(ctx context.Context, db rez.Database, sessionId uuid.UUID) error {
	return db.AcquireTxLocks(ctx, "agent_session_turn", sessionId.String())
}

func (s *AgentSessionService) lookupAgentTurnSessionAndAcquireLock(ctx context.Context, turnID uuid.UUID) (*ent.AgentSession, error) {
	querySession := s.db.Client(ctx).AgentSession.Query().
		Where(as.HasTurnsWith(at.ID(turnID)))
	if userID, isUserContext := execution.GetContext(ctx).UserID(); isUserContext {
		querySession.Where(as.OwnerUserID(userID))
	}
	sess, sessErr := querySession.Only(ctx)
	if sessErr != nil {
		return nil, sessErr
	}
	if lockErr := acquireAgentSessionTurnLock(ctx, s.db, sess.ID); lockErr != nil {
		return nil, fmt.Errorf("lock agent session: %w", lockErr)
	}
	return sess, nil
}

func (s *AgentSessionService) AbortAgentTurn(ctx context.Context, turnID uuid.UUID) (*ent.AgentTurn, error) {
	var result *ent.AgentTurn
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		_, sessErr := s.lookupAgentTurnSessionAndAcquireLock(ctx, turnID)
		if sessErr != nil {
			return fmt.Errorf("agent session: %w", sessErr)
		}
		turn, turnErr := s.GetAgentTurn(ctx, turnID)
		if turnErr != nil {
			return turnErr
		}
		if turn.Status == at.StatusAborted {
			result = turn.Unwrap()
			return nil
		}
		if turn.Status != at.StatusQueued && turn.Status != at.StatusRunning {
			return fmt.Errorf("%w: only queued or running turns can be aborted", rez.ErrConflict)
		}
		if cancelErr := s.jobs.Cancel(ctx, turn.RiverJobID); cancelErr != nil {
			return fmt.Errorf("failed to cancel job: %w", cancelErr)
		}

		update := turn.Update().
			SetStatus(at.StatusAborted).
			SetFinishReason(string(aix.AgentFinishReasonAborted)).
			SetFinishedAt(time.Now().UTC())
		if turn.Status == at.StatusQueued {
			update.ClearState()
			update.ClearError()
		}
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("abort agent turn: %w", updateErr)
		}
		result = updated.Unwrap()
		return nil
	})
}

func (s *AgentSessionService) RetryAgentTurn(ctx context.Context, turnID uuid.UUID) (*ent.AgentTurn, error) {
	var result *ent.AgentTurn
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		sess, sessErr := s.lookupAgentTurnSessionAndAcquireLock(ctx, turnID)
		if sessErr != nil {
			return fmt.Errorf("agent session: %w", sessErr)
		}

		turn, turnErr := s.GetAgentTurn(ctx, turnID)
		if turnErr != nil {
			return turnErr
		}
		if turn.Status != at.StatusFailed {
			return fmt.Errorf("%w: only failed turns can be retried", rez.ErrConflict)
		}

		jobID, jobErr := s.insertInvokeAgentTurnJob(ctx, sess.ID, turn.ID)
		if jobErr != nil {
			return fmt.Errorf("enqueue agent turn retry: %w", jobErr)
		}

		updateTurn := turn.Update().ClearStateFields().
			SetRiverJobID(jobID).
			SetStatus(at.StatusQueued)

		updated, updateErr := updateTurn.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("retry agent turn: %w", updateErr)
		}
		result = updated.Unwrap()
		return nil
	})
}

type agentTurnWorker struct {
	river.WorkerDefaults[jobs.InvokeAgentTurn]

	db      rez.Database
	ai      rez.AiService
	logger  *slog.Logger
	timeout time.Duration
}

func (w *agentTurnWorker) Timeout(job *river.Job[jobs.InvokeAgentTurn]) time.Duration {
	return w.timeout
}

func (w *agentTurnWorker) Work(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) error {
	claim, claimErr := w.claimAgentTurn(ctx, job)
	if claimErr != nil {
		if ent.IsNotFound(claimErr) {
			return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
		}
		if errors.Is(claimErr, &river.JobCancelError{}) || errors.Is(claimErr, &river.JobSnoozeError{}) {
			return claimErr
		}
		if errors.Is(claimErr, rezai.ErrAgentInterrupted) {
			return w.saveInvocationError(ctx, job, claimErr)
		}
		if job.Attempt >= job.MaxAttempts {
			return w.saveInvocationError(ctx, job, claimErr)
		}
		return claimErr
	}
	if claim == nil {
		return nil
	}

	result, invokeErr := w.invokeClaimedTurn(ctx, claim)

	w.logger.InfoContext(ctx, "agent turn invoked",
		"sessionId", claim.session.ID,
		"turnId", claim.turn.ID)

	if invokeErr != nil {
		return w.saveInvocationError(ctx, job, invokeErr)
	}
	return w.saveInvocationResult(ctx, job, result)
}

func (w *agentTurnWorker) invokeClaimedTurn(ctx context.Context, claim *claimedAgentTurn) (*rez.AgentTurnResult, error) {
	var input rez.AgentTurnInput
	if decodeErr := json.Unmarshal(claim.turn.Input, &input); decodeErr != nil {
		return nil, fmt.Errorf("decode stored turn input: %w", decodeErr)
	}
	return w.ai.InvokeAgentTurn(ctx, claim.session, claim.turn, claim.parentState, &input)
}

type claimedAgentTurn struct {
	session     *ent.AgentSession
	turn        *ent.AgentTurn
	parentState []byte
}

func (w *agentTurnWorker) claimAgentTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) (*claimedAgentTurn, error) {
	var claim *claimedAgentTurn
	return claim, w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := acquireAgentSessionTurnLock(ctx, w.db, job.Args.AgentSessionID); lockErr != nil {
			return fmt.Errorf("acquire agent session lock: %w", lockErr)
		}
		sess, sessErr := tx.AgentSession.Get(ctx, job.Args.AgentSessionID)
		if sessErr != nil {
			return fmt.Errorf("lookup agent session: %w", sessErr)
		}

		turn, turnErr := sess.QueryTurns().Where(at.ID(job.Args.AgentTurnID)).Only(ctx)
		if turnErr != nil {
			if ent.IsNotFound(turnErr) {
				return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
			}
			return fmt.Errorf("reload agent turn: %w", turnErr)
		}

		if turn.RiverJobID != job.ID {
			return river.JobCancel(fmt.Errorf("stale agent turn job"))
		}

		if turn.Status != at.StatusQueued {
			switch turn.Status {
			case at.StatusCompleted:
				return nil
			case at.StatusFailed, at.StatusAborted:
				return river.JobCancel(fmt.Errorf("agent turn is already %s", turn.Status))
			case at.StatusRunning:
				return rezai.ErrAgentInterrupted
			}
			return fmt.Errorf("invalid agent turn status %q", turn.Status)
		}

		isQueuedAndOlder := at.And(
			at.StatusEQ(at.StatusQueued),
			at.Or(
				at.CreatedAtLT(turn.CreatedAt),
				at.And(
					at.CreatedAtEQ(turn.CreatedAt),
					at.IDLT(turn.ID),
				),
			),
		)

		// query for any other turns of the same session that are currently running or have been queued for longer
		queryOthers := sess.QueryTurns().
			Where(at.AgentSessionID(sess.ID), at.IDNEQ(turn.ID),
				at.Or(at.StatusEQ(at.StatusRunning), isQueuedAndOlder))
		othersExist, queryOthersErr := queryOthers.Exist(ctx)
		if queryOthersErr != nil {
			return fmt.Errorf("query running agent turn: %w", queryOthersErr)
		}
		if othersExist {
			return river.JobSnooze(time.Second * 5)
		}

		queryParent := sess.QueryTurns().
			Where(at.StatusEQ(at.StatusCompleted), at.Not(at.HasChildrenWith(at.StatusEQ(at.StatusCompleted))))
		parent, queryParentErr := queryParent.Only(ctx)
		if queryParentErr != nil && !ent.IsNotFound(queryParentErr) {
			return fmt.Errorf("query successful agent turn: %w", queryParentErr)
		}

		var parentId *uuid.UUID
		var parentState []byte
		if parent != nil {
			parentId = &parent.ID
			if parent.State == nil || len(parent.State) == 0 {
				return fmt.Errorf("successful parent turn has no state")
			}
			parentState = append(parentState, parent.State...)
		} else {
			queryByYoungest := sess.QueryTurns().
				Order(at.ByCreatedAt(sql.OrderAsc()), at.ByID(sql.OrderAsc()))
			initialTurn, queryInitialTurnErr := queryByYoungest.First(ctx)
			if queryInitialTurnErr != nil {
				return fmt.Errorf("query initial agent turn: %w", queryInitialTurnErr)
			}
			if initialTurn.ID != turn.ID {
				return fmt.Errorf("session has no successful root turn")
			}
		}

		update := turn.Update().ClearStateFields().
			SetNillableParentID(parentId).
			SetStatus(at.StatusRunning).
			SetStartedAt(time.Now().UTC())

		started, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("claim agent turn: %w", updateErr)
		}

		claim = &claimedAgentTurn{
			session:     sess.Unwrap(),
			turn:        started.Unwrap(),
			parentState: parentState,
		}
		return nil
	})
}

func (w *agentTurnWorker) saveInvocationError(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], err error) error {
	return w.updateClaimedTurn(ctx, job, func(currStatus at.Status, u *ent.AgentTurnUpdateOne) error {
		if currStatus != at.StatusRunning && currStatus != at.StatusQueued {
			switch currStatus {
			case at.StatusCompleted, at.StatusFailed:
				return nil
			case at.StatusAborted:
				return river.JobCancel(fmt.Errorf("agent turn was aborted"))
			default:
				return fmt.Errorf("cannot fail agent turn in status %q", currStatus)
			}
		}

		encErr, jsonErr := json.Marshal(core.AsGenkitError(err))
		if jsonErr != nil {
			return fmt.Errorf("encode invocation error: %w", err)
		}
		u.SetStatus(at.StatusFailed)
		u.SetFinishedAt(time.Now().UTC())
		u.SetFinishReason(string(aix.AgentFinishReasonFailed))
		u.SetError(encErr)
		u.ClearState()
		return nil
	})
}

func (w *agentTurnWorker) saveInvocationResult(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], result *rez.AgentTurnResult) error {
	return w.updateClaimedTurn(ctx, job, func(currStatus at.Status, u *ent.AgentTurnUpdateOne) error {
		if currStatus != at.StatusRunning {
			switch currStatus {
			case at.StatusCompleted, at.StatusFailed:
				return nil
			case at.StatusAborted:
				return river.JobCancel(fmt.Errorf("agent turn was aborted"))
			}
			return fmt.Errorf("agent turn is %s, expected running", currStatus)
		}

		var resultErr *core.GenkitError
		u.ClearState()
		if result == nil {
			resultErr = core.AsGenkitError(errors.New("agent returned no result"))
		} else if result.Error != nil {
			resultErr = result.Error
			if json.Valid(result.State) {
				u.SetState(result.State)
			}
		} else if !json.Valid(result.State) {
			resultErr = core.AsGenkitError(errors.New("successful agent result has invalid state"))
			u.ClearState()
		} else {
			u.SetState(result.State)
		}

		u.SetFinishedAt(time.Now().UTC())
		if resultErr != nil {
			encodedErr, jsonErr := json.Marshal(resultErr)
			if jsonErr != nil {
				return fmt.Errorf("encode agent result error: %w", jsonErr)
			}
			u.SetStatus(at.StatusFailed)
			u.SetFinishReason(string(aix.AgentFinishReasonFailed))
			u.SetError(encodedErr)
		} else {
			u.SetStatus(at.StatusCompleted)
			u.SetFinishReason(string(result.FinishReason))
			u.ClearError()
		}

		return nil
	})
}

func (w *agentTurnWorker) updateClaimedTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], setFn func(at.Status, *ent.AgentTurnUpdateOne) error) error {
	cleanupCancel := func() {}
	if ctx.Err() != nil {
		ctx, cleanupCancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	}
	defer cleanupCancel()

	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := acquireAgentSessionTurnLock(ctx, w.db, job.Args.AgentSessionID); lockErr != nil {
			return fmt.Errorf("acquire agent session lock: %w", lockErr)
		}

		turn, lookupTurnErr := tx.AgentTurn.Get(ctx, job.Args.AgentTurnID)
		if lookupTurnErr != nil {
			return fmt.Errorf("reload agent turn: %w", lookupTurnErr)
		}
		if turn.AgentSessionID != job.Args.AgentSessionID {
			return river.JobCancel(fmt.Errorf("invalid job turn session"))
		}
		if turn.RiverJobID != job.ID {
			return river.JobCancel(fmt.Errorf("stale agent turn job"))
		}

		update := turn.Update()
		if setErr := setFn(turn.Status, update); setErr != nil {
			return setErr
		}

		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("persist terminal agent turn: %w", updateErr)
		}
		return nil
	})
}
