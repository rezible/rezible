package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindSystemComponent = "system_component"

	knowledgeAssertionSystemComponentObserved    = "system_component_exists"
	knowledgeAssertionSystemRelationshipObserved = "system_relationship_exists"
)

func systemComponentRef(kind, reference, displayName, description string, properties map[string]any) ent.KnowledgeEntityRef {
	componentProperties := make(map[string]any, len(properties)+1)
	for key, value := range properties {
		componentProperties[key] = value
	}
	componentProperties["component_kind"] = kind
	return ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindSystemComponent,
		Reference:   reference,
		DisplayName: displayName,
		Description: description,
		Properties:  componentProperties,
	}
}

func (s *ProjectionService) handleSystemComponentEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeSystemComponentEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid system component event: %w", decodeErr)
	}
	attributes := event.Attributes
	properties := make(map[string]any, len(attributes.Properties)+1)
	for key, value := range attributes.Properties {
		properties[key] = value
	}
	properties["component_kind"] = attributes.Kind

	entityRef := &ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindSystemComponent,
		Reference:   attributes.ExternalRef,
		DisplayName: attributes.DisplayName,
		Description: attributes.Description,
		Properties:  properties,
	}
	aliasRef := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
	aliasRef.SubjectEntityRef = entityRef
	evidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionSystemComponentObserved,
		EffectiveAt:     event.Event.OccurredAt,
		Properties:      properties,
		SubjectAliasRef: aliasRef,
	}

	if ingestErr := s.ingestProjectedEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest system component evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleSystemRelationshipEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeSystemRelationshipEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid system relationship event: %w", decodeErr)
	}
	attributes := event.Attributes
	sourceRef := systemComponentRef(
		attributes.SourceKind,
		attributes.SourceExternalRef,
		attributes.SourceDisplayName,
		"",
		nil,
	)
	targetRef := systemComponentRef(
		attributes.TargetKind,
		attributes.TargetExternalRef,
		attributes.TargetDisplayName,
		"",
		nil,
	)
	relationshipRef := &ent.KnowledgeRelationshipRef{
		Kind:        attributes.Kind,
		Description: attributes.Description,
		Properties:  attributes.Properties,
		EntityRefs:  [2]ent.KnowledgeEntityRef{sourceRef, targetRef},
	}

	sourceAlias := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           event.Event.Provider,
		ProviderSubjectRef: attributes.SourceExternalRef,
		Description:        "System relationship source",
		SubjectEntityRef:   &sourceRef,
	}
	targetAlias := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           event.Event.Provider,
		ProviderSubjectRef: attributes.TargetExternalRef,
		Description:        "System relationship target",
		SubjectEntityRef:   &targetRef,
	}
	relationshipAlias := event.Event.MakeSubjectAliasRef(ksa.SubjectKindRelationship, "System relationship")
	relationshipAlias.SubjectRelationshipRef = relationshipRef

	evidence := []ent.KnowledgeEvidenceRef{
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionSystemComponentObserved,
			EffectiveAt:     event.Event.OccurredAt,
			Properties:      sourceRef.Properties,
			SubjectAliasRef: sourceAlias,
		},
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionSystemComponentObserved,
			EffectiveAt:     event.Event.OccurredAt,
			Properties:      targetRef.Properties,
			SubjectAliasRef: targetAlias,
		},
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionSystemRelationshipObserved,
			EffectiveAt:     event.Event.OccurredAt,
			Properties:      attributes.Properties,
			SubjectAliasRef: relationshipAlias,
		},
	}

	if ingestErr := s.ingestProjectedEvidence(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest system relationship evidence: %w", ingestErr)
	}
	return nil, nil
}
