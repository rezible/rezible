package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"

	at "github.com/rezible/rezible/ent/agentturn"
	inv "github.com/rezible/rezible/ent/investigation"
	inver "github.com/rezible/rezible/ent/investigationevidencerevision"
	invf "github.com/rezible/rezible/ent/investigationfinding"
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
		return nil, fmt.Errorf("%w: analysis ID and question are required", rez.ErrInvalidInput)
	}
	var result *ent.Investigation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		inUseQuery := tx.Investigation.Query().
			Where(inv.SystemAnalysisID(params.AnalysisID))
		inUse, inUseErr := inUseQuery.Exist(ctx)
		if inUseErr != nil {
			return fmt.Errorf("check system analysis ownership: %w", inUseErr)
		} else if inUse {
			return fmt.Errorf("%w: system analysis is already in use", rez.ErrConflict)
		}

		sessionParams := rez.CreateAiAgentSessionParams{
			AgentName: rezai.InvestigationAgent.Name,
			Input:     rezai.InvestigationAgentSessionInput{Query: query},
		}
		session, sessionErr := s.agents.CreateAgentSession(ctx, sessionParams)
		if sessionErr != nil {
			return fmt.Errorf("create investigation agent session: %w", sessionErr)
		}

		createInvestigation := tx.Investigation.Create().
			SetSystemAnalysisID(params.AnalysisID).
			SetAgentSessionID(session.ID)
		created, saveErr := createInvestigation.Save(ctx)
		if saveErr != nil {
			if _, isConstraint := s.db.IsConstraintError(saveErr); isConstraint {
				return fmt.Errorf("%w: system analysis is already in use", rez.ErrConflict)
			}
			return fmt.Errorf("create investigation: %w", saveErr)
		}

		loaded, lookupErr := s.LookupInvestigation(ctx, inv.ID(created.ID))
		if lookupErr != nil {
			return fmt.Errorf("load created investigation: %w", lookupErr)
		}
		result = loaded.Unwrap()
		return nil
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
	var detail *rez.InvestigationDetail
	return detail, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryInvestigation := tx.Investigation.Query().
			Where(inv.ID(invId))
		curr, getErr := queryInvestigation.Only(ctx)
		if getErr != nil {
			return fmt.Errorf("get investigation detail: %w", getErr)
		}

		session, sessionErr := curr.QueryAgentSession().Only(ctx)
		if sessionErr != nil {
			return fmt.Errorf("load investigation session: %w", sessionErr)
		}
		var input rezai.InvestigationAgentSessionInput
		if decodeErr := json.Unmarshal(session.Input, &input); decodeErr != nil {
			return fmt.Errorf("decode investigation question: %w", decodeErr)
		}

		queryTurns := session.QueryTurns().
			Order(at.BySequence(sql.OrderDesc()), at.ByID(sql.OrderDesc()))

		latestTurn, latestErr := queryTurns.Clone().First(ctx)
		if latestErr != nil && !ent.IsNotFound(latestErr) {
			return fmt.Errorf("load latest investigation turn: %w", latestErr)
		}

		queryActiveTurn := queryTurns.Clone().
			Where(at.StatusIn(at.StatusQueued, at.StatusRunning))
		activeTurn, activeErr := queryActiveTurn.First(ctx)
		if activeErr != nil && !ent.IsNotFound(activeErr) {
			return fmt.Errorf("load active investigation turn: %w", activeErr)
		}

		queryPendingInputs := curr.QueryUserInputs().
			Where(invui.AgentTurnIDIsNil())
		hasPendingInputs, pendingInputsErr := queryPendingInputs.Exist(ctx)
		if pendingInputsErr != nil {
			return fmt.Errorf("check pending investigation inputs: %w", pendingInputsErr)
		}

		queryPendingRevisions := curr.QueryEvidenceRevisions().
			Where(inver.AgentTurnIDIsNil())
		hasPendingRevisions, pendingRevisionsErr := queryPendingRevisions.Exist(ctx)
		if pendingRevisionsErr != nil {
			return fmt.Errorf("check pending investigation evidence revisions: %w", pendingRevisionsErr)
		}

		detail = &rez.InvestigationDetail{
			Investigation:  curr,
			Query:          input.Query,
			LatestTurn:     latestTurn,
			ActiveTurn:     activeTurn,
			HasPendingWork: hasPendingInputs || hasPendingRevisions,
		}
		return nil
	})
}

func (s *InvestigationService) ListInvestigationUserInputs(ctx context.Context, investigationID uuid.UUID, params ent.ListParams) (*ent.ListResult[rez.InvestigationUserInput], error) {
	client := s.db.Client(ctx)
	query := client.InvestigationUserInput.Query().
		Where(invui.InvestigationID(investigationID)).
		Order(invui.ByCreatedAt(sql.OrderAsc()), invui.ByID(sql.OrderAsc()))
	listed, listErr := ent.DoListQuery[ent.InvestigationUserInput, *ent.InvestigationUserInputQuery](ctx, query, params)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation user inputs: %w", listErr)
	}

	inputTurns, turnsErr := s.loadAssignedUserInputTurns(ctx, listed.Data)
	if turnsErr != nil {
		return nil, turnsErr
	}

	inputIDs := make([]uuid.UUID, 0, len(listed.Data))
	for _, input := range listed.Data {
		inputIDs = append(inputIDs, input.ID)
	}

	answerVersions := make(map[uuid.UUID]*ent.InvestigationFindingVersion)
	if len(inputIDs) > 0 {
		queryVersions := client.InvestigationFindingVersion.Query().Where(
			invfv.HasFindingWith(invf.UserInputIDIn(inputIDs...)),
		).WithFinding().WithAgentTurn()
		versions, queryErr := queryVersions.All(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("load investigation answer versions: %w", queryErr)
		}
		for _, version := range versions {
			finding := version.Edges.Finding
			turn := version.Edges.AgentTurn
			if finding == nil || finding.UserInputID == nil || turn == nil || !s.eligibleTurnStatus(turn.Status, false) {
				continue
			}
			currentVersion, exists := answerVersions[*finding.UserInputID]
			if !exists || s.outputPublicationNewer(turn, version.CreatedAt, version.ID, currentVersion.Edges.AgentTurn, currentVersion.CreatedAt, currentVersion.ID) {
				answerVersions[*finding.UserInputID] = version
			}
		}
	}
	result := &ent.ListResult[rez.InvestigationUserInput]{
		Data:     make([]*rez.InvestigationUserInput, 0, len(listed.Data)),
		Page:     listed.Page,
		PageSize: listed.PageSize,
		Total:    listed.Total,
	}
	for _, input := range listed.Data {
		item := &rez.InvestigationUserInput{
			ID:            input.ID,
			Text:          input.Text,
			UserID:        input.UserID,
			SubmissionKey: input.Key,
			CreatedAt:     input.CreatedAt,
			AgentTurnID:   input.AgentTurnID,
		}
		if input.AgentTurnID != nil {
			turnStatus := inputTurns[*input.AgentTurnID].Status
			item.TurnStatus = &turnStatus
		}
		if answer := answerVersions[input.ID]; answer != nil {
			item.AnswerVersionID = new(answer.ID)
		}
		result.Data = append(result.Data, item)
	}
	return result, nil
}

func (s *InvestigationService) loadAssignedUserInputTurns(ctx context.Context, inputs []*ent.InvestigationUserInput) (map[uuid.UUID]*ent.AgentTurn, error) {
	turnIDs := mapset.NewSet[uuid.UUID]()
	for _, input := range inputs {
		if input.AgentTurnID != nil {
			turnIDs.Add(*input.AgentTurnID)
		}
	}

	turnsByID := make(map[uuid.UUID]*ent.AgentTurn, turnIDs.Cardinality())
	if turnIDs.Cardinality() == 0 {
		return turnsByID, nil
	}
	queryTurns := s.db.Client(ctx).AgentTurn.Query().
		Where(at.IDIn(turnIDs.ToSlice()...))
	turns, queryErr := queryTurns.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load assigned investigation turns: %w", queryErr)
	}
	for _, turn := range turns {
		turnsByID[turn.ID] = turn
	}
	for _, input := range inputs {
		if input.AgentTurnID == nil {
			continue
		}
		if _, exists := turnsByID[*input.AgentTurnID]; !exists {
			return nil, fmt.Errorf("assigned turn %s for investigation user input %s is missing", *input.AgentTurnID, input.ID)
		}
	}
	return turnsByID, nil
}

func (s *InvestigationService) ListInvestigationEvidenceRevisions(ctx context.Context, investigationID uuid.UUID, params ent.ListParams) (*ent.ListResult[ent.InvestigationEvidenceRevision], error) {
	if investigationID == uuid.Nil {
		return nil, fmt.Errorf("%w: investigation ID is required", rez.ErrInvalidInput)
	}
	client := s.db.Client(ctx)
	if _, parentErr := client.Investigation.Get(ctx, investigationID); parentErr != nil {
		return nil, fmt.Errorf("get investigation for evidence revisions: %w", parentErr)
	}
	query := client.InvestigationEvidenceRevision.Query().
		Where(inver.InvestigationID(investigationID)).
		Order(inver.ByCreatedAt(sql.OrderAsc()), inver.ByID(sql.OrderAsc()))
	listed, listErr := ent.DoListQuery[ent.InvestigationEvidenceRevision, *ent.InvestigationEvidenceRevisionQuery](ctx, query, params)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation evidence revisions: %w", listErr)
	}
	return listed, nil
}

func (s *InvestigationService) SubmitInvestigationUserInput(ctx context.Context, params rez.SubmitInvestigationUserInputParams) (*rez.InvestigationUserInput, error) {
	text := strings.TrimSpace(params.Text)
	submissionKey := strings.TrimSpace(params.SubmissionKey)
	if params.InvestigationID == uuid.Nil || text == "" || submissionKey == "" {
		return nil, fmt.Errorf("%w: investigation ID, text and submission key are required", rez.ErrInvalidInput)
	}
	userID, userIDSet := execution.GetContext(ctx).UserID()
	if !userIDSet || userID == uuid.Nil {
		return nil, fmt.Errorf("%w: authenticated user is required", rez.ErrInvalidInput)
	}

	var result *ent.InvestigationUserInput
	if txErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if _, investigationErr := tx.Investigation.Get(ctx, params.InvestigationID); investigationErr != nil {
			return fmt.Errorf("load investigation for user input: %w", investigationErr)
		}
		lockKey := params.InvestigationID.String() + ":" + submissionKey
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_user_input_submission", lockKey); lockErr != nil {
			return fmt.Errorf("lock investigation user input submission: %w", lockErr)
		}

		existingQuery := tx.InvestigationUserInput.Query().Where(
			invui.InvestigationID(params.InvestigationID),
			invui.Key(submissionKey),
		)
		existing, queryErr := existingQuery.Only(ctx)
		if queryErr == nil {
			if existing.Text != text || existing.UserID != userID {
				return fmt.Errorf("%w: submission key was already used for different content", rez.ErrConflict)
			}
			result = existing
		} else if !ent.IsNotFound(queryErr) {
			return fmt.Errorf("lookup investigation user input submission: %w", queryErr)
		} else {
			createInput := tx.InvestigationUserInput.Create().
				SetInvestigationID(params.InvestigationID).
				SetUserID(userID).
				SetText(text).
				SetKey(submissionKey)
			created, saveErr := createInput.Save(ctx)
			if saveErr != nil {
				return fmt.Errorf("save investigation user input: %w", saveErr)
			}
			result = created.Unwrap()
		}

		if requestErr := s.requestReconcile(ctx, params.InvestigationID); requestErr != nil {
			return fmt.Errorf("schedule investigation reconciliation: %w", requestErr)
		}
		return nil
	}); txErr != nil {
		return nil, txErr
	}
	return s.readInvestigationUserInput(ctx, params.InvestigationID, result.ID)
}

func (s *InvestigationService) readInvestigationUserInput(ctx context.Context, investigationID, inputID uuid.UUID) (*rez.InvestigationUserInput, error) {
	client := s.db.Client(ctx)
	if _, parentErr := client.Investigation.Get(ctx, investigationID); parentErr != nil {
		return nil, fmt.Errorf("get investigation for user input: %w", parentErr)
	}
	queryInput := client.InvestigationUserInput.Query().
		Where(invui.ID(inputID), invui.InvestigationID(investigationID))
	input, queryErr := queryInput.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("read investigation user input: %w", queryErr)
	}
	result := &rez.InvestigationUserInput{
		ID: input.ID, Text: input.Text, UserID: input.UserID,
		SubmissionKey: input.Key, CreatedAt: input.CreatedAt, AgentTurnID: input.AgentTurnID,
	}
	inputTurns, turnsErr := s.loadAssignedUserInputTurns(ctx, []*ent.InvestigationUserInput{input})
	if turnsErr != nil {
		return nil, turnsErr
	}
	if input.AgentTurnID != nil {
		result.TurnStatus = new(inputTurns[*input.AgentTurnID].Status)
	}
	versions, versionsErr := s.latestInvestigationFindingVersions(ctx, investigationID)
	if versionsErr != nil {
		return nil, fmt.Errorf("read current investigation answer: %w", versionsErr)
	}
	for _, version := range versions {
		finding := version.Edges.Finding
		if finding != nil && finding.UserInputID != nil && *finding.UserInputID == input.ID {
			result.AnswerVersionID = new(version.ID)
			break
		}
	}
	return result, nil
}

func (s *InvestigationService) RecordInvestigationEvidenceRevision(ctx context.Context, params rez.RecordInvestigationEvidenceRevisionParams) (*ent.InvestigationEvidenceRevision, error) {
	explanation := strings.TrimSpace(params.Explanation)
	key := strings.TrimSpace(params.CallerKey)
	if params.InvestigationID == uuid.Nil || explanation == "" || key == "" {
		return nil, fmt.Errorf("%w: investigation ID, explanation and caller key are required", rez.ErrInvalidInput)
	}

	var result *ent.InvestigationEvidenceRevision
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if _, investigationErr := tx.Investigation.Get(ctx, params.InvestigationID); investigationErr != nil {
			return fmt.Errorf("load investigation for evidence revision: %w", investigationErr)
		}
		lockKey := params.InvestigationID.String() + ":" + key
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_evidence_follow_up_key", lockKey); lockErr != nil {
			return fmt.Errorf("lock investigation evidence revision key: %w", lockErr)
		}

		existingQuery := tx.InvestigationEvidenceRevision.Query().Where(
			inver.InvestigationID(params.InvestigationID),
			inver.Key(key),
		)
		existing, queryErr := existingQuery.Only(ctx)
		if queryErr == nil {
			if existing.Explanation != explanation {
				return fmt.Errorf("%w: evidence revision key was already used for a different explanation", rez.ErrConflict)
			}
			result = existing
			return nil
		} else if !ent.IsNotFound(queryErr) {
			return fmt.Errorf("lookup investigation evidence revision: %w", queryErr)
		}
		createRevision := tx.InvestigationEvidenceRevision.Create().
			SetInvestigationID(params.InvestigationID).
			SetExplanation(explanation).
			SetKey(key)
		created, saveErr := createRevision.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save investigation evidence revision: %w", saveErr)
		}
		result = created.Unwrap()

		if requestErr := s.requestReconcile(ctx, params.InvestigationID); requestErr != nil {
			return fmt.Errorf("schedule investigation reconciliation: %w", requestErr)
		}
		return nil
	})
}

func (s *InvestigationService) ReconcileInvestigation(ctx context.Context, investigationID uuid.UUID) error {
	if investigationID == uuid.Nil {
		return fmt.Errorf("%w: investigation ID is required", rez.ErrInvalidInput)
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

		evidenceRevisionQuery := tx.InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(investigationID), inver.AgentTurnIDIsNil()).
			Order(inver.ByCreatedAt(sql.OrderAsc()), inver.ByID(sql.OrderAsc()))
		evidenceRevision, evidenceRevisionErr := evidenceRevisionQuery.First(ctx)
		if evidenceRevisionErr != nil && !ent.IsNotFound(evidenceRevisionErr) {
			return fmt.Errorf("select next investigation evidence revision: %w", evidenceRevisionErr)
		}
		if ent.IsNotFound(userInputErr) && ent.IsNotFound(evidenceRevisionErr) {
			return nil
		}

		messageSections := make([]string, 0, 2)
		if userInputErr == nil {
			messageSections = append(messageSections, "User question:\n"+userInput.Text)
		}
		if evidenceRevisionErr == nil {
			messageSections = append(messageSections, "Evidence revised:\n"+evidenceRevision.Explanation)
		}
		turnInput := &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(strings.Join(messageSections, "\n\n"))}
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
			return fmt.Errorf("%w: requested turn belongs to another investigation", rez.ErrConflict)
		}

		if userInputErr == nil {
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
				return fmt.Errorf("%w: investigation user input was assigned concurrently", rez.ErrConflict)
			}
		}
		if evidenceRevisionErr == nil {
			assignRevision := tx.InvestigationEvidenceRevision.Update().
				Where(
					inver.ID(evidenceRevision.ID),
					inver.InvestigationID(investigationID),
					inver.AgentTurnIDIsNil(),
				).
				SetAgentTurnID(turn.ID)
			updated, updateErr := assignRevision.Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("assign investigation evidence revision to turn: %w", updateErr)
			}
			if updated != 1 {
				return fmt.Errorf("%w: investigation evidence revision was assigned concurrently", rez.ErrConflict)
			}
		}
		return nil
	})
}
