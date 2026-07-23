package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentimpact"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttype"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindIncident = "incident"

	knowledgeRelationshipKindIncidentImpacted = "incident_impacted"

	knowledgeAssertionIncidentObserved             = "incident_observed"
	knowledgeAssertionIncidentImpacted             = "incident_impacted_entity"
	knowledgeAssertionIncidentImpactTargetObserved = "incident_impact_target_observed"
)

func (s *ProjectionService) handleIncidentEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeIncidentEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid incident event: %w", decodeErr)
	}
	attributes := event.Attributes

	openedAt := attributes.OpenedAt
	if openedAt.IsZero() {
		openedAt = event.Event.OccurredAt
	}

	incidentSubject := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "Incident")
	incidentSubject.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindIncident,
		Reference:   attributes.ExternalRef,
		DisplayName: attributes.Title,
		Description: attributes.Summary,
	}
	incidentObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionIncidentObserved,
		EffectiveAt:     openedAt,
		SubjectAliasRef: incidentSubject,
	}

	var projected []rez.ProjectedEntityRef
	projectTxFn := func(txCtx context.Context, _ *ent.Client) error {
		knowledgeEntityID, evidenceErr := s.ingestDomainEntityEvidence(txCtx, event.Event, incidentObservedEvidence)
		if evidenceErr != nil {
			return fmt.Errorf("resolve incident knowledge entity: %w", evidenceErr)
		}

		severityID, severityErr := s.saveProjectedIncidentSeverity(txCtx, attributes)
		if severityErr != nil {
			return fmt.Errorf("upsert incident severity: %w", severityErr)
		}

		typeID, typeErr := s.saveProjectedIncidentType(txCtx, attributes)
		if typeErr != nil {
			return fmt.Errorf("upsert incident type: %w", typeErr)
		}

		queryExisting := s.db.Client(txCtx).Incident.Query().
			Where(incident.KnowledgeEntityID(knowledgeEntityID))
		existing, existingErr := queryExisting.Only(txCtx)
		if existingErr != nil && !ent.IsNotFound(existingErr) {
			return fmt.Errorf("query existing incident: %w", existingErr)
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
			m.SetKnowledgeEntityID(knowledgeEntityID)
			m.SetTitle(attributes.Title)
			m.SetSummary(attributes.Summary)
			m.SetSeverityID(severityID)
			m.SetTypeID(typeID)
			if !openedAt.IsZero() {
				m.SetOpenedAt(openedAt)
			}
		}
		inc, setErr := s.incidents.Set(txCtx, id, setFn)
		if setErr != nil {
			return fmt.Errorf("set incident: %w", setErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindIncident,
			Id:   inc.ID,
		})

		return nil
	}
	return projected, s.db.WithTx(ctx, projectTxFn)
}

func (s *ProjectionService) saveProjectedIncidentSeverity(ctx context.Context, attributes projections.IncidentSubjectAttributes) (uuid.UUID, error) {
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

func (s *ProjectionService) saveProjectedIncidentType(ctx context.Context, attributes projections.IncidentSubjectAttributes) (uuid.UUID, error) {
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

func (s *ProjectionService) handleIncidentImpactEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeIncidentImpactEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid incident impact event: %w", decodeErr)
	}
	attributes := event.Attributes
	var projected []rez.ProjectedEntityRef
	projectTxFn := func(txCtx context.Context, tx *ent.Client) error {
		queryIncidentEntity := tx.KnowledgeEntity.Query().
			Where(kne.Kind(knowledgeEntityKindIncident), kne.Reference(attributes.IncidentExternalRef))
		incidentEntity, queryEntityErr := queryIncidentEntity.Only(txCtx)
		if queryEntityErr != nil {
			return fmt.Errorf("resolve incident entity: %w", queryEntityErr)
		}
		queryIncident := tx.Incident.Query().Where(incident.KnowledgeEntityID(incidentEntity.ID))
		inc, queryIncidentErr := queryIncident.Only(txCtx)
		if queryIncidentErr != nil {
			return fmt.Errorf("resolve incident: %w", queryIncidentErr)
		}

		target := systemComponentRef(attributes.EntityKind, attributes.EntityExternalRef, attributes.EntityDisplayName, "", nil)
		targetAlias := ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindEntity,
			Provider:           event.Event.Provider,
			ProviderSubjectRef: attributes.EntityExternalRef,
			Description:        "Incident impact target",
			SubjectEntityRef:   &target,
		}
		relationship := &ent.KnowledgeRelationshipRef{
			Kind: knowledgeRelationshipKindIncidentImpacted,
			EntityRefs: [2]ent.KnowledgeEntityRef{
				{Kind: incidentEntity.Kind, Reference: incidentEntity.Reference, DisplayName: incidentEntity.DisplayName, Description: incidentEntity.Description, Properties: incidentEntity.LiveProperties},
				target,
			},
		}
		targetEvidence := ent.KnowledgeEvidenceRef{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionIncidentImpactTargetObserved,
			EffectiveAt:     event.Event.OccurredAt,
			SubjectAliasRef: targetAlias,
		}
		relationshipEvidence := ent.KnowledgeEvidenceRef{
			Kind:        projectionEvidenceKind(event.Event),
			Assertion:   knowledgeAssertionIncidentImpacted,
			EffectiveAt: event.Event.OccurredAt,
			Properties:  map[string]any{"source": attributes.Source, "note": attributes.Note},
			SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
				Kind:                   ksa.SubjectKindRelationship,
				Provider:               event.Event.Provider,
				ProviderSubjectRef:     event.Event.ProviderSubjectRef,
				SubjectRelationshipRef: relationship,
			},
		}
		if ingestErr := s.ingestProjectedEvidence(txCtx, event.Event, targetEvidence, relationshipEvidence); ingestErr != nil {
			return fmt.Errorf("ingest incident impact evidence: %w", ingestErr)
		}
		targetID, targetErr := s.setEntityFromRef(txCtx, &target, false)
		if targetErr != nil {
			return fmt.Errorf("resolve impacted entity: %w", targetErr)
		}
		upsertImpact := tx.IncidentImpact.Create().
			SetIncidentID(inc.ID).
			SetKnowledgeEntityID(targetID).
			SetSource(attributes.Source).
			SetNote(attributes.Note).
			OnConflictColumns(incidentimpact.FieldTenantID, incidentimpact.FieldIncidentID, incidentimpact.FieldKnowledgeEntityID).
			UpdateNewValues()
		impactID, impactErr := upsertImpact.ID(txCtx)
		if impactErr != nil {
			return fmt.Errorf("upsert incident impact: %w", impactErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{Kind: "incident_impact", Id: impactID})
		return nil
	}
	return projected, s.db.WithTx(ctx, projectTxFn)
}
