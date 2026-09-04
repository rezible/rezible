package eventprojection

import (
	"context"
	"fmt"
	"maps"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionSystemComponentExists    = "system_component_exists"
	knowledgeAssertionSystemEndpointObserved   = "system_relationship_endpoint_observed"
	knowledgeAssertionSystemRelationshipExists = "system_relationship_exists"
)

func (s *ProjectionService) handleSystemComponentEvent(ctx context.Context, e *projections.SystemComponentEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes

	properties := make(map[string]any, len(attrs.Properties)+1)
	maps.Copy(properties, attrs.Properties)
	properties["component_kind"] = attrs.Kind

	componentResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	componentEntityRef := ent.KnowledgeEntityRef{
		Category:            attrs.Category,
		Kind:                attrs.Kind,
		ProviderResourceRef: componentResourceRef,
	}
	evidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionSystemComponentExists,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.DisplayName,
			Description: attrs.Description,
			Properties:  properties,
		},
		SubjectEntity: &componentEntityRef,
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest system component evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleSystemRelationshipEvent(ctx context.Context, e *projections.SystemRelationshipEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes

	sourceEntityRef := ent.KnowledgeEntityRef{
		Category:            attrs.Source.Category,
		Kind:                attrs.Source.Kind,
		ProviderResourceRef: attrs.Source.Ref,
	}

	targetEntityRef := ent.KnowledgeEntityRef{
		Category:            attrs.Target.Category,
		Kind:                attrs.Target.Kind,
		ProviderResourceRef: attrs.Target.Ref,
	}
	evidenceKind := projectionEvidenceKind(event)
	sourceEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionSystemEndpointObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Source.DisplayName,
			Description: attrs.Source.Description,
			Properties:  attrs.Source.Properties,
		},
		SubjectEntity: &sourceEntityRef,
	}
	targetEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionSystemEndpointObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Target.DisplayName,
			Description: attrs.Target.Description,
			Properties:  attrs.Target.Properties,
		},
		SubjectEntity: &targetEntityRef,
	}

	relationshipResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	relationshipRef := ent.KnowledgeRelationshipRef{
		Predicate:           attrs.Predicate,
		ProviderResourceRef: relationshipResourceRef,
		Source:              sourceEntityRef,
		Target:              targetEntityRef,
	}
	relationshipEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:                evidenceKind,
		Assertion:           knowledgeAssertionSystemRelationshipExists,
		EffectiveAt:         event.OccurredAt,
		SubjectRelationship: &relationshipRef,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.DisplayName,
			Description: attrs.Description,
			Properties:  attrs.Properties,
		},
	}
	if ingestErr := s.knowledge.IngestEvidence(ctx, event, sourceEvidenceRef, targetEvidenceRef, relationshipEvidenceRef); ingestErr != nil {
		return nil, fmt.Errorf("ingest system relationship evidence: %w", ingestErr)
	}
	return nil, nil
}
