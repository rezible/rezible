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
	evidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeEvidenceSubjectState{
			DisplayName: event.Attributes.DisplayName,
			Properties:  properties,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:  knowledgeEntityKindCodeRepository,
			Alias: event.Event.KnowledgeAliasRef(),
		},
	}

	if _, ingestErr := s.knowledge.IngestEntityEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest code repository evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleCodeChangeEvent(ctx context.Context, event *projections.CodeChangeEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	evidenceKind := projectionEvidenceKind(event.Event)

	changeRef := ent.KnowledgeEntityRef{
		Kind:  knowledgeEntityKindCodeChange,
		Alias: event.Event.KnowledgeAliasRef(),
	}
	codeChangeEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeChangeObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeEvidenceSubjectState{
			DisplayName: attributes.DisplayName,
		},
		SubjectEntity: &changeRef,
	}

	repoAlias := event.Event.KnowledgeAliasRef()
	repoAlias.ProviderSubjectRef = attributes.RepositoryExternalRef
	repositoryRef := ent.KnowledgeEntityRef{
		Kind:  knowledgeEntityKindCodeRepository,
		Alias: repoAlias,
	}
	repoEntityEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeEvidenceSubjectState{
			DisplayName: attributes.RepositoryExternalRef,
		},
		SubjectEntity: &repositoryRef,
	}

	codeChangeRepoEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeChangeRepository,
		EffectiveAt: event.Event.OccurredAt,
		SubjectRelationship: &ent.KnowledgeRelationshipRef{
			Kind: knowledgeRelationshipKindTouchedRepository,
			Alias: ent.KnowledgeAliasRef{
				Provider:           event.Event.Provider,
				ProviderSource:     event.Event.ProviderSource,
				ProviderSubjectRef: fmt.Sprintf("change:%s:%s", event.Event.ProviderSubjectRef, attributes.RepositoryExternalRef),
			},
			Source: changeRef,
			Target: repositoryRef,
		},
	}

	evidence := []ent.KnowledgeEvidenceRef{
		codeChangeEvidence,
		repoEntityEvidence,
		codeChangeRepoEvidence,
	}

	for _, related := range projections.SortRelatedEntityRefs(attributes.RelatedEntities) {
		evidence = append(evidence, ent.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionCodeChangeImpact,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeEvidenceSubjectState{
				DisplayName: related.DisplayName,
				Properties:  map[string]any{"component_kind": related.Kind},
			},
			SubjectRelationship: &ent.KnowledgeRelationshipRef{
				Kind: knowledgeRelationshipKindChangeImpacted,
				Alias: ent.KnowledgeAliasRef{
					Provider:           event.Event.Provider,
					ProviderSource:     event.Event.ProviderSource,
					ProviderSubjectRef: fmt.Sprintf("impacted:%s:%s", event.Event.ProviderSubjectRef, related.ExternalRef),
				},
				Source: changeRef,
				Target: ent.KnowledgeEntityRef{
					Kind: knowledgeEntityKindSystemComponent,
					Alias: ent.KnowledgeAliasRef{
						Provider:           event.Event.Provider,
						ProviderSource:     event.Event.ProviderSource,
						ProviderSubjectRef: related.ExternalRef,
					},
				},
			},
		})
	}

	if _, ingestErr := s.knowledge.IngestEvidenceBulk(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest code change evidence: %w", ingestErr)
	}
	return nil, nil
}
