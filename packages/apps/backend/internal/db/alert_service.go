package db

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ald "github.com/rezible/rezible/ent/alertdefinition"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	nev "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/situations"
)

type AlertService struct {
	cfg       rez.AlertsConfig
	db        rez.Database
	knowledge rez.KnowledgeGraphIngestionService
	jobs      rez.JobService
	signals   rez.SituationSignalService
	clock     rez.Clock
}

var _ situations.SituationSignalSource = (*AlertService)(nil)

func NewAlertService(cfg rez.AlertsConfig, clock rez.Clock, db rez.Database, js rez.JobService, knowledge rez.KnowledgeGraphIngestionService, signals rez.SituationSignalService) (*AlertService, error) {
	s := &AlertService{
		cfg:       cfg,
		clock:     clock,
		db:        db,
		jobs:      js,
		signals:   signals,
		knowledge: knowledge,
	}

	return s, nil
}

func (s *AlertService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) (*ent.ListResult[ent.AlertDefinition], error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		WithSituationSignalAttention().
		Order(ald.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(ald.TitleContainsFold(search))
	}
	return ent.DoListQuery[ent.AlertDefinition, *ent.AlertDefinitionQuery](ctx, query, params.ListParams)
}

func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*ent.AlertDefinition, error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		Where(ald.ID(id)).
		WithSituationSignalAttention()
	return query.Only(ctx)
}

// ListAlertEpisodes lists episodes by start. With a situation, it lists the episodes that are the
// situation's signals, read from its membership rows.
func (s *AlertService) ListAlertEpisodes(ctx context.Context, params rez.ListAlertEpisodesParams) (*ent.ListResult[ent.AlertEpisode], error) {
	client := s.db.Client(ctx)
	query := client.AlertEpisode.Query().
		WithAlertDefinition(func(query *ent.AlertDefinitionQuery) {
			query.WithSituationSignalAttention()
		}).
		WithInstances(func(query *ent.AlertInstanceQuery) {
			query.Order(ali.ByFiredAt(), ali.ByID())
		}).
		Order(ale.ByStartedAt(params.GetOrder()), ale.ByID(params.GetOrder()))
	if params.SituationID != uuid.Nil {
		queryMembers := client.SituationSignal.Query().
			Where(sitsig.SituationID(params.SituationID)).
			Select(sitsig.FieldKnowledgeEntityID)
		members, membersErr := queryMembers.All(ctx)
		if membersErr != nil {
			return nil, fmt.Errorf("query situation signals: %w", membersErr)
		}
		entityIDs := make([]uuid.UUID, len(members))
		for i, member := range members {
			entityIDs[i] = member.KnowledgeEntityID
		}
		query.Where(ale.KnowledgeEntityIDIn(entityIDs...))
	}
	return ent.DoListQuery[ent.AlertEpisode, *ent.AlertEpisodeQuery](ctx, query, params.ListParams)
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

// lockAlertDefinition loads a definition with a row lock. Every write to a definition's windows, episodes and settings
// happens under it, so a definition changes one notification or setting at a time. It is the first lock
// taken in its transaction.
func (s *AlertService) lockAlertDefinition(ctx context.Context, where predicate.AlertDefinition) (*ent.AlertDefinition, error) {
	lockDefinition := s.db.Client(ctx).AlertDefinition.Query().
		Where(where).
		ForUpdate()
	return lockDefinition.Only(ctx)
}

func (s *AlertService) SetAlertIdentityGroupLabels(ctx context.Context, id uuid.UUID, labels []string) (*ent.AlertDefinition, error) {
	normalized := make([]string, 0, len(labels))
	seen := mapset.NewSet[string]()
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label != "" && seen.Add(label) {
			normalized = append(normalized, label)
		}
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.AlertDefinition, error) {
		definition, lockErr := s.lockAlertDefinition(ctx, ald.ID(id))
		if lockErr != nil {
			return nil, fmt.Errorf("lock alert definition: %w", lockErr)
		}
		update := tx.AlertDefinition.UpdateOne(definition)
		if len(normalized) == 0 {
			update.ClearIdentityGroupLabels()
		} else {
			update.SetIdentityGroupLabels(normalized)
		}
		if updateErr := update.Exec(ctx); updateErr != nil {
			return nil, fmt.Errorf("update alert identity group labels: %w", updateErr)
		}
		return s.GetAlert(ctx, id)
	})
}

func (s *AlertService) SetAlertSituationSignalAttention(ctx context.Context, id uuid.UUID, level ssa.Level) (*ent.AlertDefinition, error) {
	if levelErr := ssa.LevelValidator(level); levelErr != nil {
		return nil, fmt.Errorf("%w: %w", rez.ErrInvalidInput, levelErr)
	}
	userID, hasUser := execution.GetContext(ctx).UserID()
	if !hasUser {
		return nil, fmt.Errorf("%w: setting situation signal attention requires a user", rez.ErrForbidden)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.AlertDefinition, error) {
		definition, lockErr := s.lockAlertDefinition(ctx, ald.ID(id))
		if lockErr != nil {
			return nil, fmt.Errorf("lock alert definition: %w", lockErr)
		}
		changed := false
		if definition.SituationSignalAttentionID != nil {
			updateAttention := tx.SituationSignalAttention.Update().
				Where(ssa.ID(*definition.SituationSignalAttentionID), ssa.LevelNEQ(level)).
				SetLevel(level).
				SetSetAt(s.clock.Now()).
				SetSetByUserID(userID)
			updated, updateErr := updateAttention.Save(ctx)
			if updateErr != nil {
				return nil, fmt.Errorf("update situation signal attention: %w", updateErr)
			}
			changed = updated > 0
		} else if level != ssa.LevelDefault {
			changed = true
			createAttention := tx.SituationSignalAttention.Create().
				SetLevel(level).
				SetSetAt(s.clock.Now()).
				SetSetByUserID(userID)
			attention, createErr := createAttention.Save(ctx)
			if createErr != nil {
				return nil, fmt.Errorf("create situation signal attention: %w", createErr)
			}
			linkAttention := tx.AlertDefinition.UpdateOne(definition).
				SetSituationSignalAttentionID(attention.ID)
			if linkErr := linkAttention.Exec(ctx); linkErr != nil {
				return nil, fmt.Errorf("link situation signal attention: %w", linkErr)
			}
		}
		if changed {
			if notifyErr := s.notifyUnclosedEpisodesChanged(ctx, definition.ID); notifyErr != nil {
				return nil, notifyErr
			}
		}
		return s.GetAlert(ctx, id)
	})
}

// notifyUnclosedEpisodesChanged tells situations that each of the definition's unclosed episodes changed.
func (s *AlertService) notifyUnclosedEpisodesChanged(ctx context.Context, definitionID uuid.UUID) error {
	queryEpisodes := s.db.Client(ctx).AlertEpisode.Query().
		Where(ale.AlertDefinitionID(definitionID), ale.ClosedAtIsNil()).
		Select(ale.FieldKnowledgeEntityID)
	episodes, queryErr := queryEpisodes.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query unclosed alert episodes: %w", queryErr)
	}
	for _, episode := range episodes {
		if notifyErr := s.signals.NotifySignalChanged(ctx, *episode.KnowledgeEntityID); notifyErr != nil {
			return fmt.Errorf("notify alert episode changed: %w", notifyErr)
		}
	}
	return nil
}

func (s *AlertService) SituationSignalKind() situations.SignalKind {
	return situations.SignalKindAlertEpisode
}

// LoadSignals describes the alert episodes with the given knowledge entities from their stored state, with
// their definitions' baselines when asked for history.
func (s *AlertService) LoadSignals(ctx context.Context, entityIDs []uuid.UUID, opts situations.LoadSignalsOptions) ([]situations.Signal, error) {
	if len(entityIDs) == 0 {
		return nil, nil
	}
	queryEpisodes := s.db.Client(ctx).AlertEpisode.Query().
		Where(ale.KnowledgeEntityIDIn(entityIDs...)).
		WithAlertDefinition(func(query *ent.AlertDefinitionQuery) {
			query.WithSituationSignalAttention()
		}).
		WithInstances(func(query *ent.AlertInstanceQuery) {
			query.WithEvents(func(query *ent.AlertInstanceEventQuery) {
				query.WithEvent(func(query *ent.NormalizedEventQuery) {
					query.Select(nev.FieldReceivedAt)
				})
			})
		}).
		Order(ale.ByStartedAt(), ale.ByID())
	episodes, queryErr := queryEpisodes.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query alert episodes: %w", queryErr)
	}

	baselines := make(map[uuid.UUID]schematypes.SituationBaselineFacts)
	if opts.History {
		loaded, baselinesErr := s.alertBaselines(ctx, episodes)
		if baselinesErr != nil {
			return nil, baselinesErr
		}
		baselines = loaded
	}

	asOf := s.clock.Now()
	signals := make([]situations.Signal, 0, len(episodes))
	for _, episode := range episodes {
		definition := episode.Edges.AlertDefinition
		windows := newAlertEpisodeState(episode.Edges.Instances)
		facts := windows.facts(asOf)
		facts.Severity = episode.HighestSeverity
		facts.Description = definition.Description
		facts.Definition = definition.Definition
		facts.IdentityGroupLabels = episode.IdentityGroupLabels
		facts.Baseline = baselines[episode.ID]

		attention := ssa.LevelDefault
		var attentionSetAt *time.Time
		if setting := definition.Edges.SituationSignalAttention; setting != nil {
			attention = setting.Level
			attentionSetAt = &setting.SetAt
		}
		var eventIDs []uuid.UUID
		var lastActivityAt time.Time
		for _, window := range windows {
			for _, link := range window.Edges.Events {
				eventIDs = append(eventIDs, link.EventID)
				if receivedAt := link.Edges.Event.ReceivedAt; receivedAt.After(lastActivityAt) {
					lastActivityAt = receivedAt
				}
			}
		}
		slices.SortFunc(eventIDs, func(a, b uuid.UUID) int {
			return bytes.Compare(a[:], b[:])
		})

		signals = append(signals, situations.Signal{
			EntityID:             *episode.KnowledgeEntityID,
			Kind:                 situations.SignalKindAlertEpisode,
			Title:                definition.Title,
			SourceEntityID:       definition.KnowledgeEntityID,
			SignalAttention:      attention,
			SignalAttentionSetAt: attentionSetAt,
			Severity:             episode.HighestSeverity,
			StartedAt:            episode.StartedAt,
			FinishedAt:           episode.ClosedAt,
			Active:               facts.Active,
			LastActivityAt:       lastActivityAt,
			EventIDs:             eventIDs,
			Revision:             len(eventIDs),
			Alert:                &facts,
		})
	}
	return signals, nil
}

// alertBaselines describes, for each episode, how often its definition fired within the novelty window before
// it started and for how long, from the definition's stored episodes.
func (s *AlertService) alertBaselines(ctx context.Context, episodes []*ent.AlertEpisode) (map[uuid.UUID]schematypes.SituationBaselineFacts, error) {
	baselines := make(map[uuid.UUID]schematypes.SituationBaselineFacts, len(episodes))
	if len(episodes) == 0 {
		return baselines, nil
	}
	client := s.db.Client(ctx)
	queryEarliest := client.AlertEpisode.Query().
		Order(ale.ByStartedAt()).
		Select(ale.FieldStartedAt)
	earliest, earliestErr := queryEarliest.First(ctx)
	if earliestErr != nil {
		return nil, fmt.Errorf("query earliest alert episode: %w", earliestErr)
	}

	windows := make([]predicate.AlertEpisode, len(episodes))
	for i, episode := range episodes {
		windows[i] = ale.And(
			ale.AlertDefinitionID(episode.AlertDefinitionID),
			ale.StartedAtGTE(episode.StartedAt.Add(-situations.NoveltyWindow)),
			ale.StartedAtLT(episode.StartedAt),
		)
	}
	queryHistory := client.AlertEpisode.Query().
		Where(ale.Or(windows...)).
		WithInstances()
	history, historyErr := queryHistory.All(ctx)
	if historyErr != nil {
		return nil, fmt.Errorf("query alert definition history: %w", historyErr)
	}

	for _, episode := range episodes {
		baseline := schematypes.SituationBaselineFacts{}
		if age := episode.StartedAt.Sub(earliest.StartedAt); age > 0 {
			baseline.ObservedHistoryAgeDays = int(age / (24 * time.Hour))
			baseline.HasSufficientHistory = age >= situations.SufficientHistoryAge
		}
		from := episode.StartedAt.Add(-situations.NoveltyWindow)
		var durations []time.Duration
		for _, earlier := range history {
			if earlier.AlertDefinitionID != episode.AlertDefinitionID || earlier.ID == episode.ID ||
				earlier.StartedAt.Before(from) || !earlier.StartedAt.Before(episode.StartedAt) {
				continue
			}
			baseline.Occurrences++
			// Only an episode whose windows were all resolved by the source has a measured duration.
			resolved := !slices.ContainsFunc(earlier.Edges.Instances, func(window *ent.AlertInstance) bool {
				return window.EndReason == nil || *window.EndReason != ali.EndReasonResolved
			})
			if earlier.ClosedAt != nil && resolved {
				durations = append(durations, newAlertEpisodeState(earlier.Edges.Instances).firingDuration(*earlier.ClosedAt))
			}
		}
		if len(durations) > 0 {
			slices.Sort(durations)
			middle := len(durations) / 2
			median := durations[middle]
			if len(durations)%2 == 0 {
				median = (durations[middle-1] + durations[middle]) / 2
			}
			baseline.MedianDurationSeconds = new(int64(median / time.Second))
		}
		baselines[episode.ID] = baseline
	}
	return baselines, nil
}
