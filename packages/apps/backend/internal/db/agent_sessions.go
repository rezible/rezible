package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
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
	"github.com/rezible/rezible/pkg/jobs"
)

type AgentSessionService struct {
	logger *slog.Logger
	db     rez.Database
	jobs   rez.JobService
	msgs   rez.MessageService
}

func NewAgentSessionService(tel rez.TelemetryService, db rez.Database, jobSvc rez.JobService, msgs rez.MessageService) (*AgentSessionService, error) {
	s := &AgentSessionService{
		logger: tel.NewLogger(rez.NewLoggerOptions{Name: "agent_session_service"}),
		db:     db,
		jobs:   jobSvc,
		msgs:   msgs,
	}

	return s, nil
}

func (s *AgentSessionService) GetAgentSession(ctx context.Context, id uuid.UUID) (*ent.AgentSession, error) {
	query := s.db.Client(ctx).AgentSession.Query().
		Where(as.ID(id))
	return query.Only(ctx)
}

func (s *AgentSessionService) ListAgentSessions(ctx context.Context, params rez.ListAgentSessionsParams) (*ent.ListResult[ent.AgentSession], error) {
	query := s.db.Client(ctx).AgentSession.Query().
		Order(as.ByCreatedAt(sql.OrderDesc()), as.ByID(sql.OrderDesc())).
		Where(params.Predicates...)

	for key, val := range params.Metadata {
		query.Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueEQ(as.FieldMetadata, val, sqljson.DotPath(key)))
		})
	}

	return ent.DoListQuery[ent.AgentSession, *ent.AgentSessionQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) ListAgentMessages(ctx context.Context, params rez.ListAgentMessagesParams) (*ent.ListResult[ent.AgentMessage], error) {
	query := s.db.Client(ctx).AgentMessage.Query().
		Where(params.Predicates...).
		Order(agentmessage.BySequence(sql.OrderAsc()), agentmessage.ByID(sql.OrderAsc()))
	return ent.DoListQuery[ent.AgentMessage, *ent.AgentMessageQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) ListAgentArtifacts(ctx context.Context, params rez.ListAgentArtifactsParams) (*ent.ListResult[ent.AgentArtifact], error) {
	query := s.db.Client(ctx).AgentArtifact.Query().
		Where(params.Predicates...).
		Order(agentartifact.ByName(sql.OrderAsc()), agentartifact.ByID(sql.OrderAsc()))
	return ent.DoListQuery[ent.AgentArtifact, *ent.AgentArtifactQuery](ctx, query, params.ListParams)
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
	maps.Copy(metadata, params.Metadata)
	for i, binding := range params.Bindings {
		if validateErr := binding.ProviderResourceRef.Validate(); validateErr != nil {
			return nil, fmt.Errorf("%w: binding %d: %v", rez.ErrInvalidInput, i, validateErr)
		}
	}

	var session *ent.AgentSession
	return session, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		for i, binding := range params.Bindings {
			if validateErr := s.validateSessionBindingIntegration(ctx, binding.ProviderResourceRef, binding.IntegrationID); validateErr != nil {
				return fmt.Errorf("binding %d: %w", i, validateErr)
			}
		}

		createSession := tx.AgentSession.Create().
			SetAgentName(name).
			SetInput(sessionInput).
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
	m.SetProvider(params.Provider)
	m.SetProviderNamespace(params.ProviderNamespace)
	m.SetProviderResourceRef(params.ResourceRef)

	metadata := make(map[string]any, len(params.Metadata))
	maps.Copy(metadata, params.Metadata)
	m.SetMetadata(metadata)
}

func (s *AgentSessionService) validateSessionBindingIntegration(ctx context.Context, ref rez.ProviderResourceRef, integrationID *uuid.UUID) error {
	if integrationID == nil {
		return nil
	}
	intg, queryErr := s.db.Client(ctx).Integration.Get(ctx, *integrationID)
	if queryErr != nil {
		return fmt.Errorf("load integration: %w", queryErr)
	}
	if intg.Provider != ref.Provider {
		return fmt.Errorf("%w: binding provider %q does not match integration provider %q", rez.ErrInvalidInput, ref.Provider, intg.Provider)
	}
	return nil
}

func (s *AgentSessionService) ListAgentSessionBindings(ctx context.Context, params rez.ListAgentSessionBindingsParams) (ent.AgentSessionBindings, error) {
	query := s.db.Client(ctx).AgentSessionBinding.Query().
		Where(params.Predicates...)
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
		var existing *ent.AgentSessionBinding
		if bindingId == uuid.Nil {
			mutator = tx.AgentSessionBinding.Create().SetID(uuid.New())
		} else {
			var queryErr error
			existing, queryErr = tx.AgentSessionBinding.Get(ctx, bindingId)
			if queryErr != nil {
				return fmt.Errorf("load binding: %w", queryErr)
			}
			mutator = existing.Update()
		}

		setFn(mutator.Mutation())
		mutation := mutator.Mutation()
		ref := rez.ProviderResourceRef{}
		var integrationID *uuid.UUID
		if existing != nil {
			ref.Provider = existing.Provider
			ref.ProviderNamespace = existing.ProviderNamespace
			ref.ResourceRef = existing.ProviderResourceRef
			integrationID = existing.IntegrationID
		}
		if provider, ok := mutation.Provider(); ok {
			ref.Provider = provider
		}
		if namespace, ok := mutation.ProviderNamespace(); ok {
			ref.ProviderNamespace = namespace
		}
		if resourceRef, ok := mutation.ProviderResourceRef(); ok {
			ref.ResourceRef = resourceRef
		}
		if id, ok := mutation.IntegrationID(); ok {
			integrationID = &id
		}
		if mutation.IntegrationIDCleared() {
			integrationID = nil
		}
		if validateErr := ref.Validate(); validateErr != nil {
			return fmt.Errorf("%w: binding: %v", rez.ErrInvalidInput, validateErr)
		}
		if validateErr := s.validateSessionBindingIntegration(ctx, ref, integrationID); validateErr != nil {
			return validateErr
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("set binding: %w", saveErr)
		}
		binding = saved.Unwrap()
		return nil
	})
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
	jobArgs := jobs.InvokeAgentTurn{
		AgentSessionID: sessId,
		AgentTurnID:    turnId,
	}
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

func (s *AgentSessionService) queryAgentTurns(ctx context.Context) *ent.AgentTurnQuery {
	return s.db.Client(ctx).AgentTurn.Query()
}

func (s *AgentSessionService) GetAgentTurn(ctx context.Context, id uuid.UUID) (*ent.AgentTurn, error) {
	return s.queryAgentTurns(ctx).
		Where(at.ID(id)).
		Only(ctx)
}

func (s *AgentSessionService) ListAgentTurns(ctx context.Context, params rez.ListAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error) {
	query := s.queryAgentTurns(ctx).
		Where(params.Predicates...).
		Order(at.BySequence(sql.OrderDesc()), at.ByID(sql.OrderDesc()))
	return ent.DoListQuery[ent.AgentTurn, *ent.AgentTurnQuery](ctx, query, params.ListParams)
}

func (s *AgentSessionService) GetLatestTurnForSession(ctx context.Context, sessionId uuid.UUID) (*ent.AgentTurn, error) {
	query := s.queryAgentTurns(ctx).
		Where(at.AgentSessionID(sessionId)).
		Order(at.BySequence(sql.OrderDesc()))
	return query.First(ctx)
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
		savedTurn, saveTurnErr := update.Save(ctx)
		if saveTurnErr != nil {
			return fmt.Errorf("abort agent turn: %w", saveTurnErr)
		}
		result = savedTurn.Unwrap()
		if eventErr := s.publishTurnUpdated(ctx, result); eventErr != nil {
			s.logger.WarnContext(ctx, "failed to publish agent turn update",
				"error", eventErr,
				"sessionId", turn.AgentSessionID,
				"turnId", turn.ID)
		}
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
		savedTurn, saveTurnErr := updateTurn.Save(ctx)
		if saveTurnErr != nil {
			return fmt.Errorf("retry agent turn: %w", saveTurnErr)
		}
		result = savedTurn.Unwrap()

		if eventErr := s.publishTurnUpdated(ctx, result); eventErr != nil {
			s.logger.WarnContext(ctx, "failed to publish agent turn update",
				"error", eventErr,
				"sessionId", turn.AgentSessionID,
				"turnId", turn.ID)
		}
		return nil
	})
}

func (s *AgentSessionService) publishTurnUpdated(ctx context.Context, turn *ent.AgentTurn) error {
	event := rezai.AgentTurnUpdated{
		AgentSessionId: turn.AgentSessionID,
		AgentTurnId:    turn.ID,
		Status:         turn.Status,
		FinishReason:   turn.FinishReason,
	}
	return s.msgs.Publish(ctx, event)
}
