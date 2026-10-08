package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/riverqueue/river"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	agta "github.com/rezible/rezible/ent/agentartifact"
	agtm "github.com/rezible/rezible/ent/agentmessage"
	agts "github.com/rezible/rezible/ent/agentsession"
	agtt "github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type StartAgentSessionWorker struct {
	river.WorkerDefaults[jobs.StartAgentSession]

	db       rez.Database
	agents   rez.AiAgentCatalogue
	sessions rez.AiAgentSessionService
	timeout  time.Duration
}

func NewStartAgentSessionWorker(cfg rez.AiConfig, db rez.Database, agents rez.AiAgentCatalogue, sessions rez.AiAgentSessionService) (*StartAgentSessionWorker, error) {
	w := &StartAgentSessionWorker{
		db:       db,
		agents:   agents,
		sessions: sessions,
		timeout:  cfg.Agents.WorkerTimeout,
	}
	return w, nil
}

func (w *StartAgentSessionWorker) Timeout(*river.Job[jobs.StartAgentSession]) time.Duration {
	return w.timeout
}

func (w *StartAgentSessionWorker) Work(ctx context.Context, job *river.Job[jobs.StartAgentSession]) error {
	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := acquireAgentSessionTurnLock(ctx, w.db, job.Args.SessionID); lockErr != nil {
			return lockErr
		}

		sess, sessErr := tx.AgentSession.Get(ctx, job.Args.SessionID)
		if sessErr != nil {
			if ent.IsNotFound(sessErr) {
				return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
			}
			return fmt.Errorf("get session: %w", sessErr)
		}
		hasTurns, queryTurnsErr := sess.QueryTurns().Exist(ctx)
		if queryTurnsErr != nil {
			return fmt.Errorf("query existing turns: %w", queryTurnsErr)
		} else if hasTurns {
			return nil
		}
		slog.InfoContext(ctx, "making initial agent turn input", "agent_session_id", job.Args.SessionID)

		initialInput, initErr := w.agents.MakeInitialTurnInput(ctx, sess)
		if initErr != nil {
			return fmt.Errorf("make initial agent turn input: %w", initErr)
		} else if initialInput == nil {
			return fmt.Errorf("prepare agent session: nil result")
		}
		slog.InfoContext(ctx, "requesting initial agent turn", "agent_session_id", job.Args.SessionID)

		turn, turnErr := w.sessions.RequestAgentTurn(ctx, sess.ID, &rez.RequestAiAgentTurnParams{Input: initialInput})
		if turnErr != nil {
			return fmt.Errorf("request agent turn: %w", turnErr)
		}
		slog.InfoContext(ctx, "requested initial agent turn", "agent_session_id", job.Args.SessionID, "agent_turn_id", turn.ID)
		return nil
	})
}

type InvokeAgentTurnWorker struct {
	river.WorkerDefaults[jobs.InvokeAgentTurn]

	db       rez.Database
	msgs     rez.MessageQueue
	agents   rez.AiAgentRuntime
	sessions rez.AiAgentSessionService
	timeout  time.Duration
}

func NewInvokeAgentTurnWorker(cfg rez.AiConfig, db rez.Database, msgs rez.MessageQueue, aiSvc rez.AiAgentRuntime, aiSess rez.AiAgentSessionService) (*InvokeAgentTurnWorker, error) {
	w := &InvokeAgentTurnWorker{
		db:       db,
		msgs:     msgs,
		agents:   aiSvc,
		sessions: aiSess,
		timeout:  cfg.Agents.WorkerTimeout,
	}
	return w, nil
}

func (w *InvokeAgentTurnWorker) Timeout(*river.Job[jobs.InvokeAgentTurn]) time.Duration {
	return w.timeout
}

func (w *InvokeAgentTurnWorker) Work(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) error {
	claim, claimErr := w.claimAgentTurn(ctx, job)
	if claimErr != nil {
		if ent.IsNotFound(claimErr) {
			return river.JobCancel(fmt.Errorf("agent turn no longer exists"))
		}
		if errors.Is(claimErr, &river.JobCancelError{}) || errors.Is(claimErr, &river.JobSnoozeError{}) {
			return claimErr
		}
		if errors.Is(claimErr, errAgentTurnAlreadyRunning) && job.Attempt < job.MaxAttempts {
			return claimErr
		}
		if job.Attempt >= job.MaxAttempts {
			_, saveErr := w.saveInvocationResult(ctx, job, nil, nil, claimErr)
			return saveErr
		}
		return claimErr
	}
	if claim == nil {
		return nil
	}

	return execution.Do(ctx, "agent.turn", func(ctx context.Context) error {
		return w.runClaimedTurn(ctx, job, claim)
	}, claim.spanAttributes()...)
}

// runClaimedTurn invokes the agent and saves the result, then sets how the attempt ended on the turn's span.
// An attempt queued again for a retry failed.
func (w *InvokeAgentTurnWorker) runClaimedTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], claim *agentTurnClaim) error {
	result, invokeErr := w.invokeTurn(ctx, *claim)
	saved, saveResultErr := w.saveInvocationResult(ctx, job, claim, result, invokeErr)

	outcome := "failed"
	if saved != nil {
		switch saved.Status {
		case agtt.StatusCompleted:
			outcome = "completed"
		case agtt.StatusAborted:
			outcome = "aborted"
		}
		if saved.FinishReason != "" {
			trace.SpanFromContext(ctx).SetAttributes(attribute.String("finish_reason", saved.FinishReason))
		}
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("outcome", outcome))

	if saveResultErr != nil {
		return fmt.Errorf("save result: %w", saveResultErr)
	}
	if invokeErr != nil {
		return fmt.Errorf("invoke: %w", invokeErr)
	}
	if result == nil {
		return fmt.Errorf("agent returned no result")
	}
	if result.Error != nil {
		if errors.Is(result.Error, ai.ErrMaxTurnsExceeded) {
			return river.JobCancel(fmt.Errorf("result: %w", result.Error))
		}
		return fmt.Errorf("result: %w", result.Error)
	}
	return nil
}

type agentTurnClaim struct {
	session   *ent.AgentSession
	turn      *ent.AgentTurn
	currState rez.AiAgentTurnState
	// resumeMessages are the turn's input message and the messages an earlier attempt committed, when it
	// committed any. The attempt then continues from them instead of starting from its input.
	resumeMessages []*ai.Message
	input          *rez.AiAgentTurnInput
}

// spanAttributes identify the turn on its span and logs, with the situation when the session is a situation's
// investigation.
func (c *agentTurnClaim) spanAttributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("agent", c.session.AgentName),
		attribute.String("agent_session_id", c.session.ID.String()),
		attribute.String("agent_turn_id", c.turn.ID.String()),
	}
	if investigation := c.session.Edges.Investigation; investigation != nil {
		for _, link := range investigation.Edges.Situations {
			attrs = append(attrs, attribute.String("situation_id", link.SituationID.String()))
		}
	}
	return attrs
}

func (c *agentTurnClaim) invocationState() rez.AiAgentTurnState {
	return rez.AiAgentTurnState{
		Messages:  slices.Concat(c.currState.Messages, c.resumeMessages),
		Artifacts: c.currState.Artifacts,
	}
}

func (c *agentTurnClaim) newOutputMessagesFromState(stateMessages []*ai.Message) []*ai.Message {
	if c == nil {
		return nil
	}
	if len(stateMessages) <= len(c.currState.Messages) {
		return nil
	}
	newMessages := stateMessages[len(c.currState.Messages):]
	if c.input != nil && c.input.Message != nil && len(newMessages) > 0 &&
		newMessages[0].Role == ai.RoleUser {
		newMessages = newMessages[1:]
	}
	return newMessages
}

var errAgentTurnAlreadyRunning = fmt.Errorf("agent turn already running")

func (w *InvokeAgentTurnWorker) claimAgentTurn(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn]) (*agentTurnClaim, error) {
	return ent.WithTxReturning(ctx, w.db, func(ctx context.Context, tx *ent.Client) (*agentTurnClaim, error) {
		if lockErr := acquireAgentSessionTurnLock(ctx, w.db, job.Args.AgentSessionID); lockErr != nil {
			return nil, fmt.Errorf("acquire agent session lock: %w", lockErr)
		}

		querySession := tx.AgentSession.Query().
			Where(agts.ID(job.Args.AgentSessionID)).
			WithArtifacts().
			WithInvestigation(func(q *ent.InvestigationQuery) { q.WithSituations() })
		sess, sessErr := querySession.Only(ctx)
		if sessErr != nil {
			if ent.IsNotFound(sessErr) {
				return nil, river.JobCancel(fmt.Errorf("no agent session exists"))
			}
			return nil, fmt.Errorf("lookup agent session: %w", sessErr)
		}

		queryTurn := sess.QueryTurns().
			Where(agtt.ID(job.Args.AgentTurnID), agtt.RiverJobID(job.ID))
		turn, turnErr := queryTurn.Only(ctx)
		if turnErr != nil {
			if ent.IsNotFound(turnErr) {
				return nil, river.JobCancel(fmt.Errorf("agent turn no longer exists"))
			}
			return nil, fmt.Errorf("reload agent turn: %w", turnErr)
		}

		if turn.Status != agtt.StatusQueued {
			switch turn.Status {
			case agtt.StatusCompleted:
				return nil, nil
			case agtt.StatusFailed, agtt.StatusAborted:
				return nil, river.JobCancel(fmt.Errorf("agent turn is already %s", turn.Status))
			case agtt.StatusRunning:
				return nil, errAgentTurnAlreadyRunning
			}
			return nil, fmt.Errorf("invalid agent turn status %q", turn.Status)
		}

		// query for any other turns of the same session that are currently running or have been queued for longer
		pIsQueuedAndOlder := agtt.And(agtt.StatusEQ(agtt.StatusQueued), agtt.SequenceLT(turn.Sequence))
		queryOthers := sess.QueryTurns().
			Where(agtt.IDNEQ(turn.ID), agtt.Or(agtt.StatusEQ(agtt.StatusRunning), pIsQueuedAndOlder))
		othersExist, queryOthersErr := queryOthers.Exist(ctx)
		if queryOthersErr != nil {
			return nil, fmt.Errorf("query running agent turn: %w", queryOthersErr)
		} else if othersExist {
			return nil, river.JobSnooze(time.Second * 5)
		}

		updateTurn := turn.Update().
			SetFinishReason("").
			ClearFinishedAt().
			ClearError().
			SetStatus(agtt.StatusRunning).
			// Match PostgreSQL precision so live chunks and reloaded status events share an identity.
			SetStartedAt(time.Now().UTC().Truncate(time.Microsecond))

		var updateTurnErr error
		turn, updateTurnErr = updateTurn.Save(ctx)
		if updateTurnErr != nil {
			return nil, fmt.Errorf("claim agent turn: %w", updateTurnErr)
		}

		sessArtifacts, sessArtifactsErr := sess.Edges.ArtifactsOrErr()
		if sessArtifactsErr != nil {
			return nil, fmt.Errorf("query session artifacts: %w", sessArtifactsErr)
		}

		queryCompletedMessages := tx.AgentMessage.Query().
			Where(
				agtm.AgentSessionID(sess.ID),
				agtm.HasAgentTurnWith(
					agtt.AgentSessionID(sess.ID),
					agtt.SequenceLT(turn.Sequence),
					agtt.StatusEQ(agtt.StatusCompleted),
				),
			).
			// A retried turn's output is re-sequenced after newer turns' messages; ordering by turn keeps each turn's messages together.
			Order(agtm.ByAgentTurnField(agtt.FieldSequence, sql.OrderAsc()), agtm.BySequence(sql.OrderAsc()))
		msgs, msgsErr := queryCompletedMessages.All(ctx)
		if msgsErr != nil {
			return nil, fmt.Errorf("query session messages: %w", msgsErr)
		}

		lastCompletedState := rez.AiAgentTurnState{
			Messages:  make([]*ai.Message, len(msgs)),
			Artifacts: make([]*aix.Artifact, len(sessArtifacts)),
		}
		for i, m := range msgs {
			lastCompletedState.Messages[i] = m.MakeGenkitMessage()
		}
		for i, a := range sessArtifacts {
			lastCompletedState.Artifacts[i] = &aix.Artifact{
				Name:     a.Name,
				Metadata: a.Metadata,
				Parts:    a.Parts,
			}
		}

		queryTurnMessages := turn.QueryMessages().
			Order(agtm.BySequence(sql.OrderAsc()))
		turnMsgs, turnMsgsErr := queryTurnMessages.All(ctx)
		if turnMsgsErr != nil {
			return nil, fmt.Errorf("query turn messages: %w", turnMsgsErr)
		}

		var input *rez.AiAgentTurnInput
		hasPreviousMessages := len(turnMsgs) > 0
		if turn.InputMessageID != nil {
			inputMsg, inputMsgErr := turn.QueryInputMessage().Only(ctx)
			if inputMsgErr != nil {
				return nil, fmt.Errorf("query input message: %w", inputMsgErr)
			}
			input = &rez.AiAgentTurnInput{Message: inputMsg.MakeGenkitMessage()}
			hasPreviousMessages = slices.ContainsFunc(turnMsgs, func(m *ent.AgentMessage) bool {
				return m.ID != inputMsg.ID
			})
		} else {
			input = &rez.AiAgentTurnInput{Resume: turn.InputToolResume}
		}

		var turnInputErr error
		input, turnInputErr = normalizeAgentTurnInput(input, hasPreviousMessages)
		if turnInputErr != nil {
			return nil, fmt.Errorf("invalid agent turn input: %w", turnInputErr)
		}

		var resumeMessages []*ai.Message
		if hasPreviousMessages {
			resumeMessages = make([]*ai.Message, len(turnMsgs))
			for i, m := range turnMsgs {
				resumeMessages[i] = m.MakeGenkitMessage()
			}
		}

		if publishErr := w.publishTurnUpdated(ctx, turn); publishErr != nil {
			return nil, fmt.Errorf("publish claimed agent turn update: %w", publishErr)
		}

		return &agentTurnClaim{
			session:        sess,
			currState:      lastCompletedState,
			turn:           turn,
			input:          input,
			resumeMessages: resumeMessages,
		}, nil
	})
}

func (w *InvokeAgentTurnWorker) publishTurnUpdated(ctx context.Context, turn *ent.AgentTurn) error {
	event := rezai.AgentTurnUpdated{
		StartedAt:      turn.StartedAt,
		AgentSessionId: turn.AgentSessionID,
		AgentTurnId:    turn.ID,
		Status:         turn.Status,
		FinishReason:   turn.FinishReason,
	}
	return w.msgs.Publish(ctx, event)
}

func (w *InvokeAgentTurnWorker) invokeTurn(ctx context.Context, claim agentTurnClaim) (*rez.AiAgentInvocationResult, error) {
	return w.agents.InvokeAgentTurn(ctx, rez.InvokeAiAgentTurnParams{
		Session:           claim.session,
		Turn:              claim.turn,
		State:             claim.invocationState(),
		Input:             claim.input,
		ContinueFromState: len(claim.resumeMessages) > 0,
		OnChunk: func(chunk rez.AiAgentTurnChunk) {
			chunkEvent := rezai.EventOnAgentTurnChunk{
				StartedAt:      claim.turn.StartedAt,
				AgentSessionId: claim.session.ID,
				AgentTurnId:    claim.turn.ID,
				Chunk:          chunk,
			}
			if msgErr := w.msgs.PublishLive(ctx, chunkEvent); msgErr != nil {
				slog.WarnContext(ctx, "failed to publish agent turn chunk", "error", msgErr)
			}
		},
	})
}

// saveInvocationResult records the attempt's result on the turn and returns the turn as saved, or as found when
// the result no longer applies to it.
func (w *InvokeAgentTurnWorker) saveInvocationResult(ctx context.Context, job *river.Job[jobs.InvokeAgentTurn], claim *agentTurnClaim, result *rez.AiAgentInvocationResult, invokeErr error) (*ent.AgentTurn, error) {
	cleanupCancel := func() {}
	if ctx.Err() != nil {
		ctx, cleanupCancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	}
	defer cleanupCancel()

	saved, saveErr := ent.WithTxReturning(ctx, w.db, func(ctx context.Context, tx *ent.Client) (*ent.AgentTurn, error) {
		if lockErr := acquireAgentSessionTurnLock(ctx, w.db, job.Args.AgentSessionID); lockErr != nil {
			return nil, fmt.Errorf("acquire agent session lock: %w", lockErr)
		}

		turn, lookupTurnErr := tx.AgentTurn.Get(ctx, job.Args.AgentTurnID)
		if lookupTurnErr != nil {
			return nil, fmt.Errorf("reload agent turn: %w", lookupTurnErr)
		}

		if turn.AgentSessionID != job.Args.AgentSessionID {
			return nil, river.JobCancel(fmt.Errorf("invalid job turn session"))
		}
		if turn.RiverJobID != job.ID {
			return nil, river.JobCancel(fmt.Errorf("stale agent turn job"))
		}

		isQueued := turn.Status == agtt.StatusQueued && claim == nil && invokeErr != nil && job.Attempt >= job.MaxAttempts
		if turn.Status != agtt.StatusRunning && !isQueued {
			switch turn.Status {
			case agtt.StatusCompleted, agtt.StatusFailed, agtt.StatusAborted:
				return turn, nil
			}
			return nil, fmt.Errorf("agent turn is %s, expected running", turn.Status)
		}

		u := tx.AgentTurn.UpdateOne(turn).
			SetFinishedAt(time.Now().UTC())

		setErrorFn := func(msg string, retryable bool) {
			var finishReason string
			if retryable && job.Attempt < job.MaxAttempts {
				u.SetStatus(agtt.StatusQueued)
				u.ClearFinishedAt()
			} else {
				finishReason = string(aix.AgentFinishReasonFailed)
				u.SetStatus(agtt.StatusFailed)
			}
			u.SetFinishReason(finishReason)
			u.SetError(msg)
		}

		if invokeErr != nil {
			setErrorFn(invokeErr.Error(), true)
		} else if result == nil {
			setErrorFn("agent returned no result", true)
		} else if result.Error != nil {
			// Running out of tool iterations is final; only a person's retry grants another allowance.
			setErrorFn(result.Error.Error(), !errors.Is(result.Error, ai.ErrMaxTurnsExceeded))
		} else {
			u.SetStatus(agtt.StatusCompleted)
			u.SetFinishReason(string(result.FinishReason))
			u.ClearError()
		}

		updated, updateErr := u.Save(ctx)
		if updateErr != nil {
			return nil, fmt.Errorf("save agent turn result: %w", updateErr)
		}

		if publishErr := w.publishTurnUpdated(ctx, updated); publishErr != nil {
			return nil, fmt.Errorf("publish agent turn update: %w", publishErr)
		}

		if result == nil || claim == nil {
			return updated, nil
		}

		// A failed attempt's state is the resume point Genkit committed; keep it for the turn's next attempt.
		// A failed output without state keeps what an earlier attempt saved.
		if updated.Status == agtt.StatusCompleted || len(result.State.Messages) > 0 {
			newMessages := claim.newOutputMessagesFromState(result.State.Messages)
			if replaceOutputErr := w.replaceTurnOutputMessages(ctx, turn, newMessages); replaceOutputErr != nil {
				return nil, replaceOutputErr
			}
		}

		if updated.Status != agtt.StatusCompleted {
			return updated, nil
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
				OnConflictColumns(agta.FieldAgentSessionID, agta.FieldName).
				Update(func(u *ent.AgentArtifactUpsert) {
					u.UpdateLastAgentTurnID()
					u.UpdateMetadata()
					u.UpdateParts()
					u.UpdateUpdatedAt()
				})
			if artifactsErr := upsertArtifacts.Exec(ctx); artifactsErr != nil {
				return nil, fmt.Errorf("update agent artifacts: %w", artifactsErr)
			}
		}

		if updated.Status == agtt.StatusCompleted {
			ev := &rezai.EventOnAgentTurnFinished{
				AgentSessionId:       updated.AgentSessionID,
				AgentSessionMetadata: claim.session.Metadata,
				AgentTurnId:          updated.ID,
				FinishReason:         result.FinishReason,
				Response:             result.Response,
			}
			if publishErr := w.msgs.Publish(ctx, ev); publishErr != nil {
				return nil, fmt.Errorf("publish agent turn finished event: %w", publishErr)
			}
		}

		return updated, nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	if saved.Status == agtt.StatusAborted {
		return saved, river.JobCancel(fmt.Errorf("agent turn was aborted"))
	}
	return saved, nil
}

// replaceTurnOutputMessages replaces the turn's messages other than its input message. A resumed attempt's
// state already holds the earlier attempt's messages, so appending would duplicate them.
func (w *InvokeAgentTurnWorker) replaceTurnOutputMessages(ctx context.Context, turn *ent.AgentTurn, messages []*ai.Message) error {
	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		deleteOutput := tx.AgentMessage.Delete().
			Where(agtm.AgentSessionID(turn.AgentSessionID), agtm.AgentTurnID(turn.ID))
		if turn.InputMessageID != nil {
			deleteOutput.Where(agtm.IDNEQ(*turn.InputMessageID))
		}
		if _, deleteErr := deleteOutput.Exec(ctx); deleteErr != nil {
			return fmt.Errorf("delete previous agent turn messages: %w", deleteErr)
		}
		if len(messages) == 0 {
			return nil
		}

		sessMsgs := tx.AgentMessage.Query().
			Where(agtm.AgentSessionID(turn.AgentSessionID))
		nextSequence, sequenceErr := getNextAgentSessionMessageSequence(ctx, sessMsgs)
		if sequenceErr != nil {
			return sequenceErr
		}

		createMessages := tx.AgentMessage.MapCreateBulk(messages, func(cm *ent.AgentMessageCreate, i int) {
			msg := messages[i]
			cm.SetAgentSessionID(turn.AgentSessionID)
			cm.SetAgentTurnID(turn.ID)
			cm.SetSequence(nextSequence + i)
			cm.SetContent(msg.Content)
			cm.SetRole(agtm.Role(msg.Role))
			cm.SetMetadata(msg.Metadata)
		})
		if msgsErr := createMessages.Exec(ctx); msgsErr != nil {
			return fmt.Errorf("record agent messages: %w", msgsErr)
		}
		return nil
	})
}
