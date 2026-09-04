package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentartifact"
	"github.com/rezible/rezible/ent/alertepisode"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/ent/situationinvestigation"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/messages"
)

const (
	alertEpisodeLockNamespace              = "alert_episode"
	situationHazardAssessmentLockNamespace = "situation_hazard_assessment"
	situationKnowledgeEntityKind           = "situation"
	situationLockNamespace                 = "situation"
)

type SituationService struct {
	db     rez.Database
	agents rez.AgentSessionService
}

func NewSituationService(db rez.Database, msgs rez.MessageService, agents rez.AgentSessionService) (*SituationService, error) {
	s := &SituationService{db: db, agents: agents}

	if handlersErr := s.addMessageHandlers(msgs); handlersErr != nil {
		return nil, fmt.Errorf("adding handlers: %w", handlersErr)
	}

	return s, nil
}

func (s *SituationService) addMessageHandlers(msgs rez.MessageService) error {
	return msgs.AddHandlers(
		messages.NewEventHandler("db.SituationService.onAgentTurnFinished", s.onAgentTurnFinished))
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.db.Client(ctx).Situation.Query().
		WithKnowledgeEntity().
		WithAlertEpisodes().
		WithInvestigation(func(query *ent.SituationInvestigationQuery) {
			query.WithSystemAnalysis().WithAgentSession()
		}).
		Order(situation.ByOpenedAt(params.GetOrder()), situation.ByID(params.GetOrder()))
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, id uuid.UUID) (*ent.Situation, error) {
	return s.db.Client(ctx).Situation.Query().
		Where(situation.ID(id)).
		WithKnowledgeEntity().
		WithAlertEpisodes().
		WithInvestigation(func(query *ent.SituationInvestigationQuery) {
			query.WithSystemAnalysis().WithAgentSession()
		}).
		Only(ctx)
}

func (s *SituationService) CreateSituation(ctx context.Context, params rez.CreateSituationParams) (*ent.Situation, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: situation title is required", rez.ErrInvalidInput)
	}
	openedAt := params.OpenedAt
	if openedAt.IsZero() {
		openedAt = time.Now().UTC()
	}

	var created *ent.Situation
	return created, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		knowledgeEntity, entityErr := tx.KnowledgeEntity.Create().
			SetCategory(kne.CategoryEvent).
			SetKind(situationKnowledgeEntityKind).
			Save(ctx)
		if entityErr != nil {
			return fmt.Errorf("create situation knowledge entity: %w", entityErr)
		}

		createSituation := tx.Situation.Create().
			SetKnowledgeEntityID(knowledgeEntity.ID).
			SetTitle(title).
			SetStatus(situation.StatusOpen).
			SetOpenedAt(openedAt)
		if summary := strings.TrimSpace(params.Summary); summary != "" {
			createSituation.SetSummary(summary)
		}
		createdSituation, saveErr := createSituation.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create situation: %w", saveErr)
		}
		created = createdSituation.Unwrap()
		return nil
	})
}

func isValidSituationCloseReason(reason situation.CloseReason) bool {
	return reason == situation.CloseReasonStabilized || reason == situation.CloseReasonDismissed
}

func (s *SituationService) CloseSituation(ctx context.Context, params rez.CloseSituationParams) (*ent.Situation, error) {
	if params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}
	if !isValidSituationCloseReason(params.Reason) {
		return nil, fmt.Errorf("%w: invalid situation close reason", rez.ErrInvalidInput)
	}

	var closed *ent.Situation
	return closed, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		current, queryErr := tx.Situation.Query().Where(situation.ID(params.SituationID)).Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("get situation: %w", queryErr)
		}
		if current.Status == situation.StatusClosed {
			if current.CloseReason != nil && *current.CloseReason == params.Reason {
				closed = current.Unwrap()
				return nil
			}
			return fmt.Errorf("%w: situation is already closed with another reason", rez.ErrConflict)
		}

		updated, updateErr := tx.Situation.UpdateOneID(params.SituationID).
			SetStatus(situation.StatusClosed).
			SetCloseReason(params.Reason).
			SetClosedAt(time.Now().UTC()).
			Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("close situation: %w", updateErr)
		}
		closed = updated.Unwrap()
		return nil
	})
}

func (s *SituationService) LinkAlertEpisodeToSituation(ctx context.Context, params rez.LinkAlertEpisodeToSituationParams) (*ent.AlertEpisode, error) {
	if params.AlertEpisodeID == uuid.Nil || params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: alert episode and situation ids are required", rez.ErrInvalidInput)
	}

	var linked *ent.AlertEpisode
	return linked, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, alertEpisodeLockNamespace, params.AlertEpisodeID.String()); lockErr != nil {
			return fmt.Errorf("lock alert episode: %w", lockErr)
		}

		episode, queryErr := tx.AlertEpisode.Query().Where(alertepisode.ID(params.AlertEpisodeID)).Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("get alert episode: %w", queryErr)
		}
		if episode.SituationID != nil {
			if *episode.SituationID == params.SituationID {
				linked = episode.Unwrap()
				return nil
			}
			return fmt.Errorf("%w: alert episode is already linked to another situation", rez.ErrConflict)
		}

		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}
		linkedSituation, situationErr := tx.Situation.Query().Where(situation.ID(params.SituationID)).Only(ctx)
		if situationErr != nil {
			return fmt.Errorf("get situation: %w", situationErr)
		}
		if linkedSituation.Status != situation.StatusOpen {
			return fmt.Errorf("%w: alert episode links may target only an open situation", rez.ErrConflict)
		}

		updated, updateErr := tx.AlertEpisode.UpdateOneID(params.AlertEpisodeID).
			SetSituationID(params.SituationID).
			Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("link alert episode to situation: %w", updateErr)
		}
		linked = updated.Unwrap()
		return nil
	})
}

func (s *SituationService) CreateSituationInvestigation(ctx context.Context, params rez.CreateSituationInvestigationParams) (*ent.SituationInvestigation, error) {
	if params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}

	var investigation *ent.SituationInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		current, queryErr := tx.Situation.Query().Where(situation.ID(params.SituationID)).Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("get situation: %w", queryErr)
		}
		if currentInvestigation, currentErr := tx.SituationInvestigation.Query().
			Where(situationinvestigation.SituationID(params.SituationID)).Only(ctx); currentErr == nil {
			investigation = currentInvestigation.Unwrap()
			return nil
		} else if !ent.IsNotFound(currentErr) {
			return fmt.Errorf("get existing situation investigation: %w", currentErr)
		}

		analysis, analysisErr := tx.SystemAnalysis.Create().
			SetSubjectEntityID(current.KnowledgeEntityID).
			Save(ctx)
		if analysisErr != nil {
			return fmt.Errorf("create system analysis: %w", analysisErr)
		}
		if _, entityErr := tx.SystemAnalysisEntity.Create().
			SetAnalysisID(analysis.ID).
			SetKnowledgeEntityID(current.KnowledgeEntityID).
			Save(ctx); entityErr != nil {
			return fmt.Errorf("seed system analysis entity: %w", entityErr)
		}

		session, sessionErr := s.agents.CreateAgentSession(ctx, rez.CreateAgentSessionParams{
			AgentName:        rezai.InvestigationAgent.Name,
			PermissionScopes: nil,
			Input:            rezai.InvestigationAgentInput{SituationID: params.SituationID},
			SystemAnalysisID: &analysis.ID,
			Metadata:         nil,
		})
		if sessionErr != nil {
			return fmt.Errorf("create situation investigation agent session: %w", sessionErr)
		}

		created, investigationErr := tx.SituationInvestigation.Create().
			SetSituationID(params.SituationID).
			SetSystemAnalysisID(analysis.ID).
			SetAgentSessionID(session.ID).
			Save(ctx)
		if investigationErr != nil {
			return fmt.Errorf("create situation investigation: %w", investigationErr)
		}
		investigation = created.Unwrap()
		return nil
	})
}

func (s *SituationService) GetSituationInvestigation(ctx context.Context, situationID uuid.UUID) (*ent.SituationInvestigation, error) {
	return s.db.Client(ctx).SituationInvestigation.Query().
		Where(situationinvestigation.SituationID(situationID)).
		WithSituation().
		WithSystemAnalysis().
		WithAgentSession().
		Only(ctx)
}

func (s *SituationService) onAgentTurnFinished(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	sessionInvestigationQuery := s.db.Client(ctx).SituationInvestigation.Query().
		Where(situationinvestigation.AgentSessionID(ev.AgentSessionId))
	exists, existsErr := sessionInvestigationQuery.Exist(ctx)
	if existsErr != nil {
		return fmt.Errorf("query situation investigation: %w", existsErr)
	}
	if !exists {
		return nil
	}

	lookupReportArtifactQuery := s.db.Client(ctx).AgentArtifact.Query().
		Where(
			agentartifact.AgentSessionID(ev.AgentSessionId),
			agentartifact.Name("situation_investigation_report"),
		)
	artifact, artifactErr := lookupReportArtifactQuery.Only(ctx)
	if artifactErr != nil {
		if ent.IsNotFound(artifactErr) {
			return nil
		}
		return fmt.Errorf("query situation investigation report artifact: %w", artifactErr)
	}

	report, reportErr := s.situationInvestigationReportFromArtifact(artifact.Parts)
	if reportErr != nil {
		return reportErr
	}
	_, saveErr := s.SetSituationInvestigationReport(ctx, rez.SetSituationInvestigationReportParams{
		AgentSessionID: ev.AgentSessionId,
		Report:         report,
	})
	return saveErr
}

func (s *SituationService) situationInvestigationReportFromArtifact(parts []*ai.Part) (schematypes.SituationInvestigationReport, error) {
	var report schematypes.SituationInvestigationReport
	for _, part := range parts {
		text := strings.TrimSpace(part.Text)
		if !strings.HasPrefix(text, "{") {
			continue
		}
		if jsonErr := json.Unmarshal([]byte(text), &report); jsonErr != nil {
			return report, fmt.Errorf("parse situation investigation report artifact: %w", jsonErr)
		}
		return report, nil
	}
	return report, fmt.Errorf("%w: situation investigation report artifact has no JSON part", rez.ErrInvalidInput)
}

func (s *SituationService) SetSituationInvestigationReport(ctx context.Context, params rez.SetSituationInvestigationReportParams) (*ent.SituationInvestigation, error) {
	params.Report.Text = strings.TrimSpace(params.Report.Text)
	if params.AgentSessionID == uuid.Nil {
		return nil, fmt.Errorf("%w: agent session id is required", rez.ErrInvalidInput)
	}
	if params.Report.Text == "" {
		return nil, fmt.Errorf("%w: report text is empty", rez.ErrInvalidInput)
	}

	var investigation *ent.SituationInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		current, queryErr := tx.SituationInvestigation.Query().
			Where(situationinvestigation.AgentSessionID(params.AgentSessionID)).
			Only(ctx)
		if queryErr != nil {
			return queryErr
		}
		updated, updateErr := current.Update().SetReport(params.Report).Save(ctx)
		if updateErr != nil {
			return updateErr
		}
		investigation = updated.Unwrap()
		return nil
	})
}

func (s *SituationService) ListSituationHazardAssessments(ctx context.Context, params rez.ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error) {
	query := s.db.Client(ctx).SituationHazardAssessment.Query().
		Order(
			situationhazardassessment.ByAssessedAt(params.GetOrder()),
			situationhazardassessment.ByRevision(params.GetOrder()),
			situationhazardassessment.ByID(params.GetOrder()),
		)
	if params.SituationID != uuid.Nil {
		query.Where(situationhazardassessment.SituationID(params.SituationID))
	}
	if params.SystemHazardID != uuid.Nil {
		query.Where(situationhazardassessment.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SituationHazardAssessment, *ent.SituationHazardAssessmentQuery](ctx, query, params.ListParams)
}

func (s *SituationService) AddSituationHazardAssessment(ctx context.Context, params rez.AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error) {
	if params.SituationID == uuid.Nil || params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation and system hazard ids are required", rez.ErrInvalidInput)
	}
	if (params.UserID == nil) == (params.AgentTurnID == nil) {
		return nil, fmt.Errorf("%w: exactly one assessor is required", rez.ErrInvalidInput)
	}
	if !s.isValidSituationHazardAssessmentStatus(params.Status) {
		return nil, fmt.Errorf("%w: invalid situation hazard assessment status", rez.ErrInvalidInput)
	}
	summary := strings.TrimSpace(params.Summary)
	if summary == "" {
		return nil, fmt.Errorf("%w: assessment summary is required", rez.ErrInvalidInput)
	}
	assessedAt := params.AssessedAt
	if assessedAt.IsZero() {
		assessedAt = time.Now().UTC()
	}

	var assessment *ent.SituationHazardAssessment
	return assessment, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		pairKey := params.SituationID.String() + "\x1f" + params.SystemHazardID.String()
		if lockErr := s.db.AcquireTxLocks(ctx, situationHazardAssessmentLockNamespace, pairKey); lockErr != nil {
			return fmt.Errorf("lock situation hazard pair: %w", lockErr)
		}
		if _, situationErr := tx.Situation.Get(ctx, params.SituationID); situationErr != nil {
			return fmt.Errorf("get situation: %w", situationErr)
		}
		if _, hazardErr := tx.SystemHazard.Get(ctx, params.SystemHazardID); hazardErr != nil {
			return fmt.Errorf("get system hazard: %w", hazardErr)
		}
		if params.UserID != nil {
			if _, userErr := tx.User.Get(ctx, *params.UserID); userErr != nil {
				return fmt.Errorf("get assessment user: %w", userErr)
			}
		}
		if params.AgentTurnID != nil {
			if _, turnErr := tx.AgentTurn.Get(ctx, *params.AgentTurnID); turnErr != nil {
				return fmt.Errorf("get assessment agent turn: %w", turnErr)
			}
		}

		nextRevision := 1
		latest, latestErr := tx.SituationHazardAssessment.Query().
			Where(
				situationhazardassessment.SituationID(params.SituationID),
				situationhazardassessment.SystemHazardID(params.SystemHazardID),
			).
			Order(situationhazardassessment.ByRevision(sql.OrderDesc())).
			First(ctx)
		if latestErr == nil {
			nextRevision = latest.Revision + 1
		} else if !ent.IsNotFound(latestErr) {
			return fmt.Errorf("get latest situation hazard assessment: %w", latestErr)
		}

		createAssessment := tx.SituationHazardAssessment.Create().
			SetSituationID(params.SituationID).
			SetSystemHazardID(params.SystemHazardID).
			SetRevision(nextRevision).
			SetStatus(params.Status).
			SetSummary(summary).
			SetAssessedAt(assessedAt)
		createAssessment.SetNillableUserID(params.UserID).SetNillableAgentTurnID(params.AgentTurnID)
		created, saveErr := createAssessment.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("add situation hazard assessment: %w", saveErr)
		}
		assessment = created.Unwrap()
		return nil
	})
}

func (s *SituationService) isValidSituationHazardAssessmentStatus(status situationhazardassessment.Status) bool {
	return status == situationhazardassessment.StatusSuspected ||
		status == situationhazardassessment.StatusConfirmed ||
		status == situationhazardassessment.StatusDisproven
}
