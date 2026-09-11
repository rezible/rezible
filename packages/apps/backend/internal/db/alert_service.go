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
	sog "github.com/rezible/rezible/ent/situationobservationgroup"
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
			Where(ald.ID(definitionID))
		definition, queryDefinitionErr := queryAlertDefinition.Only(ctx)
		if queryDefinitionErr != nil {
			return fmt.Errorf("query alert definition: %w", queryDefinitionErr)
		}

		queryOpenEpisodes := definition.QueryEpisodes().
			Where(ale.StatusEQ(ale.StatusOpen))
		openEpisode, queryOpenEpisodeErr := queryOpenEpisodes.Only(ctx)
		if queryOpenEpisodeErr != nil && !ent.IsNotFound(queryOpenEpisodeErr) {
			return fmt.Errorf("query open episode: %w", queryOpenEpisodeErr)
		}

		episode := openEpisode
		if episode != nil {
			expiryTime := episode.LastObservedAt.Add(alertEpisodeInactivity)
			if occurredAt.After(expiryTime) {
				episode = nil
				if closeErr := s.closeInactiveEpisode(ctx, episode.ID, expiryTime); closeErr != nil {
					return fmt.Errorf("failed to close inactive episode: %w", closeErr)
				}
			}
		}

		if episode != nil {
			isEarlierOccurrance := occurredAt.Before(episode.StartedAt)
			isLaterOccurrance := occurredAt.After(episode.LastObservedAt)
			if !(isEarlierOccurrance || isLaterOccurrance) {
				result = episode.Unwrap()
				return nil
			}

			update := episode.Update()
			if isEarlierOccurrance {
				update.SetStartedAt(occurredAt)
			}
			if isLaterOccurrance {
				update.SetLastObservedAt(occurredAt)
			}
			updated, updateEpisodeErr := update.Save(ctx)
			if updateEpisodeErr != nil {
				return fmt.Errorf("update episode: %w", updateEpisodeErr)
			}

			if sitsErr := s.notifyEpisodeLinkedSituations(ctx, episode.ID); sitsErr != nil {
				return fmt.Errorf("notify existing episode linked situations: %w", sitsErr)
			}

			result = updated.Unwrap()
			return nil
		}

		createEpParams := createAlertEpisodeParams{
			AlertDef:   definition,
			OccurredAt: occurredAt,
		}
		created, createErr := s.createAlertEpisode(ctx, createEpParams)
		if createErr != nil {
			return fmt.Errorf("create alert episode: %w", createErr)
		}
		result = created.Unwrap()
		return nil
	})
}

func (s *AlertService) closeInactiveEpisode(ctx context.Context, id uuid.UUID, inactiveAt time.Time) error {
	return s.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		if episodeLockErr := s.db.AcquireTxLocks(ctx, alertEpisodeLockNamespace, id.String()); episodeLockErr != nil {
			return fmt.Errorf("get lock for episode: %w", episodeLockErr)
		}

		setClosed := client.AlertEpisode.UpdateOneID(id).
			SetStatus(ale.StatusClosed).
			SetClosedAt(inactiveAt)
		if updateEpisodeErr := setClosed.Exec(ctx); updateEpisodeErr != nil {
			return fmt.Errorf("update existing episode: %w", updateEpisodeErr)
		}

		if evidenceErr := s.notifyEpisodeLinkedSituations(ctx, id); evidenceErr != nil {
			return fmt.Errorf("notify linked situations: %w", evidenceErr)
		}

		return nil
	})
}

func (s *AlertService) notifyEpisodeLinkedSituations(ctx context.Context, epId uuid.UUID) error {
	queryEpSituationLinks := s.db.Client(ctx).SituationObservationGroup.Query().
		Where(sog.HasAlertEpisodesWith(ale.ID(epId))).
		WithSituation()
	sitLinks, querySitLinksErr := queryEpSituationLinks.All(ctx)
	if querySitLinksErr != nil {
		return fmt.Errorf("query episode situation links: %w", querySitLinksErr)
	}
	sitEvParams := rez.SituationEvidenceItemParams{AlertEpisodeID: &epId}
	for _, link := range sitLinks {
		if evErr := s.situations.NotifySituationEvidenceItemUpdated(ctx, link.Edges.Situation.ID, sitEvParams); evErr != nil {
			return fmt.Errorf("notify: %w", evErr)
		}
	}
	return nil
}

type createAlertEpisodeParams struct {
	AlertDef   *ent.AlertDefinition
	OccurredAt time.Time
}

func (s *AlertService) createAlertEpisode(ctx context.Context, params createAlertEpisodeParams) (*ent.AlertEpisode, error) {
	var created *ent.AlertEpisode
	return created, s.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		episodeId := uuid.New()

		knowlEntId, knowlErr := s.resolveAlertEpisodeKnowledgeEntityId(ctx, episodeId)
		if knowlErr != nil {
			return fmt.Errorf("resolve alert episode knowledge entity id: %w", knowlErr)
		}

		create := client.AlertEpisode.Create().
			SetID(episodeId).
			SetAlertDefinitionID(params.AlertDef.ID).
			SetKnowledgeEntityID(knowlEntId).
			SetStartedAt(params.OccurredAt).
			SetLastObservedAt(params.OccurredAt)
		createdEpisode, createEpisodeErr := create.Save(ctx)
		if createEpisodeErr != nil {
			return fmt.Errorf("create episode: %w", createEpisodeErr)
		}
		created = createdEpisode

		createSitParams := rez.CreateSituationParams{
			Title:    params.AlertDef.Title,
			OpenedAt: params.OccurredAt,
			EvidenceItems: []rez.SituationEvidenceItemParams{
				{AlertEpisodeID: &episodeId},
			},
		}
		_, situationErr := s.situations.CreateSituation(ctx, createSitParams)
		if situationErr != nil {
			return fmt.Errorf("create situation: %w", situationErr)
		}

		return nil
	})
}

func (s *AlertService) resolveAlertEpisodeKnowledgeEntityId(ctx context.Context, epId uuid.UUID) (uuid.UUID, error) {
	episodeSubjectRef := rez.KnowledgeSubjectRef{
		Entity: &rez.KnowledgeEntityRef{
			Category:            kne.CategoryEvent,
			Kind:                "alert_episode",
			ProviderResourceRef: projections.InternalEntityResourceRef(epId),
		},
	}
	alias, aliasErr := s.knowledge.ResolveInternalSubject(ctx, episodeSubjectRef)
	if aliasErr != nil || alias.EntityID == nil {
		return uuid.Nil, fmt.Errorf("create alert episode knowledge entity: %w", aliasErr)
	}
	return *alias.EntityID, nil
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
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("failed to update episode %s: %w", id, updateErr)
		}

		// TODO: notify situations service that observation group has changed

		return nil
	})
}
