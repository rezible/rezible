package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/projections"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ald "github.com/rezible/rezible/ent/alertdefinition"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type AlertService struct {
	db         rez.Database
	situations rez.SituationService
	knowledge  rez.KnowledgeGraphService
}

func NewAlertService(db rez.Database, situations rez.SituationService, knowledge rez.KnowledgeGraphService) (*AlertService, error) {
	s := &AlertService{db: db, situations: situations, knowledge: knowledge}

	return s, nil
}

func (s *AlertService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) (*ent.ListResult[ent.AlertDefinition], error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		Order(ald.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(ald.TitleContainsFold(search))
	}
	return ent.DoListQuery[ent.AlertDefinition, *ent.AlertDefinitionQuery](ctx, query, params.ListParams)
}

func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*ent.AlertDefinition, error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		Where(ald.ID(id))
	return query.Only(ctx)
}

func (s *AlertService) GetAlertInstance(ctx context.Context, id uuid.UUID) (*ent.AlertInstance, error) {
	query := s.db.Client(ctx).AlertInstance.Query().
		Where(ali.ID(id)).
		WithEpisode(func(query *ent.AlertEpisodeQuery) {
			query.WithAlertDefinition()
		})
	return query.Only(ctx)
}

func (s *AlertService) GetAlertMetrics(ctx context.Context, params rez.GetAlertMetricsParams) (*ent.AlertMetrics, error) {
	return &ent.AlertMetrics{}, nil
}

const alertEpisodeInactivity = 30 * time.Minute
const alertDefinitionLockNamespace = "alert_definition"

func (s *AlertService) RecordAlertDefinitionInstance(ctx context.Context, definitionID uuid.UUID, event *ent.NormalizedEvent) (*ent.AlertInstance, error) {
	var result *ent.AlertInstance
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if locksErr := s.db.AcquireTxLocks(ctx, alertDefinitionLockNamespace, definitionID.String()); locksErr != nil {
			return fmt.Errorf("get locks: %w", locksErr)
		}
		existingQuery := tx.AlertInstance.Query().
			Where(ali.NormalizedEventID(event.ID))
		existing, lookupExistingErr := existingQuery.Only(ctx)
		if lookupExistingErr != nil && !ent.IsNotFound(lookupExistingErr) {
			return fmt.Errorf("lookup existing definition: %w", lookupExistingErr)
		}
		if existing != nil && lookupExistingErr == nil {
			result = existing.Unwrap()
			return nil
		}

		episode, episodeErr := s.resolveAlertEpisode(ctx, definitionID, event.OccurredAt)
		if episodeErr != nil {
			return fmt.Errorf("resolve alert episode: %w", episodeErr)
		}

		createInstance := tx.AlertInstance.Create().
			SetAlertEpisodeID(episode.ID).
			SetNormalizedEventID(event.ID)
		instance, saveErr := createInstance.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save alert instance: %w", saveErr)
		}

		result = instance.Unwrap()
		return nil
	})
}

func (s *AlertService) resolveAlertEpisode(ctx context.Context, definitionID uuid.UUID, occurredAt time.Time) (*ent.AlertEpisode, error) {
	var result *ent.AlertEpisode
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if locksErr := s.db.AcquireTxLocks(ctx, alertDefinitionLockNamespace, definitionID.String()); locksErr != nil {
			return fmt.Errorf("get locks: %w", locksErr)
		}
		queryAlertDefinition := tx.AlertDefinition.Query().
			Where(ald.ID(definitionID)).
			WithEpisodes(func(q *ent.AlertEpisodeQuery) {
				q.Where(ale.StatusEQ(ale.StatusOpen))
			})
		definition, queryDefinitionErr := queryAlertDefinition.Only(ctx)
		if queryDefinitionErr != nil {
			return fmt.Errorf("query alert definition: %w", queryDefinitionErr)
		}

		var episode *ent.AlertEpisode
		if len(definition.Edges.Episodes) > 0 {
			episode = definition.Edges.Episodes[0]
			if episodeLockErr := s.db.AcquireTxLocks(ctx, alertEpisodeLockNamespace, episode.ID.String()); episodeLockErr != nil {
				return fmt.Errorf("get lock for episode: %w", episodeLockErr)
			}
			boundary := episode.LastObservedAt.Add(alertEpisodeInactivity)
			if occurredAt.After(boundary) {
				prevSituationID := episode.SituationID
				setClosed := episode.Update().
					SetStatus(ale.StatusClosed).
					SetClosedAt(boundary)
				if updateEpisodeErr := setClosed.Exec(ctx); updateEpisodeErr != nil {
					return fmt.Errorf("failed to update existing episode: %w", updateEpisodeErr)
				}
				if prevSituationID != nil {
					params := rez.SituationEvidenceItemParams{AlertEpisodeID: &episode.ID}
					if situationErr := s.situations.RemoveSituationEvidenceItem(ctx, *prevSituationID, params); situationErr != nil {
						return fmt.Errorf("remove alert episode situation link: %w", situationErr)
					}
				}
				episode = nil
			}
		}

		if episode != nil {
			update := episode.Update()
			if occurredAt.Before(episode.StartedAt) {
				update.SetStartedAt(occurredAt)
			}
			if occurredAt.After(episode.LastObservedAt) {
				update.SetLastObservedAt(occurredAt)
			}
			if updateEpisodeErr := update.Exec(ctx); updateEpisodeErr != nil {
				return fmt.Errorf("update episode: %w", updateEpisodeErr)
			}

			if sitId := episode.SituationID; sitId != nil {
				params := rez.SituationEvidenceItemParams{AlertEpisodeID: &episode.ID}
				if evidenceErr := s.situations.NotifySituationEvidenceItemUpdated(ctx, *sitId, params); evidenceErr != nil {
					return fmt.Errorf("situation evidence: %w", evidenceErr)
				}
			}
		} else {
			episodeId := uuid.New()
			episodeEntityRef := &rez.KnowledgeEntityRef{
				Category:            kne.CategoryEvent,
				Kind:                "alert_episode",
				ProviderResourceRef: projections.InternalEntityResourceRef(episodeId),
			}
			ka, kaErr := s.knowledge.ResolveInternalSubject(ctx, rez.KnowledgeSubjectRef{Entity: episodeEntityRef})
			if kaErr != nil || ka.EntityID == nil {
				return fmt.Errorf("create alert episode knowledge entity: %w", kaErr)
			}
			create := tx.AlertEpisode.Create().
				SetAlertDefinitionID(definitionID).
				SetKnowledgeEntityID(*ka.EntityID).
				SetStartedAt(occurredAt).
				SetLastObservedAt(occurredAt)
			createdEpisode, createEpisodeErr := create.Save(ctx)
			if createEpisodeErr != nil {
				return createEpisodeErr
			}
			episode = createdEpisode

			situationEvidenceItem := rez.SituationEvidenceItemParams{
				AlertEpisodeID: &episodeId,
			}
			params := rez.CreateSituationParams{
				Title:         definition.Title,
				OpenedAt:      occurredAt,
				EvidenceItems: []rez.SituationEvidenceItemParams{situationEvidenceItem},
			}
			sit, situationErr := s.situations.CreateSituation(ctx, params)
			if situationErr != nil {
				return fmt.Errorf("create situation: %w", situationErr)
			}
			episode.SituationID = &sit.ID
		}

		result = episode.Unwrap()
		return nil
	})
}

func NewCloseInactiveAlertEpisodesWorker(db rez.Database, situations rez.SituationService) (*CloseInactiveAlertEpisodesWorker, error) {
	return &CloseInactiveAlertEpisodesWorker{db: db, situations: situations}, nil
}

type CloseInactiveAlertEpisodesWorker struct {
	jobs.WorkerDefaults[jobs.CloseInactiveAlertEpisodes]
	db         rez.Database
	situations rez.SituationService
}

func (w *CloseInactiveAlertEpisodesWorker) Work(ctx context.Context, job *jobs.Job[jobs.CloseInactiveAlertEpisodes]) error {
	systemCtx := execution.NewSystemContext(ctx)
	now := time.Now().UTC()
	query := w.db.Client(systemCtx).AlertEpisode.Query().
		Where(ale.StatusEQ(ale.StatusOpen), ale.LastObservedAtLT(now.Add(-alertEpisodeInactivity)))
	openEpisodes, queryOpenErr := query.All(systemCtx)
	if queryOpenErr != nil {
		return fmt.Errorf("querying open alert episodes: %w", queryOpenErr)
	}

	for _, candidate := range openEpisodes {
		tenantCtx := execution.NewTenantContext(ctx, candidate.TenantID)
		if closeErr := w.maybeCloseOpenEpisode(tenantCtx, now, candidate.ID); closeErr != nil {
			return fmt.Errorf("close inactive episode %s: %w", candidate.ID, closeErr)
		}
	}

	return nil
}

func (w *CloseInactiveAlertEpisodesWorker) maybeCloseOpenEpisode(ctx context.Context, now time.Time, id uuid.UUID) error {
	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if epLock := w.db.AcquireTxLocks(ctx, alertEpisodeLockNamespace, id.String()); epLock != nil {
			return fmt.Errorf("alert episode lock: %w", epLock)
		}
		queryCandidate := tx.AlertEpisode.Query().
			Where(ale.ID(id), ale.StatusEQ(ale.StatusOpen), ale.LastObservedAtGT(now.Add(-alertEpisodeInactivity)))
		shouldClose, queryCandidateErr := queryCandidate.Exist(ctx)
		if queryCandidateErr != nil {
			return fmt.Errorf("lookup alert episode: %w", queryCandidateErr)
		}
		if !shouldClose {
			return nil
		}
		update := tx.AlertEpisode.UpdateOneID(id).
			SetStatus(ale.StatusClosed).
			SetClosedAt(now)
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("failed to update episode %s: %w", id, updateErr)
		}
		if sitId := updated.SituationID; sitId != nil {
			params := rez.SituationEvidenceItemParams{AlertEpisodeID: &id}
			if evidenceErr := w.situations.NotifySituationEvidenceItemUpdated(ctx, *sitId, params); evidenceErr != nil {
				return fmt.Errorf("situation evidence: %w", evidenceErr)
			}
		}
		return nil
	})
}
