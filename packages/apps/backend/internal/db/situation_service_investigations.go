package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	agt "github.com/rezible/rezible/ent/agentturn"
	ale "github.com/rezible/rezible/ent/alertepisode"
	"github.com/rezible/rezible/ent/situation"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
)

func (s *SituationService) CreateSituationInvestigation(ctx context.Context, params rez.CreateSituationInvestigationParams) (*ent.SituationInvestigation, error) {
	if params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}

	var investigation *ent.SituationInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		querySituation := tx.Situation.Query().
			Where(situation.IDEQ(params.SituationID)).
			WithInvestigation()
		current, queryErr := querySituation.Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("get situation: %w", queryErr)
		}

		if current.Edges.Investigation != nil {
			investigation = current.Edges.Investigation.Unwrap()
			return nil
		}

		createAnalysis := tx.SystemAnalysis.Create().
			SetSubjectEntityID(current.KnowledgeEntityID)
		analysis, analysisErr := createAnalysis.Save(ctx)
		if analysisErr != nil {
			return fmt.Errorf("create system analysis: %w", analysisErr)
		}
		createAnalysisEntity := tx.SystemAnalysisEntity.Create().
			SetAnalysisID(analysis.ID).
			SetKnowledgeEntityID(current.KnowledgeEntityID)
		if entityErr := createAnalysisEntity.Exec(ctx); entityErr != nil {
			return fmt.Errorf("seed system analysis entity: %w", entityErr)
		}

		sessionParams := rez.CreateAgentSessionParams{
			AgentName:        rezai.InvestigationAgent.Name,
			Input:            rezai.InvestigationAgentInput{SituationID: params.SituationID},
			SystemAnalysisID: &analysis.ID,
		}
		session, sessionErr := s.agents.CreateAgentSession(ctx, sessionParams)
		if sessionErr != nil {
			return fmt.Errorf("create situation investigation agent session: %w", sessionErr)
		}

		createInvestigation := tx.SituationInvestigation.Create().
			SetSituationID(params.SituationID).
			SetSystemAnalysisID(analysis.ID).
			SetAgentSessionID(session.ID)
		created, investigationErr := createInvestigation.Save(ctx)
		if investigationErr != nil {
			return fmt.Errorf("create situation investigation: %w", investigationErr)
		}

		if queueJobErr := s.requestSituationInvestigation(ctx, params.SituationID); queueJobErr != nil {
			return fmt.Errorf("queue investigation job: %w", queueJobErr)
		}

		investigation = created.Unwrap()
		return nil
	})
}

func (s *SituationService) GetSituationInvestigation(ctx context.Context, situationID uuid.UUID) (*ent.SituationInvestigation, error) {
	return s.db.Client(ctx).SituationInvestigation.Query().
		Where(siti.SituationID(situationID)).
		WithSituation().
		WithSystemAnalysis().
		WithAgentSession().
		Only(ctx)
}

func (s *SituationService) RecordSituationEvidence(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if situationLockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, id.String()); situationLockErr != nil {
			return fmt.Errorf("get situation lock: %w", situationLockErr)
		}
		updateRevision := tx.Situation.UpdateOneID(id).
			AddEvidenceRevision(1)
		if updateRevisionErr := updateRevision.Exec(ctx); updateRevisionErr != nil {
			return fmt.Errorf("update evidence revision: %w", updateRevisionErr)
		}

		lookupInvestigation := tx.SituationInvestigation.Query().Where(siti.SituationID(id))
		investigationExists, queryInvestigationErr := lookupInvestigation.Exist(ctx)
		if queryInvestigationErr != nil {
			return fmt.Errorf("lookup investigation: %w", queryInvestigationErr)
		} else if investigationExists {
			return s.requestSituationInvestigation(ctx, id)
		}

		return nil
	})
}

func (s *SituationService) SetSituationInvestigationReport(ctx context.Context, params rez.SetSituationInvestigationReportParams) (*ent.SituationInvestigation, error) {
	var result *ent.SituationInvestigation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		lookupByTurn := tx.SituationInvestigation.Query().
			Where(siti.RequestedTurnID(params.AgentTurnID))
		turnInv, lookupByTurnErr := lookupByTurn.Only(ctx)
		if lookupByTurnErr != nil {
			if ent.IsNotFound(lookupByTurnErr) {
				return fmt.Errorf("%w: stale investigation turn", rez.ErrConflict)
			}
			return fmt.Errorf("lookup investigation for turn: %w", lookupByTurnErr)
		}

		if situLockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, turnInv.SituationID.String()); situLockErr != nil {
			return situLockErr
		}

		if agentTurnLockErr := acquireAgentSessionTurnLock(ctx, s.db, turnInv.AgentSessionID); agentTurnLockErr != nil {
			return agentTurnLockErr
		}

		lookupInv := tx.SituationInvestigation.Query().
			Where(siti.ID(turnInv.ID)).
			WithRequestedTurn()
		inv, lookupInvestigationErr := lookupInv.Only(ctx)
		if lookupInvestigationErr != nil {
			return fmt.Errorf("lookup investigation: %w", lookupInvestigationErr)
		}
		if inv.RequestedTurnID == nil || *inv.RequestedTurnID != params.AgentTurnID {
			return fmt.Errorf("%w: stale investigation turn", rez.ErrConflict)
		}

		if inv.CompletedRevision == inv.RequestedRevision {
			result = inv.Unwrap()
			return nil
		}

		turn, turnErr := inv.Edges.RequestedTurnOrErr()
		if turnErr != nil {
			return fmt.Errorf("get requested turn: %w", turnErr)
		}

		if turn.AgentSessionID != inv.AgentSessionID || turn.Status != agt.StatusRunning {
			return fmt.Errorf("%w: report turn must be running", rez.ErrConflict)
		}
		params.Report.Text = strings.TrimSpace(params.Report.Text)
		if params.Report.Text == "" {
			return fmt.Errorf("%w: report text is required", rez.ErrInvalidInput)
		}

		updateInv := inv.Update().
			SetReport(params.Report).
			SetCompletedRevision(inv.RequestedRevision)
		updatedInv, updateInvErr := updateInv.Save(ctx)
		if updateInvErr != nil {
			return updateInvErr
		}

		for _, assessment := range params.Assessments {
			addAssessmentParams := rez.AddSituationHazardAssessmentParams{
				SituationID:    inv.SituationID,
				SystemHazardID: assessment.SystemHazardID,
				Status:         assessment.Status,
				Summary:        assessment.Summary,
				AgentTurnID:    &turn.ID,
			}
			if _, addAssmtErr := s.AddSituationHazardAssessment(ctx, addAssessmentParams); addAssmtErr != nil {
				return fmt.Errorf("add situation hazard assessment: %w", addAssmtErr)
			}
		}

		result = updatedInv.Unwrap()
		return nil
	})
}

func (s *SituationService) onAgentTurnUpdated(ctx context.Context, event *rezai.AgentTurnUpdated) error {
	if event.Status != agt.StatusCompleted && event.Status != agt.StatusFailed && event.Status != agt.StatusAborted {
		return nil
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		query := tx.SituationInvestigation.Query().
			Where(siti.AgentSessionID(event.AgentSessionId)).
			WithSituation()
		inv, lookupInvestigationErr := query.Only(ctx)
		if lookupInvestigationErr != nil {
			if ent.IsNotFound(lookupInvestigationErr) {
				return nil
			}
			return fmt.Errorf("lookup investigation: %w", lookupInvestigationErr)
		}
		if inv.Edges.Situation.EvidenceRevision > inv.RequestedRevision {
			return s.requestSituationInvestigation(ctx, inv.SituationID)
		}
		return nil
	})
}

func (s *SituationService) requestSituationInvestigation(ctx context.Context, id uuid.UUID) error {
	_, insertErr := s.jobs.Insert(ctx, jobs.ReconcileSituationInvestigation{SituationID: id}, nil)
	return insertErr
}

func NewReconcileSituationInvestigationWorker(db rez.Database, s rez.SituationService, agents rez.AgentSessionService) *ReconcileSituationInvestigationWorker {
	return &ReconcileSituationInvestigationWorker{db: db, situations: s, agents: agents}
}

type ReconcileSituationInvestigationWorker struct {
	jobs.Worker[jobs.ReconcileSituationInvestigation]
	db         rez.Database
	situations rez.SituationService
	agents     rez.AgentSessionService
}

func (w *ReconcileSituationInvestigationWorker) Work(ctx context.Context, job *jobs.Job[jobs.ReconcileSituationInvestigation]) error {
	situationId := job.Args.SituationID
	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// Situation precedes session everywhere that scheduling and submission meet.
		if situationLockErr := w.db.AcquireTxLocks(ctx, situationLockNamespace, situationId.String()); situationLockErr != nil {
			return situationLockErr
		}

		query := tx.SituationInvestigation.Query().
			Where(siti.SituationID(situationId)).
			WithSituation()
		inv, queryInvestigationErr := query.Only(ctx)
		if queryInvestigationErr != nil {
			if ent.IsNotFound(queryInvestigationErr) {
				return nil
			}
			return fmt.Errorf("lookup investigation: %w", queryInvestigationErr)
		}

		sit := inv.Edges.Situation
		if sit == nil || inv.CompletedRevision >= sit.EvidenceRevision {
			return nil
		}

		if sessionLockErr := acquireAgentSessionTurnLock(ctx, w.db, inv.AgentSessionID); sessionLockErr != nil {
			return sessionLockErr
		}

		activeTurnQuery := tx.AgentTurn.Query().
			Where(agt.AgentSessionID(inv.AgentSessionID), agt.StatusIn(agt.StatusQueued, agt.StatusRunning))
		active, queryActiveTurnErr := activeTurnQuery.Exist(ctx)
		if queryActiveTurnErr != nil {
			return fmt.Errorf("lookup active agent turn: %w", queryActiveTurnErr)
		}
		if active {
			return nil
		}

		if inv.RequestedTurnID != nil && inv.RequestedRevision == sit.EvidenceRevision {
			reqTurn, lookupRequestedTurnErr := tx.AgentTurn.Get(ctx, *inv.RequestedTurnID)
			if lookupRequestedTurnErr != nil {
				return fmt.Errorf("lookup requested turn: %w", lookupRequestedTurnErr)
			}
			if reqTurn.Status == agt.StatusFailed || reqTurn.Status == agt.StatusAborted {
				return nil
			}
		}

		episodesQuery := tx.AlertEpisode.Query().
			Where(ale.SituationID(situationId)).
			WithAlertDefinition().
			WithInstances(func(q *ent.AlertInstanceQuery) {
				q.WithEvent()
			}).
			Order(ale.ByStartedAt(), ale.ByID())
		episodes, queryEpisodesErr := episodesQuery.All(ctx)
		if queryEpisodesErr != nil {
			return fmt.Errorf("load investigation evidence: %w", queryEpisodesErr)
		}

		evidence, jsonErr := json.Marshal(episodes)
		if jsonErr != nil {
			return fmt.Errorf("marshal episodes: %w", jsonErr)
		}
		msgText := fmt.Sprintf("Investigate this situation and save its report with save_situation_investigation_report.\nTitle: %s\nSummary: %s\nEvidence revision: %d\nAlert evidence: %s",
			sit.Title, sit.Summary, sit.EvidenceRevision, evidence)
		turnParams := &rez.RequestAgentTurnParams{
			Input: &rez.AiAgentTurnInput{
				Message: ai.NewUserTextMessage(msgText),
			},
		}
		turn, requestTurnErr := w.agents.RequestAgentTurn(ctx, inv.AgentSessionID, turnParams)
		if requestTurnErr != nil {
			return fmt.Errorf("request investigation agent turn: %w", requestTurnErr)
		}
		update := inv.Update().
			SetRequestedTurnID(turn.ID).
			SetRequestedRevision(sit.EvidenceRevision)
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update investigation: %w", updateErr)
		}

		return nil
	})
}
