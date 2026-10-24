package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	nepe "github.com/rezible/rezible/ent/normalizedeventprojectionentity"
	"github.com/rezible/rezible/ent/schema"
)

type ProviderEventIngester struct {
	db         rez.Database
	processors map[string]rez.ProviderEventProcessor
	projection rez.EventProjectionService
}

var _ rez.ProviderEventIngester = (*ProviderEventIngester)(nil)

func NewProviderEventIngester(db rez.Database, processors map[string]rez.ProviderEventProcessor, projection rez.EventProjectionService) *ProviderEventIngester {
	return &ProviderEventIngester{
		db:         db,
		processors: processors,
		projection: projection,
	}
}

func (s *ProviderEventIngester) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) ([]uuid.UUID, error) {
	proc, ok := s.processors[prov.Provider]
	if !ok {
		return nil, fmt.Errorf("no event processor registered for provider '%s'", prov.Provider)
	}

	procEvents, procErr := proc.ProcessProviderEvent(ctx, prov)
	if procErr != nil {
		return nil, fmt.Errorf("processing event: %w", procErr)
	}

	ids, saveErr := s.saveNormalizedEvents(ctx, procEvents)
	if saveErr != nil {
		return nil, fmt.Errorf("saving normalized events: %w", saveErr)
	}
	return ids, nil
}

// saveNormalizedEvents inserts the events in one transaction and returns the IDs of those newly inserted.
func (s *ProviderEventIngester) saveNormalizedEvents(ctx context.Context, evts ent.NormalizedEvents) ([]uuid.UUID, error) {
	if len(evts) == 0 {
		return nil, nil
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) ([]uuid.UUID, error) {
		ids := make([]uuid.UUID, 0, len(evts))
		for _, ev := range evts {
			id := ev.ID
			if id == uuid.Nil {
				id = uuid.New()
			}
			create := tx.NormalizedEvent.Create().
				SetID(id).
				SetKind(ev.Kind).
				SetProvider(ev.Provider).
				SetProviderNamespace(ev.ProviderNamespace).
				SetProviderResourceRef(ev.ProviderResourceRef).
				SetProviderEventSource(ev.ProviderEventSource).
				SetProviderEventRef(ev.ProviderEventRef).
				SetOccurredAt(ev.OccurredAt).
				SetReceivedAt(ev.ReceivedAt).
				SetAttributes(ev.Attributes)
			insert := create.OnConflict(sql.ConflictColumns(
				ne.FieldTenantID,
				ne.FieldProvider,
				ne.FieldProviderNamespace,
				ne.FieldProviderEventSource,
				ne.FieldProviderEventRef,
				ne.FieldProviderResourceRef,
			)).DoNothing()
			_, insertErr := insert.ID(ctx)
			if insertErr != nil {
				if errors.Is(insertErr, stdsql.ErrNoRows) {
					continue
				}
				return nil, fmt.Errorf("insert normalized event: %w", insertErr)
			}
			ids = append(ids, id)
		}
		return ids, nil
	})
}

func (s *ProviderEventIngester) ProjectNormalizedEvent(ctx context.Context, eventID uuid.UUID) error {
	ev, getEvErr := s.db.Client(ctx).NormalizedEvent.Get(ctx, eventID)
	if getEvErr != nil {
		return fmt.Errorf("query event: %w", getEvErr)
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
		// A merge can delete a knowledge subject the projector resolved; its foreign key violation makes the
		// job retry.
		projectedEntities, err = projectorFunc(schema.ReportMissingReferences(ctx), ev)
		return projectedEntities, err
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "normalized_event_projection", ev.ID.String()); lockErr != nil {
			return fmt.Errorf("lock normalized event projection: %w", lockErr)
		}

		queryProjection := tx.NormalizedEventProjection.Query().
			Where(nep.EventID(ev.ID))
		projectedExists, queryProjectionErr := queryProjection.Exist(ctx)
		if queryProjectionErr != nil {
			return fmt.Errorf("query projection receipt: %w", queryProjectionErr)
		} else if projectedExists {
			return nil
		}

		projectedEntities, projectorErr := projectEventFn(ctx, ev)
		if projectorErr != nil {
			return projectorErr
		}

		createProjection := tx.NormalizedEventProjection.Create().
			SetEventID(ev.ID)
		proj, saveProjErr := createProjection.Save(ctx)
		if saveProjErr != nil {
			return fmt.Errorf("create projection receipt: %w", saveProjErr)
		}

		if len(projectedEntities) == 0 {
			return nil
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
			return fmt.Errorf("create projection entities: %w", refsErr)
		}
		return nil
	})
}
