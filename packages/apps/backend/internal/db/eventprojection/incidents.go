package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttype"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionIncidentObserved = "incident_observed"
)

func (s *ProjectionService) handleIncidentEvent(ctx context.Context, event *projections.IncidentEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	openedAt := attributes.OpenedAt
	if openedAt.IsZero() {
		openedAt = event.Event.OccurredAt
	}

	incidentResourceRef := rez.ProviderResourceRef{
		Provider:          event.Event.Provider,
		ProviderNamespace: event.Event.ProviderNamespace,
		ResourceRef:       event.Event.ProviderResourceRef,
	}
	incidentEntityRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryEvent,
		Kind:                knowledgeEntityKindIncident,
		ProviderResourceRef: incidentResourceRef,
	}
	incidentObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionIncidentObserved,
		EffectiveAt: openedAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Title,
			Description: attributes.Summary,
		},
		SubjectEntity: &incidentEntityRef,
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event.Event, incidentObservedEvidence)
		if ingestErr != nil {
			return fmt.Errorf("incident knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}
		knowledgeEntityId := *subj.EntityID

		queryExisting := tx.Incident.Query().
			Where(incident.KnowledgeEntityID(knowledgeEntityId))
		existing, existingErr := queryExisting.Only(ctx)
		if existingErr != nil && !ent.IsNotFound(existingErr) {
			return fmt.Errorf("query existing incident: %w", existingErr)
		}

		severityID, severityErr := s.saveProjectedIncidentSeverity(ctx, attributes)
		if severityErr != nil {
			return fmt.Errorf("upsert incident severity: %w", severityErr)
		}

		typeID, typeErr := s.saveProjectedIncidentType(ctx, attributes)
		if typeErr != nil {
			return fmt.Errorf("upsert incident type: %w", typeErr)
		}

		id := uuid.Nil
		if existing != nil {
			if existing.Title == attributes.Title &&
				existing.Summary == attributes.Summary &&
				existing.SeverityID == severityID &&
				existing.TypeID == typeID {
				return nil
			}
			id = existing.ID
		}

		setFn := func(m *ent.IncidentMutation) {
			m.SetKnowledgeEntityID(knowledgeEntityId)
			m.SetTitle(attributes.Title)
			m.SetSummary(attributes.Summary)
			m.SetSeverityID(severityID)
			m.SetTypeID(typeID)
			if !openedAt.IsZero() {
				m.SetOpenedAt(openedAt)
			}
		}
		inc, setErr := s.incidents.Set(ctx, id, setFn)
		if setErr != nil {
			return fmt.Errorf("set incident: %w", setErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindIncident,
			Id:   inc.ID,
		})

		return nil
	})
}

func (s *ProjectionService) saveProjectedIncidentSeverity(ctx context.Context, attributes projections.IncidentEventAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name(attributes.SeverityRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident severity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := s.db.Client(ctx).IncidentSeverity.Create().
		SetName(attributes.SeverityRef).
		SetRank(0).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident severity: %w", createErr)
	}
	return created.ID, nil
}

func (s *ProjectionService) saveProjectedIncidentType(ctx context.Context, attributes projections.IncidentEventAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentType.Query().
		Where(incidenttype.Name(attributes.TypeRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident type: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}
	created, createErr := s.db.Client(ctx).IncidentType.Create().
		SetName(attributes.TypeRef).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident type: %w", createErr)
	}
	return created.ID, nil
}
