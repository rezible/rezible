package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
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
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: event.Attributes.DisplayName,
			Properties:  properties,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:            kne.KindCode,
			Subkind:         knowledgeEntitySubkindRepository,
			SubjectAliasRef: event.Event.KnowledgeSubjectAliasRef(),
		},
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest code repository evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleCodeChangeEvent(ctx context.Context, event *projections.CodeChangeEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	evidenceKind := projectionEvidenceKind(event.Event)

	changeRef := ent.KnowledgeEntityRef{
		Kind:            kne.KindEvent,
		Subkind:         knowledgeEntitySubkindCodeChange,
		SubjectAliasRef: event.Event.KnowledgeSubjectAliasRef(),
	}
	codeChangeEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeChangeObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.DisplayName,
		},
		SubjectEntity: &changeRef,
	}

	repoAlias := event.Event.KnowledgeSubjectAliasRef()
	repoAlias.ProviderSubjectRef = attributes.RepositoryExternalRef
	repositoryRef := ent.KnowledgeEntityRef{
		Kind:            kne.KindCode,
		Subkind:         knowledgeEntitySubkindRepository,
		SubjectAliasRef: repoAlias,
	}
	repoEntityEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.RepositoryExternalRef,
		},
		SubjectEntity: &repositoryRef,
	}

	codeChangeRepoEvidence := ent.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionCodeChangeRepository,
		EffectiveAt: event.Event.OccurredAt,
		SubjectRelationship: &ent.KnowledgeRelationshipRef{
			Kind:    knr.KindImpacts,
			Subkind: knowledgeRelationshipSubkindTouchedRepository,
			SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
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
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: related.DisplayName,
				Properties:  map[string]any{"entity_subkind": related.Subkind},
			},
			SubjectRelationship: &ent.KnowledgeRelationshipRef{
				Kind:    knr.KindImpacts,
				Subkind: knowledgeRelationshipSubkindCodeChangeImpacted,
				SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
					Provider:           event.Event.Provider,
					ProviderSource:     event.Event.ProviderSource,
					ProviderSubjectRef: fmt.Sprintf("impacted:%s:%s", event.Event.ProviderSubjectRef, related.ExternalRef),
				},
				Source: changeRef,
				Target: ent.KnowledgeEntityRef{
					Kind:    related.Kind,
					Subkind: related.Subkind,
					SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
						Provider:           event.Event.Provider,
						ProviderSource:     event.Event.ProviderSource,
						ProviderSubjectRef: related.ExternalRef,
					},
				},
			},
		})
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest code change evidence: %w", ingestErr)
	}
	return nil, nil
}
