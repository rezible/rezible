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
	knowledgeEntityKindCodeRepository = "code_repository"
	knowledgeEntityKindCodeChange     = "code_change"

	knowledgeRelationshipKindTouchedRepository = "touched_repository"
	knowledgeRelationshipKindChangeImpacted    = "code_change_impacted"

	knowledgeAssertionCodeRepositoryObserved = "code_repository_exists"
	knowledgeAssertionCodeChangeObserved     = "code_change_observed"
	knowledgeAssertionCodeChangeRepository   = "code_change_touched_repository"
	knowledgeAssertionCodeChangeImpact       = "code_change_related_entity"
)

func (s *ProjectionService) handleCodeForgeEvent(ctx context.Context, event *projections.CodeForgeEvent) ([]rez.ProjectedEntityRef, error) {
	properties := make(map[string]any)
	if event.Attributes.URL != "" {
		properties["url"] = event.Attributes.URL
	}
	entityRef := &ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindCodeRepository,
		Reference:   event.Event.ProviderSubjectRef,
		DisplayName: event.Attributes.DisplayName,
		Properties:  properties,
	}
	aliasRef := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "Code repository")
	aliasRef.SubjectEntityRef = entityRef
	evidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt:     event.Event.OccurredAt,
		Properties:      entityRef.Properties,
		SubjectAliasRef: aliasRef,
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest code repository evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleCodeChangeEvent(ctx context.Context, event *projections.CodeChangeEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	changeRef := ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindCodeChange,
		Reference:   event.Event.ProviderSubjectRef,
		DisplayName: attributes.DisplayName,
		Properties: map[string]any{
			"repository_external_ref": attributes.RepositoryExternalRef,
		},
	}
	changeAlias := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "Code change")
	changeAlias.SubjectEntityRef = &changeRef

	repositoryRef := ent.KnowledgeEntityRef{
		Kind:      knowledgeEntityKindCodeRepository,
		Reference: attributes.RepositoryExternalRef,
	}
	repositoryEntityAlias := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           event.Event.Provider,
		ProviderSubjectRef: attributes.RepositoryExternalRef,
		Description:        "Code repository",
		SubjectEntityRef:   &repositoryRef,
	}
	repositoryRelationship := &ent.KnowledgeRelationshipRef{
		Kind:       knowledgeRelationshipKindTouchedRepository,
		EntityRefs: [2]ent.KnowledgeEntityRef{changeRef, repositoryRef},
	}
	repositoryAlias := ent.KnowledgeSubjectAliasRef{
		Kind:                   ksa.SubjectKindRelationship,
		Provider:               event.Event.Provider,
		ProviderSubjectRef:     fmt.Sprintf("%s:touched_repository:%s", event.Event.ProviderSubjectRef, attributes.RepositoryExternalRef),
		Description:            "Code change touched repository",
		SubjectRelationshipRef: repositoryRelationship,
	}

	evidence := []ent.KnowledgeEvidenceRef{
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionCodeChangeObserved,
			EffectiveAt:     event.Event.OccurredAt,
			Properties:      changeRef.Properties,
			SubjectAliasRef: changeAlias,
		},
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionCodeRepositoryObserved,
			EffectiveAt:     event.Event.OccurredAt,
			SubjectAliasRef: repositoryEntityAlias,
		},
		{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionCodeChangeRepository,
			EffectiveAt:     event.Event.OccurredAt,
			SubjectAliasRef: repositoryAlias,
		},
	}

	for _, related := range attributes.RelatedEntities {
		componentRef := ent.KnowledgeEntityRef{
			Kind:        knowledgeEntityKindSystemComponent,
			Reference:   related.ExternalRef,
			DisplayName: related.DisplayName,
			Properties:  map[string]any{"component_kind": related.Kind},
		}
		componentAlias := ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindEntity,
			Provider:           event.Event.Provider,
			ProviderSubjectRef: related.ExternalRef,
			Description:        "Code change related component",
			SubjectEntityRef:   &componentRef,
		}
		relationshipRef := &ent.KnowledgeRelationshipRef{
			Kind:       knowledgeRelationshipKindChangeImpacted,
			EntityRefs: [2]ent.KnowledgeEntityRef{changeRef, componentRef},
		}
		relationshipAlias := ent.KnowledgeSubjectAliasRef{
			Kind:                   ksa.SubjectKindRelationship,
			Provider:               event.Event.Provider,
			ProviderSubjectRef:     fmt.Sprintf("%s:code_change_impacted:%s", event.Event.ProviderSubjectRef, related.ExternalRef),
			Description:            "Code change impacted component",
			SubjectRelationshipRef: relationshipRef,
		}
		evidence = append(evidence,
			ent.KnowledgeEvidenceRef{
				Kind:            projectionEvidenceKind(event.Event),
				Assertion:       knowledgeAssertionSystemComponentExists,
				EffectiveAt:     event.Event.OccurredAt,
				Properties:      componentRef.Properties,
				SubjectAliasRef: componentAlias,
			},
			ent.KnowledgeEvidenceRef{
				Kind:            projectionEvidenceKind(event.Event),
				Assertion:       knowledgeAssertionCodeChangeImpact,
				EffectiveAt:     event.Event.OccurredAt,
				SubjectAliasRef: relationshipAlias,
			},
		)
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest code change evidence: %w", ingestErr)
	}
	return nil, nil
}
