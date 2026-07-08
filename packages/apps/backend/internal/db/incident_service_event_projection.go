package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttype"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionIncidentObserved = "incident_observed"
	knowledgeKindIncident     = "incident"
)

func (s *IncidentService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) ([]rez.ProjectedDomainEntityRef, error) {
	if projections.SubjectKindIncident.Matches(event) {
		decoded, validationErr := projections.DecodeIncidentEvent(event)
		if validationErr != nil || decoded == nil {
			return nil, fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleIncidentEventProjection(ctx, decoded)
	}
	if projections.SubjectKindIncidentImpact.Matches(event) {
		//decoded, validationErr := projections.DecodeIncidentImpactEvent(event)
		//if validationErr != nil || decoded == nil {
		//	return nil, fmt.Errorf("invalid event: %w", validationErr)
		//}
		//return s.handleIncidentImpactEventProjection(ctx, decoded)
	}
	return nil, nil
}

func (s *IncidentService) lookupIncidentKnowledgeEntityId(ctx context.Context, ref ent.KnowledgeSubjectAliasRef) (uuid.UUID, error) {
	keId, lookupErr := s.knowledge.LookupEntityIdByAliasRefs(ctx, ref)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return uuid.Nil, projections.Retryable(fmt.Errorf("incident entity not found: %s", ref.ProviderSubjectRef))
		}
		return uuid.Nil, fmt.Errorf("lookup incident entity alias: %w", lookupErr)
	}
	return keId, nil
}

func (s *IncidentService) handleIncidentEventProjection(ctx context.Context, ie *projections.IncidentEvent) ([]rez.ProjectedDomainEntityRef, error) {
	attrs := ie.Attributes
	openedAt := attrs.OpenedAt
	if openedAt.IsZero() {
		openedAt = ie.Event.DeriveObservedAt()
	}

	incidentAliasRef := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           ie.Event.Provider,
		ProviderSubjectRef: ie.Event.ProviderSubjectRef,
	}
	incidentObservedEvidence := rez.ProjectedKnowledgeEvidence{
		Kind:        ke.EvidenceKindObserved,
		Assertion:   assertionIncidentObserved,
		EffectiveAt: openedAt,
		SubjectAlias: rez.ProjectedKnowledgeEvidenceSubjectAlias{
			Description: "Incident Opened",
			AliasRef:    incidentAliasRef,
			SubjectEntityRef: &ent.KnowledgeEntityRef{
				Kind:        knowledgeKindIncident,
				Reference:   attrs.ExternalRef,
				DisplayName: attrs.Title,
				Description: attrs.Summary,
			},
		},
	}
	knowledgeEvidence := []rez.ProjectedKnowledgeEvidence{incidentObservedEvidence}

	var projEnts []rez.ProjectedDomainEntityRef
	return projEnts, s.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		evidenceErr := s.knowledge.IngestProjectedEventEvidence(ctx, ie.Event, knowledgeEvidence)
		if evidenceErr != nil {
			return fmt.Errorf("resolve incident knowledge entity: %w", evidenceErr)
		}

		kneId, idErr := s.lookupIncidentKnowledgeEntityId(ctx, incidentAliasRef)
		if idErr != nil {
			return fmt.Errorf("resolve incident knowledge entity id: %w", idErr)
		}

		sevId, severityErr := s.saveProjectedIncidentSeverity(ctx, attrs)
		if severityErr != nil {
			return fmt.Errorf("upsert incident severity: %w", severityErr)
		}

		typeId, typeErr := s.saveProjectedIncidentType(ctx, attrs)
		if typeErr != nil {
			return fmt.Errorf("upsert incident type: %w", typeErr)
		}

		queryExistingByKnowledgeEntity := s.db.Client(ctx).Incident.Query().
			Where(incident.KnowledgeEntityID(kneId))
		existing, existingErr := queryExistingByKnowledgeEntity.Only(ctx)
		if existingErr != nil && !ent.IsNotFound(existingErr) {
			return fmt.Errorf("query existing incident: %w", existingErr)
		}

		id := uuid.Nil
		if existing != nil {
			// TODO: helper method on attributes struct?
			if existing.Title == attrs.Title &&
				existing.Summary == attrs.Summary &&
				existing.SeverityID == sevId &&
				existing.TypeID == typeId {
				return nil
			}
			id = existing.ID
		}

		setFn := func(m *ent.IncidentMutation) {
			m.SetKnowledgeEntityID(kneId)
			m.SetTitle(attrs.Title)
			m.SetSummary(attrs.Summary)
			m.SetSeverityID(sevId)
			m.SetTypeID(typeId)
			if !openedAt.IsZero() {
				m.SetOpenedAt(openedAt)
			}
		}
		inc, setErr := s.Set(ctx, id, setFn)
		if setErr != nil {
			return fmt.Errorf("set incident: %w", setErr)
		}
		projEnts = append(projEnts, rez.ProjectedDomainEntityRef{
			Kind: "incident",
			Id:   inc.ID,
		})

		return nil
	})
}

func (s *IncidentService) saveProjectedIncidentSeverity(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name(attrs.SeverityRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident severity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := s.db.Client(ctx).IncidentSeverity.Create().
		SetName(attrs.SeverityRef).
		SetRank(0).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident severity: %w", createErr)
	}
	return created.ID, nil
}

func (s *IncidentService) saveProjectedIncidentType(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentType.Query().
		Where(incidenttype.Name(attrs.TypeRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident type: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}
	created, createErr := s.db.Client(ctx).IncidentType.Create().
		SetName(attrs.TypeRef).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident type: %w", createErr)
	}
	return created.ID, nil
}
