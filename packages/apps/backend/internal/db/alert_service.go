package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ald "github.com/rezible/rezible/ent/alertdefinition"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	sog "github.com/rezible/rezible/ent/situationobservationgroup"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
)

type AlertService struct {
	db         rez.Database
	situations rez.SituationService
	knowledge  rez.KnowledgeGraphIngestionService
}

func NewAlertService(db rez.Database, situations rez.SituationService, knowledge rez.KnowledgeGraphIngestionService) (*AlertService, error) {
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
const alertEpisodeLockNamespace = "alert_episode"

func (s *AlertService) RecordAlertDefinitionInstance(ctx context.Context, definitionID uuid.UUID, event *ent.NormalizedEvent) (*ent.AlertInstance, error) {
	if definitionID == uuid.Nil || event == nil || event.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: alert definition and normalized event are required", rez.ErrInvalidInput)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.AlertInstance, error) {
		if locksErr := s.db.AcquireTxLocks(ctx, alertDefinitionLockNamespace, definitionID.String()); locksErr != nil {
			return nil, fmt.Errorf("get locks: %w", locksErr)
		}
		existingQuery := tx.AlertInstance.Query().
			Where(ali.NormalizedEventID(event.ID))
		existing, lookupExistingErr := existingQuery.Only(ctx)
		if lookupExistingErr != nil && !ent.IsNotFound(lookupExistingErr) {
			return nil, fmt.Errorf("lookup existing definition: %w", lookupExistingErr)
		}
		if existing != nil && lookupExistingErr == nil {
			return existing, nil
		}

		episode, episodeErr := s.resolveAlertEpisode(ctx, definitionID, event.OccurredAt)
		if episodeErr != nil {
			return nil, fmt.Errorf("resolve alert episode: %w", episodeErr)
		}

		createInstance := tx.AlertInstance.Create().
			SetAlertEpisodeID(episode.ID).
			SetNormalizedEventID(event.ID)
		instance, saveErr := createInstance.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save alert instance: %w", saveErr)
		}
		if refreshErr := s.refreshEpisodeLinkedSituations(ctx, episode.ID); refreshErr != nil {
			return nil, fmt.Errorf("refresh situation analysis after alert instance: %w", refreshErr)
		}
		return instance, nil
	})
}

func (s *AlertService) resolveAlertEpisode(ctx context.Context, definitionID uuid.UUID, occurredAt time.Time) (*ent.AlertEpisode, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.AlertEpisode, error) {
		if locksErr := s.db.AcquireTxLocks(ctx, alertDefinitionLockNamespace, definitionID.String()); locksErr != nil {
			return nil, fmt.Errorf("get locks: %w", locksErr)
		}

		queryAlertDefinition := tx.AlertDefinition.Query().
			Where(ald.ID(definitionID))
		definition, queryDefinitionErr := queryAlertDefinition.Only(ctx)
		if queryDefinitionErr != nil {
			return nil, fmt.Errorf("query alert definition: %w", queryDefinitionErr)
		}

		queryOpenEpisodes := definition.QueryEpisodes().
			Where(ale.StatusEQ(ale.StatusOpen))
		openEpisode, queryOpenEpisodeErr := queryOpenEpisodes.Only(ctx)
		if queryOpenEpisodeErr != nil && !ent.IsNotFound(queryOpenEpisodeErr) {
			return nil, fmt.Errorf("query open episode: %w", queryOpenEpisodeErr)
		}

		episode := openEpisode
		if episode != nil {
			expiryTime := episode.LastObservedAt.Add(alertEpisodeInactivity)
			if occurredAt.After(expiryTime) {
				if closeErr := s.closeInactiveEpisode(ctx, episode.ID, expiryTime); closeErr != nil {
					return nil, fmt.Errorf("failed to close inactive episode: %w", closeErr)
				}
				episode = nil
			}
		}

		if episode != nil {
			isEarlierOccurrance := occurredAt.Before(episode.StartedAt)
			isLaterOccurrance := occurredAt.After(episode.LastObservedAt)
			if !(isEarlierOccurrance || isLaterOccurrance) {
				return episode, nil
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
				return nil, fmt.Errorf("update episode: %w", updateEpisodeErr)
			}

			return updated, nil
		}

		createEpParams := createAlertEpisodeParams{
			AlertDef:   definition,
			OccurredAt: occurredAt,
		}
		created, createErr := s.createAlertEpisode(ctx, createEpParams)
		if createErr != nil {
			return nil, fmt.Errorf("create alert episode: %w", createErr)
		}
		return created, nil
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

		if refreshErr := s.refreshEpisodeLinkedSituations(ctx, id); refreshErr != nil {
			return fmt.Errorf("refresh linked situation analysis: %w", refreshErr)
		}

		return nil
	})
}

func (s *AlertService) CloseInactiveAlertEpisodes(ctx context.Context) error {
	now := time.Now().UTC()
	query := s.db.Client(ctx).AlertEpisode.Query().
		Where(ale.StatusEQ(ale.StatusOpen), ale.LastObservedAtLT(now.Add(-alertEpisodeInactivity)))
	openEpisodes, queryOpenErr := query.All(ctx)
	if queryOpenErr != nil {
		return fmt.Errorf("querying open alert episodes: %w", queryOpenErr)
	}

	for _, candidate := range openEpisodes {
		tenantCtx := execution.NewTenantContext(ctx, candidate.TenantID)
		if closeErr := s.closeInactiveEpisode(tenantCtx, candidate.ID, now); closeErr != nil {
			return fmt.Errorf("close inactive episode %s: %w", candidate.ID, closeErr)
		}
	}

	return nil
}

func (s *AlertService) refreshEpisodeLinkedSituations(ctx context.Context, epId uuid.UUID) error {
	queryEpSituationLinks := s.db.Client(ctx).SituationObservationGroup.Query().
		Where(sog.HasAlertEpisodesWith(ale.ID(epId))).
		WithSituation().
		QuerySituation()
	situationIds, querySituationIdsErr := queryEpSituationLinks.IDs(ctx)
	if querySituationIdsErr != nil {
		return fmt.Errorf("query episode situation links: %w", querySituationIdsErr)
	}
	for situationId := range mapset.NewSet(situationIds...).Iter() {
		if refreshErr := s.situations.RefreshSituationEpisodeAnalysis(ctx, situationId, epId); refreshErr != nil {
			return fmt.Errorf("refresh episode analysis: %w", refreshErr)
		}
	}
	return nil
}

type createAlertEpisodeParams struct {
	AlertDef   *ent.AlertDefinition
	OccurredAt time.Time
}

func (s *AlertService) createAlertEpisode(ctx context.Context, params createAlertEpisodeParams) (*ent.AlertEpisode, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, client *ent.Client) (*ent.AlertEpisode, error) {
		episodeId := uuid.New()

		knowlEntId, knowlErr := s.resolveAlertEpisodeKnowledgeEntityId(ctx, episodeId)
		if knowlErr != nil {
			return nil, fmt.Errorf("resolve alert episode knowledge entity id: %w", knowlErr)
		}

		create := client.AlertEpisode.Create().
			SetID(episodeId).
			SetAlertDefinitionID(params.AlertDef.ID).
			SetKnowledgeEntityID(knowlEntId).
			SetStartedAt(params.OccurredAt).
			SetLastObservedAt(params.OccurredAt)
		createdEpisode, createEpisodeErr := create.Save(ctx)
		if createEpisodeErr != nil {
			return nil, fmt.Errorf("create episode: %w", createEpisodeErr)
		}
		created := createdEpisode

		sitParams := rez.CreateSituationParams{
			Title:    params.AlertDef.Title,
			OpenedAt: params.OccurredAt,
			ObservationGroups: []rez.SituationObservationGroupParams{{
				Title:           params.AlertDef.Title,
				AlertEpisodeIDs: []uuid.UUID{episodeId},
			}},
			StartInvestigation: time.Since(params.OccurredAt) < time.Hour,
		}
		_, situationErr := s.situations.CreateSituation(ctx, sitParams)
		if situationErr != nil {
			return nil, fmt.Errorf("create situation: %w", situationErr)
		}

		return created, nil
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

func NewCloseInactiveAlertEpisodesWorker(alerts rez.AlertService) (*CloseInactiveAlertEpisodesWorker, error) {
	return &CloseInactiveAlertEpisodesWorker{alerts: alerts}, nil
}

type CloseInactiveAlertEpisodesWorker struct {
	jobs.WorkerDefaults[jobs.CloseInactiveAlertEpisodes]
	alerts rez.AlertService
}

func (w *CloseInactiveAlertEpisodesWorker) Work(ctx context.Context, job *jobs.Job[jobs.CloseInactiveAlertEpisodes]) error {
	systemCtx := execution.NewSystemContext(ctx)
	return w.alerts.CloseInactiveAlertEpisodes(systemCtx)
}
