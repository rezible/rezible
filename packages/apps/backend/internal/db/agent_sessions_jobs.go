package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentartifact"
	"github.com/rezible/rezible/ent/agentmessage"
	at "github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

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

func (w *StartAgentSessionWorker) Timeout(*river.Job[jobs.StartAgentSession]) time.Duration {
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

func (w *InvokeAgentTurnWorker) Timeout(*river.Job[jobs.InvokeAgentTurn]) time.Duration {
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
		if errors.Is(claimErr, errAgentTurnAlreadyRunning) {
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

var errAgentTurnAlreadyRunning = fmt.Errorf("agent turn already running")

func (w *InvokeAgentTurnWorker) claimAgentTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) (*agentTurnClaim, error) {
	var claim *agentTurnClaim
	txErr := w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
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
				return errAgentTurnAlreadyRunning
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
			claim.state.Artifacts[i] = &aix.Artifact{
				Name:     a.Name,
				Metadata: a.Metadata,
				Parts:    a.Parts,
			}
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
	if txErr != nil || claim == nil {
		return claim, txErr
	}
	w.publishTurnUpdated(ctx, claim.turn)
	return claim, nil
}

func (w *InvokeAgentTurnWorker) publishTurnUpdated(ctx context.Context, turn *ent.AgentTurn) {
	event := rezai.AgentTurnUpdated{
		AgentSessionId: turn.AgentSessionID,
		AgentTurnId:    turn.ID,
		Status:         turn.Status,
		FinishReason:   turn.FinishReason,
	}
	if publishErr := w.msgs.Publish(ctx, event); publishErr != nil {
		w.logger.WarnContext(ctx, "failed to publish agent turn update", "error", publishErr, "sessionId", turn.AgentSessionID, "turnId", turn.ID)
	}
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

	var updatedTurn *ent.AgentTurn
	txErr := w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
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

		u := tx.AgentTurn.UpdateOne(turn).
			SetFinishedAt(time.Now().UTC())

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
		var updateErr error
		updatedTurn, updateErr = u.Save(ctx)
		if updateErr != nil {
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
	if txErr != nil {
		return txErr
	}
	if updatedTurn != nil {
		w.publishTurnUpdated(ctx, updatedTurn)
	}
	return nil
}
