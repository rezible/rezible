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
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentartifact"
	"github.com/rezible/rezible/ent/agentmessage"
	as "github.com/rezible/rezible/ent/agentsession"
	at "github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

type AgentSessionService struct {
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
}

func NewAgentSessionService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService) (*AgentSessionService, error) {
	s := &AgentSessionService{
		logger: tel.NewLogger(rez.NewLoggerOptions{Name: "agent_session_service"}),
		db:     db,
		jobs:   jobSvc,
	}

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

func (s *AgentSessionService) CreateAgentSession(ctx context.Context, params rez.CreateAgentSessionParams) (*ent.AgentSession, error) {
	name := strings.TrimSpace(params.AgentName)
	if name == "" {
		return nil, fmt.Errorf("%w: agent name is required", rez.ErrInvalidInput)
	}

	sessionInput, inputErr := json.Marshal(params.Input)
	if inputErr != nil {
		return nil, fmt.Errorf("marshalling session input: %w", inputErr)
	}

	metadata := make(map[string]any, len(params.Metadata))
	for key, value := range params.Metadata {
		metadata[key] = value
	}

	var session *ent.AgentSession
	return session, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		createSession := tx.AgentSession.Create().
			SetAgentName(name).
			SetInput(sessionInput).
			SetNillableOwnerUserID(params.OwnerUserID).
			SetNillableSystemAnalysisID(params.SystemAnalysisID).
			SetScopes(params.PermissionScopes).
			SetMetadata(metadata)
		createdSession, createErr := createSession.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create agent session: %w", createErr)
		}

		if len(params.Bindings) > 0 {
			createBindings := tx.AgentSessionBinding.
				MapCreateBulk(params.Bindings, func(c *ent.AgentSessionBindingCreate, i int) {
					c.SetAgentSessionID(createdSession.ID)
					s.setSessionBindingParams(c.Mutation(), params.Bindings[i])
				})
			if bindingsErr := createBindings.Exec(ctx); bindingsErr != nil {
				return fmt.Errorf("create agent session bindings: %w", bindingsErr)
			}
		}

		startJobArgs := jobs.StartAgentSession{SessionID: createdSession.ID}
		_, jobErr := s.jobs.Insert(ctx, startJobArgs, nil)
		if jobErr != nil {
			return fmt.Errorf("insert start session job: %w", jobErr)
		}

		session = createdSession.Unwrap()
		return nil
	})
}

func (s *AgentSessionService) setSessionBindingParams(m *ent.AgentSessionBindingMutation, params rez.AgentSessionBindingParams) {
	if params.IntegrationID == nil {
		m.ClearIntegrationID()
	} else {
		m.SetIntegrationID(*params.IntegrationID)
	}
	m.SetSource(strings.TrimSpace(params.Source))
	m.SetResourceKind(strings.TrimSpace(params.ResourceKind))
	m.SetResourceRef(strings.TrimSpace(params.ResourceRef))

	metadata := make(map[string]any, len(params.Metadata))
	for key, value := range params.Metadata {
		metadata[key] = value
	}
	m.SetMetadata(metadata)
}

func (s *AgentSessionService) ListAgentSessionBindings(ctx context.Context, params rez.ListAgentSessionBindingsParams) (ent.AgentSessionBindings, error) {
	query := s.db.Client(ctx).AgentSessionBinding.Query().
		Where(params.Predicates...)
	//if !params.IncludeClosed {
	//	bindingQuery.Where(asb.ClosedAtIsNil())
	//}
	return query.All(ctx)
}

func (s *AgentSessionService) LookupAgentSessionBinding(ctx context.Context, preds ...predicate.AgentSessionBinding) (*ent.AgentSessionBinding, error) {
	query := s.db.Client(ctx).AgentSessionBinding.Query().
		Where(preds...).
		WithAgentSession().
		WithIntegration()
	return query.Only(ctx)
}

func (s *AgentSessionService) SetAgentSessionBinding(ctx context.Context, bindingId uuid.UUID, setFn func(*ent.AgentSessionBindingMutation)) (*ent.AgentSessionBinding, error) {
	var binding *ent.AgentSessionBinding
	return binding, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.AgentSessionBinding, *ent.AgentSessionBindingMutation]
		if bindingId == uuid.Nil {
			mutator = tx.AgentSessionBinding.Create().SetID(uuid.New())
		} else {
			mutator = tx.AgentSessionBinding.UpdateOneID(bindingId)
		}

		setFn(mutator.Mutation())

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create binding: %w", saveErr)
		}
		binding = saved.Unwrap()
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

func normalizeAgentTurnInput(input *rez.AiAgentTurnInput) (*rez.AiAgentTurnInput, error) {
	if input == nil {
		return nil, rez.ErrInvalidInput
	}
	norm := *input
	if norm.Message != nil && norm.Message.Role != ai.RoleUser {
		return nil, fmt.Errorf("input message must be user role")
	}
	if res := norm.Resume; res != nil && len(res.Respond)+len(res.Restart) == 0 {
		norm.Resume = nil
	}
	if (norm.Message == nil && norm.Resume == nil) || (norm.Message != nil && norm.Resume != nil) {
		return nil, rez.ErrInvalidInput
	}
	return &norm, nil
}

func acquireAgentSessionTurnLock(ctx context.Context, db rez.Database, sessionId uuid.UUID) error {
	return db.AcquireTxLocks(ctx, "agent_session", sessionId.String())
}

func (s *AgentSessionService) RequestAgentTurn(ctx context.Context, sessionID uuid.UUID, params *rez.RequestAgentTurnParams) (*ent.AgentTurn, error) {
	if params == nil {
		return nil, fmt.Errorf("%w: turn input nil", rez.ErrInvalidInput)
	}
	input, inputErr := normalizeAgentTurnInput(params.Input)
	if inputErr != nil {
		return nil, fmt.Errorf("turn input: %w", inputErr)
	}

	var turn *ent.AgentTurn
	return turn, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := acquireAgentSessionTurnLock(ctx, s.db, sessionID); lockErr != nil {
			return fmt.Errorf("acquire agent session lock: %w", lockErr)
		}

		querySession := tx.AgentSession.Query().Where(as.ID(sessionID))
		if userID, isUserContext := execution.GetContext(ctx).UserID(); isUserContext {
			querySession.Where(as.OwnerUserID(userID))
		}
		sess, sessErr := querySession.Only(ctx)
		if sessErr != nil {
			return fmt.Errorf("query agent session: %w", sessErr)
		}

		countTurns, countTurnsErr := sess.QueryTurns().Count(ctx)
		if countTurnsErr != nil {
			return fmt.Errorf("count agent turns: %w", countTurnsErr)
		}

		if input.Resume != nil && countTurns == 0 {
			return fmt.Errorf("cannot resume with 0 turns")
		}

		activeExists, activeErr := sess.QueryTurns().
			Where(at.StatusIn(at.StatusQueued, at.StatusRunning)).
			Exist(ctx)
		if activeErr != nil {
			return fmt.Errorf("query active agent turn: %w", activeErr)
		}
		if activeExists {
			return fmt.Errorf("%w: agent session already has an active turn", rez.ErrConflict)
		}

		turnID := uuid.New()

		jobID, jobErr := s.insertInvokeAgentTurnJob(ctx, sessionID, turnID)
		if jobErr != nil {
			return fmt.Errorf("enqueue agent turn job: %w", jobErr)
		}

		createTurn := tx.AgentTurn.Create().
			SetID(turnID).
			SetSequence(countTurns + 1).
			SetAgentSessionID(sessionID).
			SetRiverJobID(jobID).
			SetStatus(at.StatusQueued)
		if input.Resume != nil {
			createTurn.SetInputToolResume(input.Resume)
		}
		savedTurn, saveTurnErr := createTurn.Save(ctx)
		if saveTurnErr != nil {
			return fmt.Errorf("create turn: %w", saveTurnErr)
		}

		turn = savedTurn.Unwrap()

		if input.Message != nil {
			countMsgs, countMsgsErr := sess.QueryMessages().Count(ctx)
			if countMsgsErr != nil {
				return fmt.Errorf("count agent messages: %w", countMsgsErr)
			}

			role := agentmessage.Role(input.Message.Role)
			if roleErr := agentmessage.RoleValidator(role); roleErr != nil {
				return fmt.Errorf("message role validator: %w", roleErr)
			}

			inputMsgId := uuid.New()
			createInputMsg := tx.AgentMessage.Create().
				SetID(inputMsgId).
				SetAgentSessionID(sessionID).
				SetAgentTurnID(turnID).
				SetSequence(countMsgs + 1).
				SetRole(agentmessage.RoleUser).
				SetMetadata(input.Message.Metadata).
				SetContent(input.Message.Content)
			if inputMsgErr := createInputMsg.Exec(ctx); inputMsgErr != nil {
				return fmt.Errorf("save input message: %w", inputMsgErr)
			}
			updateTurnMessageID := tx.AgentTurn.UpdateOneID(turnID).SetInputMessageID(inputMsgId)
			if updateTurnErr := updateTurnMessageID.Exec(ctx); updateTurnErr != nil {
				return fmt.Errorf("set turn input message id: %w", updateTurnErr)
			}
			turn.InputMessageID = &inputMsgId
		}

		return nil
	})
}

func (s *AgentSessionService) insertInvokeAgentTurnJob(ctx context.Context, sessId uuid.UUID, turnId uuid.UUID) (int64, error) {
	jobArgs := jobs.InvokeAgentTurn{AgentSessionID: sessId, AgentTurnID: turnId}
	result, insertErr := s.jobs.Insert(ctx, jobArgs, nil)
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

func (s *AgentSessionService) GetAgentTurn(ctx context.Context, id uuid.UUID) (*ent.AgentTurn, error) {
	return s.queryAgentTurns(ctx).Where(at.ID(id)).Only(ctx)
}

func (s *AgentSessionService) ListAgentTurns(ctx context.Context, params rez.ListAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error) {
	query := s.queryAgentTurns(ctx).
		Where(params.Predicates...).
		Order(at.BySequence(sql.OrderDesc()))
	return ent.DoListQuery[ent.AgentTurn, *ent.AgentTurnQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) GetLatestTurnForSession(ctx context.Context, sessionId uuid.UUID) (*ent.AgentTurn, error) {
	queryByYoungest := s.queryAgentTurns(ctx).
		Where(at.AgentSessionID(sessionId)).
		Order(at.BySequence(sql.OrderDesc()))
	return queryByYoungest.First(ctx)
}

func (s *AgentSessionService) GetLastSuccessfulAgentTurn(ctx context.Context, sessionID uuid.UUID) (*ent.AgentTurn, error) {
	queryTurn := s.queryAgentTurns(ctx).
		Where(at.AgentSessionID(sessionID), at.StatusEQ(at.StatusCompleted)).
		Order(at.BySequence(sql.OrderDesc()))
	turn, turnErr := queryTurn.First(ctx)
	if turnErr != nil && !ent.IsNotFound(turnErr) {
		return nil, turnErr
	}
	return turn, nil
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
		queryActive := sess.QueryTurns().Where(at.IDNEQ(turn.ID), at.StatusIn(at.StatusQueued, at.StatusRunning))
		activeExists, activeErr := queryActive.Exist(ctx)
		if activeErr != nil {
			return fmt.Errorf("query active agent turn: %w", activeErr)
		}
		if activeExists {
			return fmt.Errorf("%w: agent session already has an active turn", rez.ErrConflict)
		}

		jobID, jobErr := s.insertInvokeAgentTurnJob(ctx, sess.ID, turn.ID)
		if jobErr != nil {
			return fmt.Errorf("enqueue agent turn retry: %w", jobErr)
		}

		updateTurn := turn.Update().
			ClearError().
			SetFinishReason("").
			ClearStartedAt().
			ClearFinishedAt().
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

type StartAgentSessionWorker struct {
	river.WorkerDefaults[jobs.StartAgentSession]

	db      rez.Database
	ai      rez.AiService
	aiSess  rez.AgentSessionService
	logger  *slog.Logger
	timeout time.Duration
}

func NewStartAgentSessionWorker(cfg rez.AiConfig, tel rez.TelemetryService, db rez.Database, aiSvc rez.AiService, aiSess rez.AgentSessionService) (*StartAgentSessionWorker, error) {
	w := &StartAgentSessionWorker{
		db:      db,
		ai:      aiSvc,
		aiSess:  aiSess,
		logger:  tel.NewLogger(rez.NewLoggerOptions{Name: "start_agent_session_worker"}),
		timeout: cfg.Agents.WorkerTimeout,
	}
	return w, nil
}

func (w *StartAgentSessionWorker) Timeout(job *river.Job[jobs.StartAgentSession]) time.Duration {
	return w.timeout
}

func (w *StartAgentSessionWorker) Work(ctx context.Context, job *river.Job[jobs.StartAgentSession]) error {
	logger := w.logger.With("session_id", job.Args.SessionID)

	sess, sessErr := w.db.Client(ctx).AgentSession.Get(ctx, job.Args.SessionID)
	if sessErr != nil {
		if ent.IsNotFound(sessErr) {
			return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
		}
		return fmt.Errorf("get session: %w", sessErr)
	}
	logger.Info("making initial agent turn input")

	initialInput, initErr := w.ai.MakeInitialAgentTurnInput(ctx, sess)
	if initErr != nil {
		return fmt.Errorf("prepare agent session: %w", initErr)
	} else if initialInput == nil {
		return fmt.Errorf("prepare agent session: nil result")
	}
	logger.Info("requesting initial agent turn")

	turn, turnErr := w.aiSess.RequestAgentTurn(ctx, sess.ID, &rez.RequestAgentTurnParams{Input: initialInput})
	if turnErr != nil {
		return fmt.Errorf("request agent turn: %w", turnErr)
	}
	logger.Info("requested initial agent turn", "turn_id", turn.ID)

	return nil
}

type InvokeAgentTurnWorker struct {
	river.WorkerDefaults[jobs.InvokeAgentTurn]

	db      rez.Database
	msgs    rez.MessageService
	ai      rez.AiService
	aiSess  rez.AgentSessionService
	logger  *slog.Logger
	timeout time.Duration
}

func NewInvokeAgentTurnWorker(cfg rez.AiConfig, tel rez.TelemetryService, db rez.Database, msgs rez.MessageService, aiSvc rez.AiService, aiSess rez.AgentSessionService) (*InvokeAgentTurnWorker, error) {
	w := &InvokeAgentTurnWorker{
		db:      db,
		msgs:    msgs,
		ai:      aiSvc,
		aiSess:  aiSess,
		logger:  tel.NewLogger(rez.NewLoggerOptions{Name: "invoke_agent_turn_worker"}),
		timeout: cfg.Agents.WorkerTimeout,
	}
	return w, nil
}

func (w *InvokeAgentTurnWorker) Timeout(job *river.Job[jobs.InvokeAgentTurn]) time.Duration {
	return w.timeout
}

func (w *InvokeAgentTurnWorker) Work(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) error {
	logger := w.logger.With("sessionId", job.Args.AgentSessionID, "turnId", job.Args.AgentTurnID)

	claim, claimErr := w.claimAgentTurn(ctx, job)
	if claimErr != nil {
		if ent.IsNotFound(claimErr) {
			return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
		}
		if errors.Is(claimErr, &river.JobCancelError{}) || errors.Is(claimErr, &river.JobSnoozeError{}) {
			return claimErr
		}
		if errors.Is(claimErr, rezai.ErrAgentInterrupted) {
			return w.saveInvocationResult(ctx, job, nil, nil, claimErr)
		}
		if job.Attempt >= job.MaxAttempts {
			return w.saveInvocationResult(ctx, job, nil, nil, claimErr)
		}
		return claimErr
	}
	if claim == nil {
		return nil
	}

	result, invokeErr := w.invokeTurn(ctx, *claim)

	if resultErr := w.saveInvocationResult(ctx, job, claim, result, invokeErr); resultErr != nil {
		return fmt.Errorf("save result: %w", resultErr)
	}
	if result != nil {
		turnFinishedEvent := rezai.EventOnAgentTurnFinished{
			AgentSessionId:       claim.session.ID,
			AgentSessionMetadata: claim.session.Metadata,
			AgentTurnId:          claim.turn.ID,
			FinishReason:         result.FinishReason,
			Response:             result.Response,
		}
		if eventErr := w.msgs.Publish(ctx, &turnFinishedEvent); eventErr != nil {
			logger.Warn("failed to publish event", "error", eventErr)
		}
	}

	return nil
}

type agentTurnClaim struct {
	session *ent.AgentSession
	turn    *ent.AgentTurn
	state   rez.AiAgentTurnState
	input   *rez.AiAgentTurnInput
}

func (c *agentTurnClaim) newOutputMessagesFromState(stateMessages []*ai.Message) []*ai.Message {
	if c == nil {
		return nil
	}
	if len(stateMessages) <= len(c.state.Messages) {
		return nil
	}
	newMessages := stateMessages[len(c.state.Messages):]
	if c.input != nil && c.input.Message != nil && len(newMessages) > 0 &&
		newMessages[0].Role == ai.RoleUser {
		newMessages = newMessages[1:]
	}
	return newMessages
}

func (w *InvokeAgentTurnWorker) claimAgentTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) (*agentTurnClaim, error) {
	var claim *agentTurnClaim
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

		// query for any other turns of the same session that are currently running or have been queued for longer
		queryOthers := sess.QueryTurns().
			Where(
				at.AgentSessionID(sess.ID),
				at.IDNEQ(turn.ID),
				at.Or(
					at.StatusEQ(at.StatusRunning),
					at.And(at.StatusEQ(at.StatusQueued), at.SequenceLT(turn.Sequence))))
		othersExist, queryOthersErr := queryOthers.Exist(ctx)
		if queryOthersErr != nil {
			return fmt.Errorf("query running agent turn: %w", queryOthersErr)
		} else if othersExist {
			return river.JobSnooze(time.Second * 5)
		}

		update := turn.Update().
			SetFinishReason("").
			ClearFinishedAt().
			ClearError().
			SetStatus(at.StatusRunning).
			SetStartedAt(time.Now().UTC())
		startedTurn, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("claim agent turn: %w", updateErr)
		}

		queryMessages := tx.AgentMessage.Query().
			Where(
				agentmessage.AgentSessionID(sess.ID),
				agentmessage.HasAgentTurnWith(
					at.AgentSessionID(sess.ID),
					at.SequenceLT(turn.Sequence),
					at.StatusEQ(at.StatusCompleted),
				),
			).
			Order(agentmessage.BySequence(sql.OrderAsc()))
		msgs, msgsErr := queryMessages.All(ctx)
		if msgsErr != nil {
			return fmt.Errorf("query session messages: %w", msgsErr)
		}

		artifacts, artifactsErr := sess.QueryArtifacts().All(ctx)
		if artifactsErr != nil {
			return fmt.Errorf("query session artifacts: %w", artifactsErr)
		}

		claim = &agentTurnClaim{
			session: sess.Unwrap(),
			turn:    startedTurn.Unwrap(),
			state: rez.AiAgentTurnState{
				Messages:  make([]*ai.Message, len(msgs)),
				Artifacts: make([]*aix.Artifact, len(artifacts)),
			},
		}

		for i, m := range msgs {
			claim.state.Messages[i] = m.MakeGenkitMessage()
		}
		for i, a := range artifacts {
			claim.state.Artifacts[i] = &aix.Artifact{Name: a.Name, Metadata: a.Metadata, Parts: a.Parts}
		}

		var input *rez.AiAgentTurnInput
		if turn.InputMessageID != nil {
			queryInputMsg := tx.AgentMessage.Query().
				Where(
					agentmessage.ID(*turn.InputMessageID),
					agentmessage.AgentSessionID(sess.ID),
					agentmessage.AgentTurnID(turn.ID),
				)
			inputMsg, inputMsgErr := queryInputMsg.Only(ctx)
			if inputMsgErr != nil {
				return fmt.Errorf("query input message: %w", inputMsgErr)
			}
			input = &rez.AiAgentTurnInput{Message: inputMsg.MakeGenkitMessage()}
		} else {
			input = &rez.AiAgentTurnInput{Resume: turn.InputToolResume}
		}

		var inputErr error
		claim.input, inputErr = normalizeAgentTurnInput(input)
		if inputErr != nil {
			return fmt.Errorf("invalid agent turn input: %w", inputErr)
		}

		return nil
	})
}

func (w *InvokeAgentTurnWorker) invokeTurn(ctx context.Context, claim agentTurnClaim) (*rez.AiAgentInvocationResult, error) {
	return w.ai.InvokeAgentTurn(ctx, rez.InvokeAgentTurnParams{
		Session: claim.session,
		Turn:    claim.turn,
		State:   claim.state,
		Input:   claim.input,
		OnChunk: func(chunk rez.AiAgentTurnChunk) {
			chunkEvent := rezai.EventOnAgentTurnChunk{
				AgentSessionId: claim.session.ID,
				AgentTurnId:    claim.turn.ID,
				Chunk:          chunk,
			}
			if msgErr := w.msgs.Publish(ctx, chunkEvent); msgErr != nil {
				w.logger.WarnContext(ctx, "failed to publish agent turn chunk", "error", msgErr)
			}
		},
	})
}

func (w *InvokeAgentTurnWorker) saveInvocationResult(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], claim *agentTurnClaim, result *rez.AiAgentInvocationResult, invokeErr error) error {
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

		if turn.Status != at.StatusRunning {
			switch turn.Status {
			case at.StatusCompleted, at.StatusFailed:
				return nil
			case at.StatusAborted:
				return river.JobCancel(fmt.Errorf("agent turn was aborted"))
			}
			return fmt.Errorf("agent turn is %s, expected running", turn.Status)
		}

		u := tx.AgentTurn.UpdateOne(turn).SetFinishedAt(time.Now().UTC())

		setErrorFn := func(msg string) {
			u.SetStatus(at.StatusFailed)
			u.SetFinishReason(string(aix.AgentFinishReasonFailed))
			u.SetError(msg)
		}

		if invokeErr != nil {
			setErrorFn(invokeErr.Error())
		} else if result == nil {
			setErrorFn("agent returned no result")
		} else if result.Error != nil {
			setErrorFn(result.Error.Error())
		} else {
			u.SetStatus(at.StatusCompleted)
			u.SetFinishReason(string(result.FinishReason))
			u.ClearError()
		}
		if updateErr := u.Exec(ctx); updateErr != nil {
			return fmt.Errorf("save agent turn result: %w", updateErr)
		}

		if result == nil {
			return nil
		}

		var newMessages []*ai.Message
		if claim != nil {
			newMessages = claim.newOutputMessagesFromState(result.State.Messages)
		}
		if len(newMessages) > 0 {
			nextSequence := 1
			queryMessages := tx.AgentMessage.Query().
				Where(agentmessage.AgentSessionID(turn.AgentSessionID)).
				Order(agentmessage.BySequence(sql.OrderDesc()))
			lastMsg, lastMsgErr := queryMessages.First(ctx)
			if lastMsgErr != nil && !ent.IsNotFound(lastMsgErr) {
				return fmt.Errorf("query last agent message: %w", lastMsgErr)
			}
			if lastMsg != nil {
				nextSequence = lastMsg.Sequence + 1
			}

			createMessages := tx.AgentMessage.MapCreateBulk(newMessages, func(cm *ent.AgentMessageCreate, i int) {
				msg := newMessages[i]
				cm.SetAgentSessionID(turn.AgentSessionID)
				cm.SetAgentTurnID(turn.ID)
				cm.SetSequence(nextSequence + i)
				cm.SetContent(msg.Content)
				cm.SetRole(agentmessage.Role(msg.Role))
				cm.SetMetadata(msg.Metadata)
			})
			if msgsErr := createMessages.Exec(ctx); msgsErr != nil {
				return fmt.Errorf("record agent messages: %w", msgsErr)
			}
		}

		if artifacts := result.State.Artifacts; len(artifacts) > 0 {
			upsertArtifacts := tx.AgentArtifact.
				MapCreateBulk(artifacts, func(ca *ent.AgentArtifactCreate, i int) {
					art := artifacts[i]
					ca.SetAgentSessionID(turn.AgentSessionID)
					ca.SetAgentTurnID(turn.ID)
					ca.SetLastAgentTurnID(turn.ID)
					ca.SetName(art.Name)
					ca.SetMetadata(art.Metadata)
					ca.SetParts(art.Parts)
				}).
				OnConflictColumns(agentartifact.FieldAgentSessionID, agentartifact.FieldName).
				Update(func(u *ent.AgentArtifactUpsert) {
					u.UpdateLastAgentTurnID()
					u.UpdateMetadata()
					u.UpdateParts()
					u.UpdateUpdatedAt()
				})
			if artifactsErr := upsertArtifacts.Exec(ctx); artifactsErr != nil {
				return fmt.Errorf("update agent artifacts: %w", artifactsErr)
			}
		}

		return nil
	})
}
