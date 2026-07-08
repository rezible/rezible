package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeObserved          = "code_change_observed"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"
	assertionCodeChangeRelatedEntity     = "code_change_related_entity"
	assertionSystemComponentExists       = "system_component_exists"
	assertionSystemRelationshipExists    = "system_relationship_exists"

	knowledgeKindCodeRepository = "code_repository"
	knowledgeKindCodeChange     = "code_change"
	relationshipKindTouched     = "touched_repository"
)

func (s *KnowledgeFactService) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedDomainEntityRef, error) {
	proj := newKnowledgeEntityEventProjector(ev, s)
	result, eventErr := proj.projectEvent(ev)
	if eventErr != nil {
		return nil, fmt.Errorf("project event: %w", eventErr)
	}
	if len(result) > 0 {
		evErr := s.IngestProjectedEventEvidence(ctx, ev, result)
		if evErr != nil {
			return nil, fmt.Errorf("ingest projected evidence: %w", evErr)
		}
	}
	return nil, nil
}

type knowledgeEntityEventProjector struct {
	event     *ent.NormalizedEvent
	knowledge *KnowledgeFactService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, knowledge *KnowledgeFactService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{event: ev, knowledge: knowledge}
}

func (kp *knowledgeEntityEventProjector) projectEvent(ev *ent.NormalizedEvent) ([]rez.ProjectedKnowledgeEvidence, error) {
	var decErr error
	switch projections.SubjectKind(ev.SubjectKind) {
	case projections.SubjectKindCodeForge:
		{
			var obs *projections.CodeForgeEvent
			if obs, decErr = projections.DecodeCodeForgeEvent(ev); decErr != nil {
				return nil, decErr
			}
			return kp.projectCodeForgeEvent(obs), nil
		}
	case projections.SubjectKindCodeChange:
		{
			var obs *projections.CodeChangeEvent
			if obs, decErr = projections.DecodeCodeChangeEvent(ev); decErr != nil {
				return nil, decErr
			}
			return kp.projectCodeChangeEvent(obs), nil
		}
	case projections.SubjectKindSystemComponent:
		{
			var obs *projections.SystemComponentEvent
			if obs, decErr = projections.DecodeSystemComponentEvent(ev); decErr != nil {
				return nil, decErr
			}
			return kp.projectSystemComponentEvent(obs), nil
		}
	case projections.SubjectKindSystemRelationship:
		{
			var obs *projections.SystemRelationshipEvent
			if obs, decErr = projections.DecodeSystemRelationshipEvent(ev); decErr != nil {
				return nil, decErr
			}
			return kp.projectSystemRelationshipEvent(obs), nil
		}
	}
	return nil, decErr
}

// Event projections

func (kp *knowledgeEntityEventProjector) makeCodeRepositoryEntityRef(ref string, name string, desc string) *ent.KnowledgeEntityRef {
	return &ent.KnowledgeEntityRef{
		Kind:        knowledgeKindCodeRepository,
		Reference:   ref,
		DisplayName: name,
		Description: desc,
	}
}

func (kp *knowledgeEntityEventProjector) makeSubjectEntityAliasRef(desc string, subject *ent.KnowledgeEntityRef) rez.ProjectedKnowledgeEvidenceSubjectAlias {
	return rez.ProjectedKnowledgeEvidenceSubjectAlias{
		AliasRef: ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindEntity,
			Provider:           kp.event.Provider,
			ProviderSubjectRef: kp.event.ProviderSubjectRef,
		},
		Description:      desc,
		SubjectEntityRef: subject,
	}
}

func (kp *knowledgeEntityEventProjector) makeObservedEvidence(assertion string, props map[string]any, aliasRef rez.ProjectedKnowledgeEvidenceSubjectAlias) rez.ProjectedKnowledgeEvidence {
	return rez.ProjectedKnowledgeEvidence{
		Kind:         ke.EvidenceKindObserved,
		Assertion:    assertion,
		EffectiveAt:  kp.event.OccurredAt,
		Properties:   props,
		SubjectAlias: aliasRef,
	}
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) []rez.ProjectedKnowledgeEvidence {
	name := pe.Attributes.DisplayName
	repoEntityRef := kp.makeCodeRepositoryEntityRef(pe.Attributes.URL, name, "")
	repoSubjectRef := kp.makeSubjectEntityAliasRef("", repoEntityRef)
	return []rez.ProjectedKnowledgeEvidence{
		kp.makeObservedEvidence(assertionCodeRepositoryExists, nil, repoSubjectRef),
	}
}

func (kp *knowledgeEntityEventProjector) makeRelatedEntityRef(rel projections.RelatedEntityRef) *ent.KnowledgeEntityRef {
	return &ent.KnowledgeEntityRef{
		Kind:        rel.Kind,
		Reference:   rel.ExternalRef,
		DisplayName: rel.DisplayName,
	}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEvent(pe *projections.CodeChangeEvent) []rez.ProjectedKnowledgeEvidence {
	attrs := pe.Attributes
	codeChangeEntity := ent.KnowledgeEntityRef{
		Kind:        "code_change",
		Reference:   "", //pe.Attributes.ExternalRef,
		DisplayName: attrs.DisplayName,
		Description: "",
	}

	repoEntityRef := kp.makeCodeRepositoryEntityRef(attrs.RepositoryExternalRef, attrs.DisplayName, "")
	repoSubjectRef := kp.makeSubjectEntityAliasRef("", repoEntityRef)

	repoChangedEvidence := kp.makeObservedEvidence(assertionCodeChangeTouchedRepository, nil, repoSubjectRef)

	results := []rez.ProjectedKnowledgeEvidence{
		repoChangedEvidence,
	}

	for _, rel := range attrs.RelatedEntities {
		relEntityRef := kp.makeRelatedEntityRef(rel)
		relationshipAliasRef := ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindRelationship,
			Provider:           pe.Event.Provider,
			ProviderSubjectRef: pe.Event.ProviderSubjectRef,
		}
		results = append(results, rez.ProjectedKnowledgeEvidence{
			Kind:        ke.EvidenceKindObserved,
			Assertion:   assertionCodeChangeRelatedEntity,
			EffectiveAt: kp.event.OccurredAt,
			Properties:  nil,
			SubjectAlias: rez.ProjectedKnowledgeEvidenceSubjectAlias{
				Description: "",
				AliasRef:    relationshipAliasRef,
				SubjectRelationshipRef: &ent.KnowledgeRelationshipRef{
					Kind:        "code_change_impacted",
					Description: "",
					EntityRefs: [2]ent.KnowledgeEntityRef{
						codeChangeEntity,
						*relEntityRef,
					},
				},
			},
		})
	}

	return results
}

func (kp *knowledgeEntityEventProjector) makeSystemComponentExistsEvidence(ref, name, desc, kind string) rez.ProjectedKnowledgeEvidence {
	componentSubjectRef := kp.makeSubjectEntityAliasRef("", &ent.KnowledgeEntityRef{
		Kind:        "system_component",
		Reference:   ref,
		DisplayName: name,
		Description: desc,
	})
	props := map[string]any{"component_kind": kind}
	return kp.makeObservedEvidence(assertionSystemComponentExists, props, componentSubjectRef)
}

func (kp *knowledgeEntityEventProjector) projectSystemComponentEvent(pe *projections.SystemComponentEvent) []rez.ProjectedKnowledgeEvidence {
	attrs := pe.Attributes
	return []rez.ProjectedKnowledgeEvidence{
		kp.makeSystemComponentExistsEvidence(attrs.ExternalRef, attrs.DisplayName, attrs.Description, attrs.Kind),
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemRelationshipEvent(pe *projections.SystemRelationshipEvent) []rez.ProjectedKnowledgeEvidence {
	attrs := pe.Attributes

	sourceExists := kp.makeSystemComponentExistsEvidence(attrs.SourceExternalRef, attrs.SourceDisplayName, "", attrs.SourceKind)
	sourceEntityRef := *sourceExists.SubjectAlias.SubjectEntityRef

	targetExists := kp.makeSystemComponentExistsEvidence(attrs.TargetExternalRef, attrs.TargetDisplayName, "", attrs.TargetKind)
	targetEntityRef := *targetExists.SubjectAlias.SubjectEntityRef

	relationshipAlias := rez.ProjectedKnowledgeEvidenceSubjectAlias{
		Description: attrs.Description,
		AliasRef: ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindRelationship,
			Provider:           kp.event.Provider,
			ProviderSubjectRef: kp.event.ProviderSubjectRef,
		},
		SubjectRelationshipRef: &ent.KnowledgeRelationshipRef{
			Kind:        attrs.Kind,
			Description: attrs.Description,
			EntityRefs:  [2]ent.KnowledgeEntityRef{sourceEntityRef, targetEntityRef},
		},
	}
	relationshipExists := kp.makeObservedEvidence(assertionSystemRelationshipExists, nil, relationshipAlias)

	return []rez.ProjectedKnowledgeEvidence{
		sourceExists,
		targetExists,
		relationshipExists,
	}
}
