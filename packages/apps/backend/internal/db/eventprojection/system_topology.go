package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
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

	componentResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	componentEntityRef := rez.KnowledgeEntityRef{
		Category:            attrs.Category,
		Kind:                attrs.Kind,
		ProviderResourceRef: componentResourceRef,
		LinkingAttributes:   attrs.LinkingAttributes,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionSystemComponentExists,
		EffectiveAt: event.OccurredAt,
		StateMode:   ke.StateModeReplace,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.DisplayName,
			Description: attrs.Description,
			Properties:  attrs.Properties,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &componentEntityRef},
	}
	if attrs.Listing {
		evidence.StateMode = ke.StateModeMerge
		evidence.SubjectState = schematypes.KnowledgeGraphSubjectState{DisplayName: attrs.DisplayName}
		evidence.ValidFor = attrs.ValidFor
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest system component evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleSystemRelationshipEvent(ctx context.Context, e *projections.SystemRelationshipEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes

	sourceEntityRef := attrs.Source.KnowledgeEntityRef()
	targetEntityRef := attrs.Target.KnowledgeEntityRef()
	evidenceKind := projectionEvidenceKind(event)
	sourceEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionSystemEndpointObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Source.DisplayName,
			Description: attrs.Source.Description,
			Properties:  attrs.Source.Properties,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &sourceEntityRef},
	}
	targetEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionSystemEndpointObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Target.DisplayName,
			Description: attrs.Target.Description,
			Properties:  attrs.Target.Properties,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &targetEntityRef},
	}

	relationshipResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	relationshipRef := rez.KnowledgeRelationshipRef{
		Predicate:           attrs.Predicate,
		ProviderResourceRef: relationshipResourceRef,
		Source:              sourceEntityRef,
		Target:              targetEntityRef,
	}
	relationshipEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionSystemRelationshipExists,
		EffectiveAt: event.OccurredAt,
		StateMode:   ke.StateModeReplace,
		Subject:     rez.KnowledgeSubjectRef{Relationship: &relationshipRef},
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
