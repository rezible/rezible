package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionSystemComponentExists    = "system_component_exists"
	knowledgeAssertionSystemRelationshipExists = "system_relationship_exists"
)

func (s *ProjectionService) handleSystemComponentEvent(ctx context.Context, event *projections.SystemComponentEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	properties := make(map[string]any, len(attributes.Properties)+1)
	for key, value := range attributes.Properties {
		properties[key] = value
	}
	properties["component_subkind"] = attributes.Subkind

	evidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionSystemComponentExists,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.DisplayName,
			Description: attributes.Description,
			Properties:  properties,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:    attributes.Kind,
			Subkind: attributes.Subkind,
			Alias:   event.Event.KnowledgeAliasRef(),
		},
	}

	if _, ingestErr := s.knowledge.IngestEntityEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest system component evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleSystemRelationshipEvent(ctx context.Context, event *projections.SystemRelationshipEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	evidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionSystemRelationshipExists,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.DisplayName,
			Description: attributes.Description,
			Properties:  attributes.Properties,
		},
		SubjectRelationship: &ent.KnowledgeRelationshipRef{
			Kind:    attributes.Kind,
			Subkind: attributes.Subkind,
			Alias:   event.Event.KnowledgeAliasRef(),
			Source: ent.KnowledgeEntityRef{
				Kind:    attributes.SourceKind,
				Subkind: attributes.SourceSubkind,
				Alias: ent.KnowledgeAliasRef{
					Provider:           event.Event.Provider,
					ProviderSource:     event.Event.ProviderSource,
					ProviderSubjectRef: attributes.SourceExternalRef,
				},
			},
			Target: ent.KnowledgeEntityRef{
				Kind:    attributes.TargetKind,
				Subkind: attributes.TargetSubkind,
				Alias: ent.KnowledgeAliasRef{
					Provider:           event.Event.Provider,
					ProviderSource:     event.Event.ProviderSource,
					ProviderSubjectRef: attributes.TargetExternalRef,
				},
			},
		},
	}
	if _, ingestErr := s.knowledge.IngestEvidenceBulk(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest system relationship evidence: %w", ingestErr)
	}
	return nil, nil
}
