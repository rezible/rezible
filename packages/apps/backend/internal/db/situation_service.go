package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ae "github.com/rezible/rezible/ent/alertepisode"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/situation"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"
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
	jobs   rez.JobService
	agents rez.AgentSessionService
}

func NewSituationService(db rez.Database, msgs rez.MessageService, jobs rez.JobService, agents rez.AgentSessionService) (*SituationService, error) {
	s := &SituationService{db: db, agents: agents, jobs: jobs}

	if handlersErr := s.addMessageHandlers(msgs); handlersErr != nil {
		return nil, fmt.Errorf("adding handlers: %w", handlersErr)
	}

	return s, nil
}

func (s *SituationService) addMessageHandlers(msgs rez.MessageService) error {
	return msgs.AddHandlers(
		messages.NewEventHandler("db.SituationService.onAgentTurnUpdated", s.onAgentTurnUpdated))
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.db.Client(ctx).Situation.Query().
		WithKnowledgeEntity().
		WithAlertEpisodes(func(query *ent.AlertEpisodeQuery) {
			query.WithAlertDefinition()
		}).
		WithInvestigation(func(query *ent.SituationInvestigationQuery) {
			query.WithSystemAnalysis().WithAgentSession()
		}).
		Order(situation.ByOpenedAt(params.GetOrder()), situation.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query = query.Where(situation.TitleContainsFold(search))
	}
	if params.Status != "" {
		query = query.Where(situation.StatusEQ(params.Status))
	}
	if params.OpenedAfter != nil {
		query = query.Where(situation.OpenedAtGTE(*params.OpenedAfter))
	}
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, id uuid.UUID) (*ent.Situation, error) {
	query := s.db.Client(ctx).Situation.Query().
		Where(situation.ID(id)).
		WithKnowledgeEntity().
		WithAlertEpisodes(func(query *ent.AlertEpisodeQuery) {
			query.WithAlertDefinition()
		}).
		WithInvestigation(func(query *ent.SituationInvestigationQuery) {
			query.WithSystemAnalysis().WithAgentSession()
		})
	return query.Only(ctx)
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
		createEntity := tx.KnowledgeEntity.Create().
			SetCategory(kne.CategoryEvent).
			SetKind(situationKnowledgeEntityKind)
		knowledgeEntity, entityErr := createEntity.Save(ctx)
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

		if params.FoundingAlertEpisodeID != nil {
			if locksErr := s.db.AcquireTxLocks(ctx, alertEpisodeLockNamespace, params.FoundingAlertEpisodeID.String()); locksErr != nil {
				return fmt.Errorf("get alert episode lock: %w", locksErr)
			}
			updateEpisode := tx.AlertEpisode.UpdateOneID(*params.FoundingAlertEpisodeID).
				Where(ae.SituationIDIsNil()).
				SetSituationID(createdSituation.ID)
			if updateErr := updateEpisode.Exec(ctx); updateErr != nil {
				return fmt.Errorf("failed to update alert episode: %w", updateErr)
			}
		}

		created = createdSituation.Unwrap()
		return nil
	})
}

func (s *SituationService) CloseSituation(ctx context.Context, params rez.CloseSituationParams) (*ent.Situation, error) {
	if params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}
	if !(params.Reason == situation.CloseReasonStabilized || params.Reason == situation.CloseReasonDismissed) {
		return nil, fmt.Errorf("%w: invalid situation close reason", rez.ErrInvalidInput)
	}

	var closed *ent.Situation
	return closed, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		current, queryErr := tx.Situation.Get(ctx, params.SituationID)
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

		update := tx.Situation.UpdateOneID(params.SituationID).
			SetStatus(situation.StatusClosed).
			SetCloseReason(params.Reason).
			SetClosedAt(time.Now().UTC())
		updated, updateErr := update.Save(ctx)
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

		episode, queryErr := tx.AlertEpisode.Get(ctx, params.AlertEpisodeID)
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
		linkedSituation, situationErr := tx.Situation.Get(ctx, params.SituationID)
		if situationErr != nil {
			return fmt.Errorf("get situation: %w", situationErr)
		}
		if linkedSituation.Status != situation.StatusOpen {
			return fmt.Errorf("%w: alert episode links may target only an open situation", rez.ErrConflict)
		}

		update := tx.AlertEpisode.UpdateOneID(params.AlertEpisodeID).
			SetSituationID(params.SituationID)
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("link alert episode to situation: %w", updateErr)
		}

		if evidenceErr := s.RecordSituationEvidence(ctx, params.SituationID); evidenceErr != nil {
			return fmt.Errorf("situation evidence: %w", evidenceErr)
		}

		linked = updated.Unwrap()
		return nil
	})
}

func (s *SituationService) ListSituationHazardAssessments(ctx context.Context, params rez.ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error) {
	order := params.GetOrder()
	query := s.db.Client(ctx).SituationHazardAssessment.Query().
		Order(sha.ByAssessedAt(order), sha.ByRevision(order), sha.ByID(order))
	if params.SituationID != uuid.Nil {
		query.Where(sha.SituationID(params.SituationID))
	}
	if params.SystemHazardID != uuid.Nil {
		query.Where(sha.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SituationHazardAssessment, *ent.SituationHazardAssessmentQuery](ctx, query, params.ListParams)
}

var validSituationHazardAssessmentStatus = mapset.NewSet(sha.StatusSuspected, sha.StatusConfirmed, sha.StatusDisproven)

func (s *SituationService) AddSituationHazardAssessment(ctx context.Context, params rez.AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error) {
	if params.SituationID == uuid.Nil || params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation and system hazard ids are required", rez.ErrInvalidInput)
	}
	if (params.UserID == nil) == (params.AgentTurnID == nil) {
		return nil, fmt.Errorf("%w: exactly one assessor is required", rez.ErrInvalidInput)
	}
	if !validSituationHazardAssessmentStatus.Contains(params.Status) {
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

		nextRevision := 1
		latestQuery := tx.SituationHazardAssessment.Query().
			Where(sha.SituationID(params.SituationID), sha.SystemHazardID(params.SystemHazardID)).
			Order(sha.ByRevision(sql.OrderDesc()))
		latest, latestErr := latestQuery.First(ctx)
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
			SetAssessedAt(assessedAt).
			SetNillableUserID(params.UserID).
			SetNillableAgentTurnID(params.AgentTurnID)
		created, saveErr := createAssessment.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("add situation hazard assessment: %w", saveErr)
		}
		assessment = created.Unwrap()
		return nil
	})
}

func (s *SituationService) StabilizeSituation(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, id.String()); lockErr != nil {
			return fmt.Errorf("get situation lock: %w", lockErr)
		}

		lookupSituation := tx.Situation.Query().
			Where(situation.ID(id)).
			WithAlertEpisodes()
		sit, situationErr := lookupSituation.Only(ctx)
		if situationErr != nil {
			return fmt.Errorf("get situation: %w", situationErr)
		}
		if sit.Status != situation.StatusOpen {
			return nil
		}

		episodes, episodesErr := sit.Edges.AlertEpisodesOrErr()
		if episodesErr != nil {
			return fmt.Errorf("situation alert episodes: %w", episodesErr)
		}
		var closedAt *time.Time
		for _, ep := range episodes {
			if ep.Status != ae.StatusClosed || ep.ClosedAt == nil {
				continue
			}
			if closedAt == nil || ep.ClosedAt.After(*closedAt) {
				closedAt = ep.ClosedAt
			}
		}
		if closedAt != nil {
			update := tx.Situation.UpdateOneID(id).
				SetStatus(situation.StatusClosed).
				SetCloseReason(situation.CloseReasonStabilized).
				SetClosedAt(*closedAt)
			if updateErr := update.Exec(ctx); updateErr != nil {
				return fmt.Errorf("update situation: %w", updateErr)
			}
		}
		return nil
	})
}
