package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	nepe "github.com/rezible/rezible/ent/normalizedeventprojectionentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/jobs"
)

type ProviderEventIngester struct {
	db              rez.Database
	processors      map[string]rez.ProviderEventProcessor
	projection      rez.EventProjectionService
	ingestion       *KnowledgeGraphIngestionService
	evidenceSkipped metric.Int64Counter
}

var _ rez.ProviderEventIngester = (*ProviderEventIngester)(nil)

func NewProviderEventIngester(db rez.Database, processors map[string]rez.ProviderEventProcessor, projection rez.EventProjectionService, ingestion *KnowledgeGraphIngestionService) *ProviderEventIngester {
	counter, counterErr := otel.Meter("github.com/rezible/rezible").Int64Counter("rezible.knowledge_graph.evidence_skipped", metric.WithDescription("Relationship evidence skipped during projection"))
	if counterErr != nil {
		panic("telemetry instruments err: " + counterErr.Error())
	}
	return &ProviderEventIngester{
		db:              db,
		processors:      processors,
		projection:      projection,
		ingestion:       ingestion,
		evidenceSkipped: counter,
	}
}

func (s *ProviderEventIngester) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (*rez.NormalizedEventIngestResult, error) {
	proc, ok := s.processors[prov.Provider]
	if !ok {
		return nil, fmt.Errorf("no event processor registered for provider '%s'", prov.Provider)
	}

	procEvents, procErr := proc.ProcessProviderEvent(ctx, prov)
	if procErr != nil {
		return nil, fmt.Errorf("processing event: %w", procErr)
	}

	result, saveErr := s.saveNormalizedEvents(ctx, procEvents)
	if saveErr != nil {
		return nil, fmt.Errorf("saving normalized events: %w", saveErr)
	}
	return result, nil
}

func (s *ProviderEventIngester) makeEventUpsert(c *ent.NormalizedEventCreate, occAt time.Time) *ent.NormalizedEventUpsertOne {
	return c.OnConflict(
		sql.ConflictColumns(
			ne.FieldTenantID,
			ne.FieldProvider,
			ne.FieldProviderNamespace,
			ne.FieldProviderEventSource,
			ne.FieldProviderEventRef,
			ne.FieldProviderResourceRef,
		),
		sql.ResolveWith(func(u *sql.UpdateSet) {
			u.SetExcluded(ne.FieldLastOccurredAt)
		}),
		sql.UpdateWhere(sql.LT(ne.FieldLastOccurredAt, occAt)),
	)
}

// saveNormalizedEvents atomically distinguishes new statements from later observations and redeliveries.
func (s *ProviderEventIngester) saveNormalizedEvents(ctx context.Context, evts ent.NormalizedEvents) (*rez.NormalizedEventIngestResult, error) {
	if len(evts) == 0 {
		return &rez.NormalizedEventIngestResult{}, nil
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.NormalizedEventIngestResult, error) {
		result := &rez.NormalizedEventIngestResult{}
		for _, ev := range evts {
			id := ne.DefaultID()
			create := tx.NormalizedEvent.Create().
				SetID(id).
				SetKind(ev.Kind).
				SetProvider(ev.Provider).
				SetProviderNamespace(ev.ProviderNamespace).
				SetProviderResourceRef(ev.ProviderResourceRef).
				SetProviderEventSource(ev.ProviderEventSource).
				SetProviderEventRef(ev.ProviderEventRef).
				SetReceivedAt(ev.ReceivedAt).
				SetAttributes(ev.Attributes).
				SetOccurredAt(ev.OccurredAt).
				SetLastOccurredAt(ev.OccurredAt)
			upsert := s.makeEventUpsert(create, ev.OccurredAt)
			savedID, insertErr := upsert.ID(ctx)
			if insertErr != nil {
				if errors.Is(insertErr, stdsql.ErrNoRows) {
					continue
				}
				return nil, fmt.Errorf("insert normalized event: %w", insertErr)
			}
			if savedID == id {
				result.NewEventIDs = append(result.NewEventIDs, id)
			} else {
				result.Reobservations = append(result.Reobservations, rez.NormalizedEventReobservation{EventID: savedID, At: ev.OccurredAt})
			}
		}
		return result, nil
	})
}

func (s *ProviderEventIngester) ProjectNormalizedEvent(ctx context.Context, eventID uuid.UUID, retryMissingEndpoints bool) error {
	ev, evErr := s.db.Client(ctx).NormalizedEvent.Get(ctx, eventID)
	if evErr != nil {
		return fmt.Errorf("query event: %w", evErr)
	}

	projectorFunc, ok := s.projection.GetEventProjectorFunc(ev)
	if !ok {
		return nil
	}

	projectEventFn := func(ctx context.Context, ev *ent.NormalizedEvent) (projectedEntities []rez.ProjectedEntityRef, err error) {
		defer func() {
			if v := recover(); v != nil {
				err = fmt.Errorf("projector panic: %v", v)
			}
		}()
		projectedEntities, err = projectorFunc(ctx, ev)
		return projectedEntities, err
	}

	skipped, projectionErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) ([]schematypes.SkippedEvidenceRef, error) {
		// Serialize receipt creation and repeats for this event before taking identity locks, so an early
		// repeat is either folded by projection or by the following re-observation job. A row lock on the event
		// would also block saveNormalizedEvents' upsert of a repeat until projection commits.
		if lockErr := s.db.AcquireTxLocks(ctx, "normalized_event_projection", ev.ID.String()); lockErr != nil {
			return nil, fmt.Errorf("lock normalized event projection: %w", lockErr)
		}

		queryProjection := tx.NormalizedEventProjection.Query().
			Where(nep.EventID(eventID))
		projectedExists, queryProjectionErr := queryProjection.Exist(ctx)
		if queryProjectionErr != nil {
			return nil, fmt.Errorf("query projection receipt: %w", queryProjectionErr)
		} else if projectedExists {
			return nil, nil
		}

		ev, reloadErr := tx.NormalizedEvent.Get(ctx, eventID)
		if reloadErr != nil {
			return nil, fmt.Errorf("reload event after projection lock: %w", reloadErr)
		}

		diagnostics := &projectionDiagnostics{
			retryMissingEndpoints: retryMissingEndpoints,
			skipped:               []schematypes.SkippedEvidenceRef{},
		}
		// Only provider projections install this collector, so internal commands still fail invalid refs.
		projectionCtx := withProjectionDiagnostics(ctx, diagnostics)
		projectedEntities, projectorErr := projectEventFn(projectionCtx, ev)
		if projectorErr != nil {
			return nil, projectorErr
		}

		createProjection := tx.NormalizedEventProjection.Create().
			SetEventID(eventID).
			SetSkippedRefs(diagnostics.skipped)
		proj, saveProjErr := createProjection.Save(ctx)
		if saveProjErr != nil {
			return nil, fmt.Errorf("create projection receipt: %w", saveProjErr)
		}

		if len(projectedEntities) == 0 {
			return diagnostics.skipped, nil
		}

		upsertRefs := tx.NormalizedEventProjectionEntity.
			MapCreateBulk(projectedEntities, func(c *ent.NormalizedEventProjectionEntityCreate, i int) {
				ref := projectedEntities[i]
				c.SetProjectionID(proj.ID)
				c.SetDomainEntityKind(ref.Kind)
				c.SetDomainEntityID(ref.Id)
			}).
			OnConflictColumns(nepe.FieldTenantID, nepe.FieldProjectionID, nepe.FieldDomainEntityID).
			DoNothing()
		if refsErr := upsertRefs.Exec(ctx); refsErr != nil {
			return nil, fmt.Errorf("create projection entities: %w", refsErr)
		}
		return diagnostics.skipped, nil
	})
	if projectionErr == nil {
		s.recordSkippedEvidence(ctx, eventID, skipped)
	}
	return projectionErr
}

type projectionDiagnosticsKey struct{}
type projectionDiagnostics struct {
	retryMissingEndpoints bool
	skipped               []schematypes.SkippedEvidenceRef
}

// withProjectionDiagnostics makes evidence ingestion under ctx record invalid relationship refs in d instead of
// failing on them.
func withProjectionDiagnostics(ctx context.Context, d *projectionDiagnostics) context.Context {
	return context.WithValue(ctx, projectionDiagnosticsKey{}, d)
}

// projectionDiagnosticsFrom returns the diagnostics installed by withProjectionDiagnostics, or nil.
func projectionDiagnosticsFrom(ctx context.Context) *projectionDiagnostics {
	diagnostics, _ := ctx.Value(projectionDiagnosticsKey{}).(*projectionDiagnostics)
	return diagnostics
}

func (d *projectionDiagnostics) record(ref rez.KnowledgeEvidenceRef, reason schematypes.SkippedEvidenceReason, skipErr error) error {
	// Let endpoint events queued alongside this event project before accepting a partial projection.
	if reason == schematypes.SkippedEvidenceReasonMissingEndpoint && d.retryMissingEndpoints {
		return jobs.MarkRetryableError(skipErr)
	}
	d.skipped = append(d.skipped, schematypes.SkippedEvidenceRef{
		Reason: reason, ResourceRef: ref.Subject.Relationship.ProviderResourceRef.ResourceRef, Error: skipErr.Error(),
	})
	return nil
}

func (s *ProviderEventIngester) recordSkippedEvidence(ctx context.Context, eventID uuid.UUID, skipped []schematypes.SkippedEvidenceRef) {
	for _, entry := range skipped {
		level := slog.LevelWarn
		if entry.Reason == schematypes.SkippedEvidenceReasonTopologyMismatch {
			level = slog.LevelError
		}
		slog.Log(ctx, level, "skipped relationship evidence", "event_id", eventID, "resource_ref", entry.ResourceRef, "reason", entry.Reason, "error", entry.Error)
		s.evidenceSkipped.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", string(entry.Reason))))
	}
}

func (s *ProviderEventIngester) ReobserveNormalizedEvent(ctx context.Context, eventID uuid.UUID, at time.Time) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// Serialize with projection for this event before taking identity locks. An unprojected repeat is
		// retained in last_occurred_at and applied by projection after it acquires the same lock. This is not a row
		// lock on the event, which would block saveNormalizedEvents' upsert of a repeat.
		if lockErr := s.db.AcquireTxLocks(ctx, "normalized_event_projection", eventID.String()); lockErr != nil {
			return fmt.Errorf("lock normalized event re-observation: %w", lockErr)
		}
		queryProjection := tx.NormalizedEventProjection.Query().Where(nep.EventID(eventID))
		projected, queryErr := queryProjection.Exist(ctx)
		if queryErr != nil {
			return fmt.Errorf("query projection receipt: %w", queryErr)
		}
		if !projected {
			return nil
		}
		return s.ingestion.reobserveEvidence(ctx, eventID, at)
	})
}
