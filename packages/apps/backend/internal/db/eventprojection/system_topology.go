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
	knowledgeAssertionSystemRelationshipExists = "system_relationship_exists"
)

func (s *ProjectionService) handleSystemComponentEvent(ctx context.Context, e *projections.SystemComponentEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes

	properties := make(map[string]any, len(attrs.Properties)+1)
	maps.Copy(properties, attrs.Properties)
	properties["component_kind"] = attrs.Kind

	evidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionSystemComponentExists,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.DisplayName,
			Description: attrs.Description,
			Properties:  properties,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Category:        attrs.Category,
			Kind:            attrs.Kind,
			SubjectAliasRef: event.KnowledgeSubjectAliasRef(),
		},
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
		Category: attrs.SourceCategory,
		Kind:     attrs.SourceKind,
		SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
			Provider:           event.Provider,
			ProviderSource:     event.ProviderSource,
			ProviderSubjectRef: attrs.SourceExternalRef,
		},
	}

	targetEntityRef := ent.KnowledgeEntityRef{
		Category: attrs.TargetCategory,
		Kind:     attrs.TargetKind,
		SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
			Provider:           event.Provider,
			ProviderSource:     event.ProviderSource,
			ProviderSubjectRef: attrs.TargetExternalRef,
		},
	}

	relationshipEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionSystemRelationshipExists,
		EffectiveAt: event.OccurredAt,
		SubjectRelationship: &ent.KnowledgeRelationshipRef{
			Predicate:       attrs.Predicate,
			SubjectAliasRef: event.KnowledgeSubjectAliasRef(),
			Source:          sourceEntityRef,
			Target:          targetEntityRef,
		},
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.DisplayName,
			Description: attrs.Description,
			Properties:  attrs.Properties,
		},
	}
	if ingestErr := s.knowledge.IngestEvidence(ctx, event, relationshipEvidenceRef); ingestErr != nil {
		return nil, fmt.Errorf("ingest system relationship evidence: %w", ingestErr)
	}
	return nil, nil
}
