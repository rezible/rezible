package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionCodeRepositoryObserved = "code_repository_exists"
	knowledgeAssertionCodeChangeObserved     = "code_change_observed"
	knowledgeAssertionCodeChangeRepository   = "code_change_touched_repository"
	knowledgeAssertionCodeEntityObserved     = "code_change_related_entity_observed"
	knowledgeAssertionCodeChangeImpact       = "code_change_related_entity"
)

func (s *ProjectionService) handleCodeForgeEvent(ctx context.Context, event *projections.CodeForgeEvent) ([]rez.ProjectedEntityRef, error) {
	properties := make(map[string]any)
	if event.Attributes.URL != "" {
		properties["url"] = event.Attributes.URL
	}
	repositoryResourceRef := rez.ProviderResourceRef{
		Provider:          event.Event.Provider,
		ProviderNamespace: event.Event.ProviderNamespace,
		ResourceRef:       event.Event.ProviderResourceRef,
	}
	repositoryEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryCode,
		Kind:                knowledgeEntityKindRepository,
		ProviderResourceRef: repositoryResourceRef,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: event.Attributes.DisplayName,
			Properties:  properties,
		},
		SubjectEntity: &repositoryEntityRef,
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest code repository evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleCodeChangeEvent(ctx context.Context, event *projections.CodeChangeEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	evidenceKind := projectionEvidenceKind(event.Event)

	changeResourceRef := rez.ProviderResourceRef{
		Provider:          event.Event.Provider,
		ProviderNamespace: event.Event.ProviderNamespace,
		ResourceRef:       event.Event.ProviderResourceRef,
	}
	changeRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryEvent,
		Kind:                knowledgeEntityKindCodeChange,
		ProviderResourceRef: changeResourceRef,
	}
	codeChangeEvidence := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeChangeObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.DisplayName,
		},
		SubjectEntity: &changeRef,
	}

	repositoryRef := rez.KnowledgeEntityRef{
		Category:            attributes.Repository.Category,
		Kind:                attributes.Repository.Kind,
		ProviderResourceRef: attributes.Repository.Ref,
	}
	repoEntityEvidence := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Repository.DisplayName,
			Description: attributes.Repository.Description,
			Properties:  attributes.Repository.Properties,
		},
		SubjectEntity: &repositoryRef,
	}

	changeRepositoryResourceRef := projections.DerivedRelationshipRef(knr.PredicateTouches, changeResourceRef, attributes.Repository.Ref)
	changeRepositoryRelationship := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateTouches,
		ProviderResourceRef: changeRepositoryResourceRef,
		Source:              changeRef,
		Target:              repositoryRef,
	}
	codeChangeRepoEvidence := rez.KnowledgeEvidenceRef{
		Kind:                evidenceKind,
		Assertion:           knowledgeAssertionCodeChangeRepository,
		EffectiveAt:         event.Event.OccurredAt,
		SubjectRelationship: &changeRepositoryRelationship,
	}

	evidence := []rez.KnowledgeEvidenceRef{codeChangeEvidence, repoEntityEvidence, codeChangeRepoEvidence}

	for _, related := range projections.SortEntityObservations(attributes.ImpactedEntities) {
		relatedEntityRef := rez.KnowledgeEntityRef{
			Category:            related.Category,
			Kind:                related.Kind,
			ProviderResourceRef: related.Ref,
		}
		relatedEntityEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionCodeEntityObserved,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: related.DisplayName,
				Description: related.Description,
				Properties:  related.Properties,
			},
			SubjectEntity: &relatedEntityRef,
		}
		impactResourceRef := projections.DerivedRelationshipRef(knr.PredicateImpacts, changeResourceRef, related.Ref)
		impactRelationship := rez.KnowledgeRelationshipRef{
			Predicate:           knr.PredicateImpacts,
			ProviderResourceRef: impactResourceRef,
			Source:              changeRef,
			Target:              relatedEntityRef,
		}
		impactEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionCodeChangeImpact,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: related.DisplayName,
				Properties:  map[string]any{"entity_kind": related.Kind},
			},
			SubjectRelationship: &impactRelationship,
		}
		evidence = append(evidence, relatedEntityEvidence, impactEvidence)
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest code change evidence: %w", ingestErr)
	}
	return nil, nil
}
