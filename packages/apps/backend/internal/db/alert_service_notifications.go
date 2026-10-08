package db

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rezible/rezible/ent/schema/schematypes"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ald "github.com/rezible/rezible/ent/alertdefinition"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	aie "github.com/rezible/rezible/ent/alertinstanceevent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	nev "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
)

// alertNotification is a validated alert notification. Its window is identified by the definition, the
// instance key and the provider's window start.
type alertNotification struct {
	event       *ent.NormalizedEvent
	occurredAt  time.Time
	definition  rez.AlertDefinitionValues
	instance    rez.AlertInstanceValues
	instanceKey string
	// start is the provider's window start at database precision.
	start time.Time
	// end is a resolve's end: the provider's end when supplied, otherwise its occurrence time, never
	// before the start.
	end time.Time
}

func newAlertNotification(params rez.RecordAlertInstanceParams) (*alertNotification, error) {
	event := params.Event
	if event == nil || event.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: normalized event is required", errs.ErrInvalidInput)
	}
	if params.Definition.KnowledgeEntityID == uuid.Nil {
		return nil, fmt.Errorf("%w: alert definition knowledge entity is required", errs.ErrInvalidInput)
	}
	if strings.TrimSpace(params.Definition.Title) == "" {
		return nil, fmt.Errorf("%w: alert definition title is required", errs.ErrInvalidInput)
	}
	if timeout := params.Definition.ResolutionTimeoutSeconds; timeout != nil && *timeout < 0 {
		return nil, fmt.Errorf("%w: alert resolution timeout must not be negative", errs.ErrInvalidInput)
	}
	if params.Instance.StartedAt.IsZero() {
		return nil, fmt.Errorf("%w: alert window start is required", errs.ErrInvalidInput)
	}
	instance := params.Instance
	if instance.Severity == "" {
		instance.Severity = schematypes.SignalSeverityUnknown
	}
	if severityErr := ali.SeverityValidator(instance.Severity); severityErr != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidInput, severityErr)
	}

	occurredAt := event.OccurredAt.UTC().Truncate(time.Microsecond)
	start := instance.StartedAt.UTC().Truncate(time.Microsecond)
	end := occurredAt
	if !instance.Firing && instance.EndedAt != nil {
		end = instance.EndedAt.UTC().Truncate(time.Microsecond)
		if end.Before(start) {
			return nil, fmt.Errorf("%w: alert window ended before it started", errs.ErrInvalidInput)
		}
	}
	if end.Before(start) {
		end = start
	}
	return &alertNotification{
		event:       event,
		occurredAt:  occurredAt,
		definition:  params.Definition,
		instance:    instance,
		instanceKey: projections.AlertInstanceKey(instance.InstanceID, instance.Labels),
		start:       start,
		end:         end,
	}, nil
}

// isAtLeast reports whether the notification is at least as recent as the one at (occurredAt, eventRef);
// the provider event ref breaks ties so the result does not depend on arrival order.
func (n *alertNotification) isAtLeast(occurredAt time.Time, eventRef string) bool {
	if !n.occurredAt.Equal(occurredAt) {
		return n.occurredAt.After(occurredAt)
	}
	return n.event.ProviderEventRef >= eventRef
}

func (s *AlertService) RecordAlertInstance(ctx context.Context, params rez.RecordAlertInstanceParams) (*ent.AlertDefinition, error) {
	n, validateErr := newAlertNotification(params)
	if validateErr != nil {
		return nil, validateErr
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, _ *ent.Client) (*ent.AlertDefinition, error) {
		definition, lockErr := s.upsertAndLockAlertDefinition(ctx, n)
		if lockErr != nil {
			return nil, fmt.Errorf("lock alert definition: %w", lockErr)
		}
		if n.isAtLeast(definition.MetadataObservedAt, definition.MetadataEventRef) {
			updated, metadataErr := s.applyAlertDefinitionMetadata(ctx, definition, n)
			if metadataErr != nil {
				return nil, fmt.Errorf("apply alert definition metadata: %w", metadataErr)
			}
			definition = updated
		}
		window, windowErr := s.recordAlertWindow(ctx, definition, n)
		if windowErr != nil {
			return nil, windowErr
		}
		if linkErr := s.linkAlertWindowEvent(ctx, window.ID, n.event.ID); linkErr != nil {
			return nil, fmt.Errorf("link alert window event: %w", linkErr)
		}
		if settleErr := s.storeAlertEpisodeState(ctx, definition, window.AlertEpisodeID, true); settleErr != nil {
			return nil, settleErr
		}
		return definition, nil
	})
}

// upsertAndLockAlertDefinition creates the definition from the notification if it does not exist, then locks it.
func (s *AlertService) upsertAndLockAlertDefinition(ctx context.Context, n *alertNotification) (*ent.AlertDefinition, error) {
	timeoutSeconds := int(s.cfg.DefaultResolutionTimeout / time.Second)
	if n.definition.ResolutionTimeoutSeconds != nil {
		timeoutSeconds = *n.definition.ResolutionTimeoutSeconds
	}
	createDefinition := s.db.Client(ctx).AlertDefinition.Create().
		SetKnowledgeEntityID(n.definition.KnowledgeEntityID).
		SetTitle(n.definition.Title).
		SetDescription(n.definition.Description).
		SetDefinition(n.definition.Definition).
		SetResolutionTimeoutSeconds(timeoutSeconds).
		SetMetadataObservedAt(n.occurredAt).
		SetMetadataEventRef(n.event.ProviderEventRef)
	if len(n.definition.IdentityGroupLabels) > 0 {
		createDefinition.SetIdentityGroupLabels(n.definition.IdentityGroupLabels)
	}
	upsertDefinition := createDefinition.
		OnConflictColumns(ald.FieldTenantID, ald.FieldKnowledgeEntityID).
		Ignore()
	definitionID, upsertErr := upsertDefinition.ID(ctx)
	if upsertErr != nil {
		return nil, fmt.Errorf("upsert alert definition: %w", upsertErr)
	}
	return s.lockAlertDefinition(ctx, ald.ID(definitionID))
}

// applyAlertDefinitionMetadata stores the notification's definition metadata; the caller checks it is the newest.
func (s *AlertService) applyAlertDefinitionMetadata(ctx context.Context, definition *ent.AlertDefinition, n *alertNotification) (*ent.AlertDefinition, error) {
	update := s.db.Client(ctx).AlertDefinition.UpdateOne(definition).
		SetTitle(n.definition.Title).
		SetDescription(n.definition.Description).
		SetDefinition(n.definition.Definition).
		SetMetadataObservedAt(n.occurredAt).
		SetMetadataEventRef(n.event.ProviderEventRef)
	if timeout := n.definition.ResolutionTimeoutSeconds; timeout != nil {
		update.SetResolutionTimeoutSeconds(*timeout)
	}
	if len(definition.IdentityGroupLabels) == 0 && len(n.definition.IdentityGroupLabels) > 0 {
		update.SetIdentityGroupLabels(n.definition.IdentityGroupLabels)
	}
	return update.Save(ctx)
}

// recordAlertWindow applies the notification to its window, creating the window on its first
// notification. Each stored value is an aggregate, so the result does not depend on arrival order.
func (s *AlertService) recordAlertWindow(ctx context.Context, definition *ent.AlertDefinition, n *alertNotification) (*ent.AlertInstance, error) {
	client := s.db.Client(ctx)
	queryWindow := client.AlertInstance.Query().
		Where(
			ali.InstanceKey(n.instanceKey),
			ali.FiredAt(n.start),
			ali.HasEpisodeWith(ale.AlertDefinitionID(definition.ID)),
		)
	window, queryErr := queryWindow.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return s.createAlertWindow(ctx, definition, n)
		}
		return nil, fmt.Errorf("query alert window: %w", queryErr)
	}

	peak := window.Severity
	if n.instance.Severity.Compare(peak) > 0 {
		peak = n.instance.Severity
	}
	update := client.AlertInstance.UpdateOne(window).
		SetSeverity(peak)
	if n.instance.Firing && n.occurredAt.After(window.LastObservedAt) {
		update.SetLastObservedAt(n.occurredAt)
	}
	if !n.instance.Firing && (window.ResolvedAt == nil || n.end.After(*window.ResolvedAt)) {
		update.SetResolvedAt(n.end)
	}
	queryNewest := client.AlertInstance.QueryEvents(window).
		QueryEvent().
		Order(nev.ByOccurredAt(sql.OrderDesc()), nev.ByProviderEventRef(sql.OrderDesc()))
	newest, newestErr := queryNewest.First(ctx)
	if newestErr != nil && !ent.IsNotFound(newestErr) {
		return nil, fmt.Errorf("query newest alert window notification: %w", newestErr)
	}
	if newest == nil || n.isAtLeast(newest.OccurredAt, newest.ProviderEventRef) {
		update.SetLabels(n.instance.Labels).SetSummary(n.instance.Summary)
	}
	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("update alert window: %w", updateErr)
	}
	return updated, nil
}

// createAlertWindow creates the notification's window in the earliest of the definition's episodes it
// overlaps, or in a new episode. Episodes never merge and windows never move.
func (s *AlertService) createAlertWindow(ctx context.Context, definition *ent.AlertDefinition, n *alertNotification) (*ent.AlertInstance, error) {
	timeout := time.Duration(definition.ResolutionTimeoutSeconds) * time.Second
	lastObservedAt := n.start
	var resolvedAt, end *time.Time
	if !n.instance.Firing {
		resolvedAt = &n.end
		end = resolvedAt
	} else {
		if n.occurredAt.After(n.start) {
			lastObservedAt = n.occurredAt
		}
		if timeout > 0 {
			end = new(lastObservedAt.Add(timeout))
		}
	}

	episode, episodeErr := s.overlappingAlertEpisode(ctx, definition, n.start, end)
	if episodeErr != nil {
		return nil, episodeErr
	}
	if episode == nil {
		created, createErr := s.createAlertEpisode(ctx, definition, n.start)
		if createErr != nil {
			return nil, createErr
		}
		episode = created
	} else if n.start.Before(episode.StartedAt) {
		moveStart := s.db.Client(ctx).AlertEpisode.UpdateOne(episode).
			SetStartedAt(n.start)
		if updateErr := moveStart.Exec(ctx); updateErr != nil {
			return nil, fmt.Errorf("update alert episode start: %w", updateErr)
		}
	}

	createWindow := s.db.Client(ctx).AlertInstance.Create().
		SetAlertEpisodeID(episode.ID).
		SetInstanceKey(n.instanceKey).
		SetGroupingKey(projections.AlertGroupingKey(episode.IdentityGroupLabels, n.instance.Labels, n.instanceKey)).
		SetLabels(n.instance.Labels).
		SetSummary(n.instance.Summary).
		SetSeverity(n.instance.Severity).
		SetFiredAt(n.start).
		SetLastObservedAt(lastObservedAt).
		SetNillableResolvedAt(resolvedAt)
	created, createErr := createWindow.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create alert window: %w", createErr)
	}
	return created, nil
}

// overlappingAlertEpisode is the earliest of the definition's episodes holding a window that overlaps
// [start, end + grace], or nil. A window ends at its effective end, otherwise at its timeout when the
// definition has one, otherwise never; a nil end is open-ended.
func (s *AlertService) overlappingAlertEpisode(ctx context.Context, definition *ent.AlertDefinition, start time.Time, end *time.Time) (*ent.AlertEpisode, error) {
	grace := s.cfg.FlapGrace
	timeout := time.Duration(definition.ResolutionTimeoutSeconds) * time.Second
	queryCandidates := s.db.Client(ctx).AlertInstance.Query().
		Where(
			ali.HasEpisodeWith(ale.AlertDefinitionID(definition.ID)),
			ali.Or(ali.EndedAtIsNil(), ali.EndedAtGTE(start.Add(-grace))),
		).
		WithEpisode()
	if end != nil {
		queryCandidates.Where(ali.FiredAtLTE(end.Add(grace)))
	}
	candidates, queryErr := queryCandidates.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query overlapping alert windows: %w", queryErr)
	}
	var overlapping []*ent.AlertEpisode
	for _, window := range candidates {
		windowEnd := window.EndedAt
		if windowEnd == nil && timeout > 0 {
			windowEnd = new(window.LastObservedAt.Add(timeout))
		}
		if windowEnd == nil || !windowEnd.Add(grace).Before(start) {
			overlapping = append(overlapping, window.Edges.Episode)
		}
	}
	if len(overlapping) == 0 {
		return nil, nil
	}
	return slices.MinFunc(overlapping, func(a, b *ent.AlertEpisode) int {
		if !a.StartedAt.Equal(b.StartedAt) {
			return a.StartedAt.Compare(b.StartedAt)
		}
		return bytes.Compare(a.ID[:], b.ID[:])
	}), nil
}

func (s *AlertService) createAlertEpisode(ctx context.Context, definition *ent.AlertDefinition, startedAt time.Time) (*ent.AlertEpisode, error) {
	episodeID := uuid.New()
	episodeSubjectRef := rez.KnowledgeSubjectRef{
		Entity: &rez.KnowledgeEntityRef{
			Category:            kne.CategoryEvent,
			Kind:                "alert_episode",
			ProviderResourceRef: projections.InternalEntityResourceRef(episodeID),
		},
	}
	alias, aliasErr := s.knowledge.ResolveInternalSubject(ctx, episodeSubjectRef)
	if aliasErr != nil {
		return nil, fmt.Errorf("create alert episode knowledge entity: %w", aliasErr)
	}
	if alias.EntityID == nil {
		return nil, fmt.Errorf("alert episode knowledge alias %s has no entity", alias.ID)
	}
	createEpisode := s.db.Client(ctx).AlertEpisode.Create().
		SetID(episodeID).
		SetAlertDefinitionID(definition.ID).
		SetKnowledgeEntityID(*alias.EntityID).
		SetStartedAt(startedAt)
	if len(definition.IdentityGroupLabels) > 0 {
		createEpisode.SetIdentityGroupLabels(definition.IdentityGroupLabels)
	}
	created, createErr := createEpisode.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create alert episode: %w", createErr)
	}
	return created, nil
}

// linkAlertWindowEvent links a notification's event to its window once; recording the same
// notification again changes nothing.
func (s *AlertService) linkAlertWindowEvent(ctx context.Context, windowID, eventID uuid.UUID) error {
	client := s.db.Client(ctx)
	queryLinked := client.AlertInstanceEvent.Query().
		Where(aie.EventID(eventID))
	linked, queryErr := queryLinked.Exist(ctx)
	if queryErr != nil || linked {
		return queryErr
	}
	createLink := client.AlertInstanceEvent.Create().
		SetAlertInstanceID(windowID).
		SetEventID(eventID)
	return createLink.Exec(ctx)
}

// storeAlertEpisodeState recomputes the episode's state as of processing time, including its highest
// severity, stores any change and schedules settlement at its next deadline. It tells situations the
// episode changed when its stored state changed, or always for newEvidence: a recorded notification is
// evidence even when it changes no state. The definition lock must be held.
func (s *AlertService) storeAlertEpisodeState(ctx context.Context, definition *ent.AlertDefinition, episodeID uuid.UUID, newEvidence bool) error {
	client := s.db.Client(ctx)
	queryEpisode := client.AlertEpisode.Query().
		Where(ale.ID(episodeID)).
		WithInstances()
	episode, queryErr := queryEpisode.Only(ctx)
	if queryErr != nil {
		return fmt.Errorf("query alert episode: %w", queryErr)
	}
	windows := newAlertEpisodeState(episode.Edges.Instances)
	timeout := time.Duration(definition.ResolutionTimeoutSeconds) * time.Second
	settlement := windows.settle(timeout, s.cfg.FlapGrace, s.clock.Now())

	changed := newEvidence
	for i, window := range windows {
		end := settlement.ends[i]
		if end.isStoredOn(window) {
			continue
		}
		changed = true
		updateWindow := client.AlertInstance.UpdateOne(window)
		if end == nil {
			updateWindow.ClearEndedAt().ClearEndReason()
		} else {
			updateWindow.SetEndedAt(end.at).SetEndReason(end.reason)
		}
		if updateErr := updateWindow.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update alert window end: %w", updateErr)
		}
	}

	highestSeverity := windows.highestSeverity()
	closedAt := settlement.closedAt
	closureChanged := (closedAt == nil) != (episode.ClosedAt == nil) || (closedAt != nil && !closedAt.Equal(*episode.ClosedAt))
	if closureChanged || highestSeverity != episode.HighestSeverity {
		changed = true
		updateEpisode := client.AlertEpisode.UpdateOne(episode).
			SetHighestSeverity(highestSeverity)
		if closedAt == nil {
			updateEpisode.ClearClosedAt()
		} else {
			updateEpisode.SetClosedAt(*closedAt)
		}
		if updateErr := updateEpisode.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update alert episode state: %w", updateErr)
		}
	}
	if changed {
		if notifyErr := s.signals.NotifySignalChanged(ctx, *episode.KnowledgeEntityID); notifyErr != nil {
			return fmt.Errorf("notify alert episode changed: %w", notifyErr)
		}
	}

	if settlement.nextDeadline != nil {
		settleArgs := jobs.SettleAlertEpisode{
			EpisodeID: episode.ID,
			DueAt:     *settlement.nextDeadline,
		}
		settleOpts := &river.InsertOpts{ScheduledAt: *settlement.nextDeadline}
		if _, insertErr := s.jobs.Insert(ctx, settleArgs, settleOpts); insertErr != nil {
			return fmt.Errorf("schedule alert episode settlement: %w", insertErr)
		}
	}

	return nil
}

func NewSettleAlertEpisodeWorker(service *AlertService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.settleAlertEpisode)
}

// settleAlertEpisode stores an episode's state at a deadline and schedules the next one.
func (s *AlertService) settleAlertEpisode(ctx context.Context, args jobs.SettleAlertEpisode) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		definition, lockErr := s.lockAlertDefinition(ctx, ald.HasEpisodesWith(ale.ID(args.EpisodeID)))
		if lockErr != nil {
			if ent.IsNotFound(lockErr) {
				return nil
			}
			return fmt.Errorf("lock alert definition: %w", lockErr)
		}
		return s.storeAlertEpisodeState(ctx, definition, args.EpisodeID, false)
	})
}
