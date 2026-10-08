package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"

	at "github.com/rezible/rezible/ent/agentturn"
	inv "github.com/rezible/rezible/ent/investigation"
	inver "github.com/rezible/rezible/ent/investigationevidencerevision"
	invfv "github.com/rezible/rezible/ent/investigationfindingversion"
	invui "github.com/rezible/rezible/ent/investigationuserinput"
)

type InvestigationService struct {
	db     rez.Database
	agents rez.AiAgentSessionService
	jobs   rez.JobService
}

func NewInvestigationService(database rez.Database, agents rez.AiAgentSessionService, jobService rez.JobService) *InvestigationService {
	return &InvestigationService{db: database, agents: agents, jobs: jobService}
}

func (s *InvestigationService) MessageHandlers() []rez.MessageEventHandler {
	return []rez.MessageEventHandler{
		messages.NewEventHandler("db.InvestigationService.onAgentTurnUpdated", s.onAgentTurnUpdated),
	}
}

func (s *InvestigationService) onAgentTurnUpdated(ctx context.Context, event *rezai.AgentTurnUpdated) error {
	if event.Status != at.StatusCompleted && event.Status != at.StatusFailed && event.Status != at.StatusAborted {
		return nil
	}
	queryInvestigation := s.db.Client(ctx).Investigation.Query().
		Where(inv.AgentSessionID(event.AgentSessionId))
	current, queryErr := queryInvestigation.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil
		}
		return fmt.Errorf("lookup investigation for terminal turn: %w", queryErr)
	}
	if requestErr := s.requestReconcile(ctx, current.ID); requestErr != nil {
		return fmt.Errorf("queue investigation after terminal turn: %w", requestErr)
	}
	return nil
}

func (s *InvestigationService) requestReconcile(ctx context.Context, investigationID uuid.UUID) error {
	args := jobs.ReconcileInvestigation{InvestigationID: investigationID}
	if _, insertErr := s.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert investigation reconciliation job: %w", insertErr)
	}
	return nil
}

type ReconcileInvestigationWorker struct {
	jobs.WorkerDefaults[jobs.ReconcileInvestigation]
	investigations *InvestigationService
}

func NewReconcileInvestigationWorker(investigations *InvestigationService) *ReconcileInvestigationWorker {
	return &ReconcileInvestigationWorker{investigations: investigations}
}

func (w *ReconcileInvestigationWorker) Work(ctx context.Context, job *jobs.Job[jobs.ReconcileInvestigation]) error {
	return w.investigations.ReconcileInvestigation(ctx, job.Args.InvestigationID)
}

func (s *InvestigationService) CreateInvestigation(ctx context.Context, params rez.CreateInvestigationParams) (*ent.Investigation, error) {
	query := strings.TrimSpace(params.Query)
	if params.AnalysisID == uuid.Nil || query == "" {
		return nil, fmt.Errorf("%w: analysis ID and question are required", errs.ErrInvalidInput)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Investigation, error) {
		inUseQuery := tx.Investigation.Query().
			Where(inv.SystemAnalysisID(params.AnalysisID))
		inUse, inUseErr := inUseQuery.Exist(ctx)
		if inUseErr != nil {
			return nil, fmt.Errorf("check system analysis ownership: %w", inUseErr)
		} else if inUse {
			return nil, fmt.Errorf("%w: system analysis is already in use", errs.ErrConflict)
		}

		sessionParams := rez.CreateAiAgentSessionParams{
			AgentName: rezai.InvestigationAgent.Name,
			Input:     rezai.InvestigationAgentSessionInput{Query: query},
		}
		session, sessionErr := s.agents.CreateAgentSession(ctx, sessionParams)
		if sessionErr != nil {
			return nil, fmt.Errorf("create investigation agent session: %w", sessionErr)
		}

		createInvestigation := tx.Investigation.Create().
			SetSystemAnalysisID(params.AnalysisID).
			SetAgentSessionID(session.ID)
		created, saveErr := createInvestigation.Save(ctx)
		if saveErr != nil {
			if _, isConstraint := s.db.IsConstraintError(saveErr); isConstraint {
				return nil, fmt.Errorf("%w: system analysis is already in use", errs.ErrConflict)
			}
			return nil, fmt.Errorf("create investigation: %w", saveErr)
		}

		loaded, lookupErr := s.LookupInvestigation(ctx, inv.ID(created.ID))
		if lookupErr != nil {
			return nil, fmt.Errorf("load created investigation: %w", lookupErr)
		}
		return loaded, nil
	})
}

func (s *InvestigationService) GetInvestigation(ctx context.Context, id uuid.UUID) (*ent.Investigation, error) {
	return s.LookupInvestigation(ctx, inv.ID(id))
}

func (s *InvestigationService) LookupInvestigation(ctx context.Context, predicates ...predicate.Investigation) (*ent.Investigation, error) {
	return s.db.Client(ctx).Investigation.Query().
		Where(predicates...).
		WithSystemAnalysis().
		WithAgentSession().
		WithSituations().
		Only(ctx)
}

func (s *InvestigationService) ReadInvestigationDetail(ctx context.Context, invId uuid.UUID) (*rez.InvestigationDetail, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.InvestigationDetail, error) {
		queryInvestigation := tx.Investigation.Query().
			Where(inv.ID(invId))
		curr, getErr := queryInvestigation.Only(ctx)
		if getErr != nil {
			return nil, fmt.Errorf("get investigation detail: %w", getErr)
		}

		session, sessionErr := curr.QueryAgentSession().Only(ctx)
		if sessionErr != nil {
			return nil, fmt.Errorf("load investigation session: %w", sessionErr)
		}
		var input rezai.InvestigationAgentSessionInput
		if decodeErr := json.Unmarshal(session.Input, &input); decodeErr != nil {
			return nil, fmt.Errorf("decode investigation question: %w", decodeErr)
		}

		queryTurns := session.QueryTurns().
			Order(at.BySequence(sql.OrderDesc()), at.ByID(sql.OrderDesc()))

		latestTurn, latestErr := queryTurns.Clone().First(ctx)
		if latestErr != nil && !ent.IsNotFound(latestErr) {
			return nil, fmt.Errorf("load latest investigation turn: %w", latestErr)
		}

		queryActiveTurn := queryTurns.Clone().
			Where(at.StatusIn(at.StatusQueued, at.StatusRunning))
		activeTurn, activeErr := queryActiveTurn.First(ctx)
		if activeErr != nil && !ent.IsNotFound(activeErr) {
			return nil, fmt.Errorf("load active investigation turn: %w", activeErr)
		}

		queryPendingInputs := curr.QueryUserInputs().
			Where(invui.AgentTurnIDIsNil())
		hasPendingInputs, pendingInputsErr := queryPendingInputs.Exist(ctx)
		if pendingInputsErr != nil {
			return nil, fmt.Errorf("check pending investigation inputs: %w", pendingInputsErr)
		}

		queryPendingRevisions := curr.QueryEvidenceRevisions().
			Where(inver.AgentTurnIDIsNil())
		pendingRevisions, pendingRevisionsErr := queryPendingRevisions.Count(ctx)
		if pendingRevisionsErr != nil {
			return nil, fmt.Errorf("count pending investigation evidence revisions: %w", pendingRevisionsErr)
		}

		automaticUpdatesPaused := false
		if pendingRevisions > 0 {
			evidenceTurns, countErr := s.countEvidenceTurns(ctx, curr.ID)
			if countErr != nil {
				return nil, countErr
			}
			automaticUpdatesPaused = evidenceTurns >= MaxEvidenceTurns
		}

		evidenceCurrentAsOf := curr.CreatedAt
		queryReflectedRevision := curr.QueryEvidenceRevisions().
			Where(inver.HasAgentTurnWith(at.StatusEQ(at.StatusCompleted))).
			Order(inver.ByCreatedAt(sql.OrderDesc()), inver.ByID(sql.OrderDesc()))
		reflectedRevision, reflectedErr := queryReflectedRevision.First(ctx)
		if reflectedErr == nil {
			evidenceCurrentAsOf = reflectedRevision.CreatedAt
		} else if !ent.IsNotFound(reflectedErr) {
			return nil, fmt.Errorf("load newest reflected investigation evidence revision: %w", reflectedErr)
		}

		detail := &rez.InvestigationDetail{
			Investigation:            curr,
			Query:                    input.Query,
			LatestTurn:               latestTurn,
			ActiveTurn:               activeTurn,
			HasPendingWork:           hasPendingInputs || pendingRevisions > 0,
			PendingEvidenceRevisions: pendingRevisions,
			AutomaticUpdatesPaused:   automaticUpdatesPaused,
			EvidenceCurrentAsOf:      evidenceCurrentAsOf,
		}
		return detail, nil
	})
}

func (s *InvestigationService) ListInvestigationUserInputs(ctx context.Context, params rez.ListInvestigationUserInputsParams) (*ent.ListResult[ent.InvestigationUserInput], error) {
	query := s.userInputsQuery(ctx).
		Where(invui.InvestigationID(params.InvestigationID)).
		Order(invui.ByCreatedAt(sql.OrderAsc()), invui.ByID(sql.OrderAsc()))
	listed, listErr := ent.DoListQuery[ent.InvestigationUserInput, *ent.InvestigationUserInputQuery](ctx, query, params.ListParams)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation user inputs: %w", listErr)
	}
	return listed, nil
}

// userInputsQuery loads each input's assigned turn and its answer findings, with answer versions from
// running or completed turns ordered newest first (see ent.InvestigationUserInput.CurrentAnswerVersion).
func (s *InvestigationService) userInputsQuery(ctx context.Context) *ent.InvestigationUserInputQuery {
	return s.db.Client(ctx).InvestigationUserInput.Query().
		WithAgentTurn().
		WithFindings(func(fq *ent.InvestigationFindingQuery) {
			fq.WithVersions(func(vq *ent.InvestigationFindingVersionQuery) {
				vq.Where(invfv.HasAgentTurnWith(at.StatusIn(at.StatusRunning, at.StatusCompleted))).
					Order(
						invfv.ByAgentTurnField(at.FieldSequence, sql.OrderDesc()),
						invfv.ByCreatedAt(sql.OrderDesc()),
						invfv.ByID(sql.OrderDesc()),
					)
			})
		})
}

func (s *InvestigationService) ListInvestigationEvidenceRevisions(ctx context.Context, params rez.ListInvestigationEvidenceRevisionsParams) (*ent.ListResult[ent.InvestigationEvidenceRevision], error) {
	investigationID := params.InvestigationID
	if investigationID == uuid.Nil {
		return nil, fmt.Errorf("%w: investigation ID is required", errs.ErrInvalidInput)
	}
	client := s.db.Client(ctx)
	if _, parentErr := client.Investigation.Get(ctx, investigationID); parentErr != nil {
		return nil, fmt.Errorf("get investigation for evidence revisions: %w", parentErr)
	}
	query := client.InvestigationEvidenceRevision.Query().
		Where(inver.InvestigationID(investigationID)).
		Order(inver.ByCreatedAt(sql.OrderAsc()), inver.ByID(sql.OrderAsc()))
	listed, listErr := ent.DoListQuery[ent.InvestigationEvidenceRevision, *ent.InvestigationEvidenceRevisionQuery](ctx, query, params.ListParams)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation evidence revisions: %w", listErr)
	}
	return listed, nil
}

func (s *InvestigationService) SubmitInvestigationUserInput(ctx context.Context, params rez.SubmitInvestigationUserInputParams) (*ent.InvestigationUserInput, error) {
	text := strings.TrimSpace(params.Text)
	submissionKey := strings.TrimSpace(params.SubmissionKey)
	if params.InvestigationID == uuid.Nil || text == "" || submissionKey == "" {
		return nil, fmt.Errorf("%w: investigation ID, text and submission key are required", errs.ErrInvalidInput)
	}
	userID, userIDSet := execution.GetContext(ctx).UserID()
	if !userIDSet || userID == uuid.Nil {
		return nil, fmt.Errorf("%w: authenticated user is required", errs.ErrInvalidInput)
	}

	inputID, txErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (uuid.UUID, error) {
		var inputID uuid.UUID
		if _, investigationErr := tx.Investigation.Get(ctx, params.InvestigationID); investigationErr != nil {
			return uuid.Nil, fmt.Errorf("load investigation for user input: %w", investigationErr)
		}
		lockKey := params.InvestigationID.String() + ":" + submissionKey
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_user_input_submission", lockKey); lockErr != nil {
			return uuid.Nil, fmt.Errorf("lock investigation user input submission: %w", lockErr)
		}

		existingQuery := tx.InvestigationUserInput.Query().Where(
			invui.InvestigationID(params.InvestigationID),
			invui.Key(submissionKey),
		)
		existing, queryErr := existingQuery.Only(ctx)
		if queryErr == nil {
			if existing.Text != text || existing.UserID != userID {
				return uuid.Nil, fmt.Errorf("%w: submission key was already used for different content", errs.ErrConflict)
			}
			inputID = existing.ID
		} else if !ent.IsNotFound(queryErr) {
			return uuid.Nil, fmt.Errorf("lookup investigation user input submission: %w", queryErr)
		} else {
			createInput := tx.InvestigationUserInput.Create().
				SetInvestigationID(params.InvestigationID).
				SetUserID(userID).
				SetText(text).
				SetKey(submissionKey)
			created, saveErr := createInput.Save(ctx)
			if saveErr != nil {
				return uuid.Nil, fmt.Errorf("save investigation user input: %w", saveErr)
			}
			inputID = created.ID
		}

		if requestErr := s.requestReconcile(ctx, params.InvestigationID); requestErr != nil {
			return uuid.Nil, fmt.Errorf("schedule investigation reconciliation: %w", requestErr)
		}
		return inputID, nil
	})
	if txErr != nil {
		return nil, txErr
	}
	queryInput := s.userInputsQuery(ctx).
		Where(invui.ID(inputID))
	input, queryErr := queryInput.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("read investigation user input: %w", queryErr)
	}
	return input, nil
}

func (s *InvestigationService) RecordInvestigationEvidenceRevision(ctx context.Context, params rez.RecordInvestigationEvidenceRevisionParams) (*ent.InvestigationEvidenceRevision, error) {
	explanation := strings.TrimSpace(params.Explanation)
	key := strings.TrimSpace(params.CallerKey)
	if params.InvestigationID == uuid.Nil || explanation == "" || key == "" {
		return nil, fmt.Errorf("%w: investigation ID, explanation and caller key are required", errs.ErrInvalidInput)
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.InvestigationEvidenceRevision, error) {
		if _, investigationErr := tx.Investigation.Get(ctx, params.InvestigationID); investigationErr != nil {
			return nil, fmt.Errorf("load investigation for evidence revision: %w", investigationErr)
		}
		lockKey := params.InvestigationID.String() + ":" + key
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_evidence_follow_up_key", lockKey); lockErr != nil {
			return nil, fmt.Errorf("lock investigation evidence revision key: %w", lockErr)
		}

		existingQuery := tx.InvestigationEvidenceRevision.Query().Where(
			inver.InvestigationID(params.InvestigationID),
			inver.Key(key),
		)
		existing, queryErr := existingQuery.Only(ctx)
		if queryErr == nil {
			if existing.Explanation != explanation {
				return nil, fmt.Errorf("%w: evidence revision key was already used for a different explanation", errs.ErrConflict)
			}
			return existing, nil
		} else if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("lookup investigation evidence revision: %w", queryErr)
		}
		createRevision := tx.InvestigationEvidenceRevision.Create().
			SetInvestigationID(params.InvestigationID).
			SetExplanation(explanation).
			SetKey(key)
		created, saveErr := createRevision.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save investigation evidence revision: %w", saveErr)
		}
		result := created

		if requestErr := s.requestReconcile(ctx, params.InvestigationID); requestErr != nil {
			return nil, fmt.Errorf("schedule investigation reconciliation: %w", requestErr)
		}
		return result, nil
	})
}

// MaxEvidenceTurns is how many turns driven only by evidence revisions an investigation runs on its own.
// After that, pending revisions wait for a person to update it or ask a question.
const MaxEvidenceTurns = 10

func (s *InvestigationService) ReconcileInvestigation(ctx context.Context, investigationID uuid.UUID) error {
	return s.reconcile(ctx, investigationID, false)
}

// UpdateInvestigation starts one turn for pending work, ignoring the automatic evidence turn limit.
func (s *InvestigationService) UpdateInvestigation(ctx context.Context, investigationID uuid.UUID) error {
	return s.reconcile(ctx, investigationID, true)
}

// reconcile gives the oldest pending question, if any, and every pending evidence revision to one new turn
// when no turn is queued or running.
func (s *InvestigationService) reconcile(ctx context.Context, investigationID uuid.UUID, ignoreLimit bool) error {
	if investigationID == uuid.Nil {
		return fmt.Errorf("%w: investigation ID is required", errs.ErrInvalidInput)
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryInvestigation := tx.Investigation.Query().
			Where(inv.ID(investigationID)).
			WithAgentSession()
		rereadQuery := queryInvestigation.Clone()
		current, queryErr := queryInvestigation.Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("load investigation for reconciliation: %w", queryErr)
		}
		if lockErr := s.db.AcquireTxLocks(ctx, "agent_session", current.AgentSessionID.String()); lockErr != nil {
			return fmt.Errorf("lock investigation agent session: %w", lockErr)
		}

		current, queryErr = rereadQuery.Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("reread investigation for reconciliation: %w", queryErr)
		}

		queryTurns := tx.AgentTurn.Query().
			Where(at.AgentSessionID(current.AgentSessionID))
		turnsExist, turnsErr := queryTurns.Exist(ctx)
		if !turnsExist {
			if turnsErr != nil {
				return fmt.Errorf("check investigation startup turn: %w", turnsErr)
			}
			return nil
		}

		queryActiveTurn := tx.AgentTurn.Query().
			Where(at.AgentSessionID(current.AgentSessionID), at.StatusIn(at.StatusQueued, at.StatusRunning))
		activeTurn, activeErr := queryActiveTurn.Exist(ctx)
		if activeErr != nil {
			return fmt.Errorf("check active investigation turn: %w", activeErr)
		}
		if activeTurn {
			return nil
		}

		userInputQuery := tx.InvestigationUserInput.Query().
			Where(invui.InvestigationID(investigationID), invui.AgentTurnIDIsNil()).
			Order(invui.ByCreatedAt(sql.OrderAsc()), invui.ByID(sql.OrderAsc()))
		userInput, userInputErr := userInputQuery.First(ctx)
		if userInputErr != nil && !ent.IsNotFound(userInputErr) {
			return fmt.Errorf("select next investigation user input: %w", userInputErr)
		}
		hasUserInput := userInputErr == nil

		evidenceRevisionsQuery := tx.InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(investigationID), inver.AgentTurnIDIsNil()).
			Order(inver.ByCreatedAt(sql.OrderAsc()), inver.ByID(sql.OrderAsc()))
		evidenceRevisions, evidenceRevisionsErr := evidenceRevisionsQuery.All(ctx)
		if evidenceRevisionsErr != nil {
			return fmt.Errorf("select pending investigation evidence revisions: %w", evidenceRevisionsErr)
		}
		if !hasUserInput && len(evidenceRevisions) == 0 {
			return nil
		}

		if !hasUserInput && !ignoreLimit {
			evidenceTurns, countErr := s.countEvidenceTurns(ctx, investigationID)
			if countErr != nil {
				return countErr
			}
			if evidenceTurns >= MaxEvidenceTurns {
				return nil
			}
		}

		turnInput := &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(s.turnMessage(userInput, evidenceRevisions))}
		turn, requestTurnErr := s.agents.RequestAgentTurn(ctx, current.AgentSessionID, &rez.RequestAiAgentTurnParams{Input: turnInput})
		if requestTurnErr != nil {
			return fmt.Errorf("request investigation turn: %w", requestTurnErr)
		}

		querySessionTurn := tx.AgentTurn.Query().
			Where(at.ID(turn.ID), at.AgentSessionID(current.AgentSessionID))
		turnBelongsToSession, verifyTurnErr := querySessionTurn.Exist(ctx)
		if verifyTurnErr != nil {
			return fmt.Errorf("validate assigned investigation turn: %w", verifyTurnErr)
		}
		if !turnBelongsToSession {
			return fmt.Errorf("%w: requested turn belongs to another investigation", errs.ErrConflict)
		}

		if hasUserInput {
			assignInput := tx.InvestigationUserInput.Update().
				Where(
					invui.ID(userInput.ID),
					invui.InvestigationID(investigationID),
					invui.AgentTurnIDIsNil(),
				).
				SetAgentTurnID(turn.ID)
			updated, updateErr := assignInput.Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("assign investigation user input to turn: %w", updateErr)
			}
			if updated != 1 {
				return fmt.Errorf("%w: investigation user input was assigned concurrently", errs.ErrConflict)
			}
		}
		if len(evidenceRevisions) > 0 {
			revisionIDs := make([]uuid.UUID, 0, len(evidenceRevisions))
			for _, revision := range evidenceRevisions {
				revisionIDs = append(revisionIDs, revision.ID)
			}
			assignRevisions := tx.InvestigationEvidenceRevision.Update().
				Where(
					inver.IDIn(revisionIDs...),
					inver.InvestigationID(investigationID),
					inver.AgentTurnIDIsNil(),
				).
				SetAgentTurnID(turn.ID)
			updated, updateErr := assignRevisions.Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("assign investigation evidence revisions to turn: %w", updateErr)
			}
			if updated != len(revisionIDs) {
				return fmt.Errorf("%w: investigation evidence revisions were assigned concurrently", errs.ErrConflict)
			}
		}
		return nil
	})
}

// turnMessage lists the question, if any, then each distinct evidence revision explanation, oldest first.
func (s *InvestigationService) turnMessage(userInput *ent.InvestigationUserInput, evidenceRevisions []*ent.InvestigationEvidenceRevision) string {
	messageSections := make([]string, 0, 2)
	if userInput != nil {
		messageSections = append(messageSections, "User question:\n"+userInput.Text)
	}
	if len(evidenceRevisions) > 0 {
		seen := mapset.NewSet[string]()
		explanations := make([]string, 0, len(evidenceRevisions))
		for _, revision := range evidenceRevisions {
			if seen.Add(revision.Explanation) {
				explanations = append(explanations, revision.Explanation)
			}
		}
		messageSections = append(messageSections, "Evidence revised:\n"+strings.Join(explanations, "\n"))
	}
	return strings.Join(messageSections, "\n\n")
}

// countEvidenceTurns counts the investigation's turns that took evidence revisions and no question.
func (s *InvestigationService) countEvidenceTurns(ctx context.Context, investigationID uuid.UUID) (int, error) {
	client := s.db.Client(ctx)
	queryQuestionTurns := client.InvestigationUserInput.Query().
		Where(invui.InvestigationID(investigationID), invui.AgentTurnIDNotNil()).
		Select(invui.FieldAgentTurnID)
	var questionTurnIDs []uuid.UUID
	if scanErr := queryQuestionTurns.Scan(ctx, &questionTurnIDs); scanErr != nil {
		return 0, fmt.Errorf("list investigation question turns: %w", scanErr)
	}

	queryEvidenceTurns := client.InvestigationEvidenceRevision.Query().
		Where(
			inver.InvestigationID(investigationID),
			inver.AgentTurnIDNotNil(),
			inver.AgentTurnIDNotIn(questionTurnIDs...),
		).
		Unique(true).
		Select(inver.FieldAgentTurnID)
	evidenceTurns, countErr := queryEvidenceTurns.Count(ctx)
	if countErr != nil {
		return 0, fmt.Errorf("count investigation evidence turns: %w", countErr)
	}
	return evidenceTurns, nil
}
