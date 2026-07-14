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
	knowledgeEntityKindSystemComponent = "system_component"
	assertionSystemComponentExists     = "system_component_exists"
	assertionSystemRelationshipExists  = "system_relationship_exists"

	knowledgeEntityKindCodeRepository    = "code_repository"
	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"
	assertionCodeChangeRelatedEntity     = "code_change_related_entity"
)

func (s *KnowledgeIngestionService) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	proj := newKnowledgeEntityEventProjector(ev, s)
	result, eventErr := proj.projectEvent(ev)
	if eventErr != nil {
		return nil, fmt.Errorf("project event: %w", eventErr)
	}
	if len(result) > 0 {
		evErr := s.IngestProjectedEvidence(ctx, ev, result...)
		if evErr != nil {
			return nil, fmt.Errorf("ingest projected evidence: %w", evErr)
		}
	}
	return nil, nil
}

type knowledgeEntityEventProjector struct {
	event     *ent.NormalizedEvent
	knowledge *KnowledgeIngestionService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, knowledge *KnowledgeIngestionService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{event: ev, knowledge: knowledge}
}

func (kp *knowledgeEntityEventProjector) projectEvent(ev *ent.NormalizedEvent) ([]ent.KnowledgeEvidenceRef, error) {
	var decErr error
	switch projections.SubjectKind(ev.SubjectKind) {
	case projections.SubjectKindCodeForge:
		{
			var obs *projections.CodeForgeEvent
			if obs, decErr = projections.DecodeCodeForgeEvent(ev); decErr == nil {
				return kp.projectCodeForgeEvent(obs), nil
			}
		}
	case projections.SubjectKindCodeChange:
		{
			var obs *projections.CodeChangeEvent
			if obs, decErr = projections.DecodeCodeChangeEvent(ev); decErr == nil {
				return kp.projectCodeChangeEvent(obs), nil
			}
		}
	case projections.SubjectKindSystemComponent:
		{
			var obs *projections.SystemComponentEvent
			if obs, decErr = projections.DecodeSystemComponentEvent(ev); decErr == nil {
				return kp.projectSystemComponentEvent(obs), nil
			}
		}
	case projections.SubjectKindSystemRelationship:
		{
			var obs *projections.SystemRelationshipEvent
			if obs, decErr = projections.DecodeSystemRelationshipEvent(ev); decErr == nil {
				return kp.projectSystemRelationshipEvent(obs), nil
			}
		}
	}
	return nil, decErr
}

// Event projections

func (kp *knowledgeEntityEventProjector) makeEntityRef(kind string, ref string, name string, desc string) *ent.KnowledgeEntityRef {
	return &ent.KnowledgeEntityRef{
		Kind:        kind,
		Reference:   ref,
		DisplayName: name,
		Description: desc,
	}
}

func (kp *knowledgeEntityEventProjector) makeRelationshipRef(kind string, desc string, sr, tr ent.KnowledgeEntityRef) *ent.KnowledgeRelationshipRef {
	return &ent.KnowledgeRelationshipRef{
		Kind:        kind,
		Description: desc,
		EntityRefs:  [2]ent.KnowledgeEntityRef{sr, tr},
	}
}

func (kp *knowledgeEntityEventProjector) makeSubjectEntityAliasRef(ser *ent.KnowledgeEntityRef, desc string) ent.KnowledgeSubjectAliasRef {
	return ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           kp.event.Provider,
		ProviderSubjectRef: kp.event.ProviderSubjectRef,
		SubjectEntityRef:   ser,
		Description:        desc,
	}
}
func (kp *knowledgeEntityEventProjector) makeSubjectRelationshipAliasRef(srr *ent.KnowledgeRelationshipRef, desc string) ent.KnowledgeSubjectAliasRef {
	return ent.KnowledgeSubjectAliasRef{
		Kind:                   ksa.SubjectKindRelationship,
		Provider:               kp.event.Provider,
		ProviderSubjectRef:     kp.event.ProviderSubjectRef,
		Description:            desc,
		SubjectRelationshipRef: srr,
	}
}

func (kp *knowledgeEntityEventProjector) makeObservedEvidence(assertion string, props map[string]any, aliasRef ent.KnowledgeSubjectAliasRef) ent.KnowledgeEvidenceRef {
	return ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       assertion,
		EffectiveAt:     kp.event.OccurredAt,
		Properties:      props,
		SubjectAliasRef: aliasRef,
	}
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) []ent.KnowledgeEvidenceRef {
	name := pe.Attributes.DisplayName
	repoEntityRef := kp.makeEntityRef(knowledgeEntityKindCodeRepository, pe.Attributes.URL, name, "")
	repoSubjectRef := kp.makeSubjectEntityAliasRef(repoEntityRef, "")
	return []ent.KnowledgeEvidenceRef{
		kp.makeObservedEvidence(assertionCodeRepositoryExists, nil, repoSubjectRef),
	}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEvent(pe *projections.CodeChangeEvent) []ent.KnowledgeEvidenceRef {
	attrs := pe.Attributes
	codeChangeEntity := ent.KnowledgeEntityRef{
		Kind:        "code_change",
		Reference:   "", //pe.Attributes.ExternalRef,
		DisplayName: attrs.DisplayName,
		Description: "",
	}

	repoEntityRef := kp.makeEntityRef(knowledgeEntityKindCodeRepository, attrs.RepositoryExternalRef, attrs.DisplayName, "")
	repoSubjectRef := kp.makeSubjectEntityAliasRef(repoEntityRef, "")

	repoChangedEvidence := kp.makeObservedEvidence(assertionCodeChangeTouchedRepository, nil, repoSubjectRef)

	results := []ent.KnowledgeEvidenceRef{
		repoChangedEvidence,
	}

	for _, rel := range attrs.RelatedEntities {
		entityRef := kp.makeEntityRef(rel.Kind, rel.ExternalRef, rel.DisplayName, "")
		relationshipRef := kp.makeRelationshipRef("code_change_impacted", "", codeChangeEntity, *entityRef)
		results = append(results, ent.KnowledgeEvidenceRef{
			Kind:            ke.EvidenceKindObserved,
			Assertion:       assertionCodeChangeRelatedEntity,
			SubjectAliasRef: kp.makeSubjectRelationshipAliasRef(relationshipRef, ""),
			EffectiveAt:     kp.event.OccurredAt,
			Properties:      nil,
		})
	}

	return results
}

func (kp *knowledgeEntityEventProjector) makeSystemComponentExistsEvidence(kind, ref, name, desc string) ent.KnowledgeEvidenceRef {
	componentRef := kp.makeEntityRef(knowledgeEntityKindSystemComponent, ref, name, desc)
	componentSubjectRef := kp.makeSubjectEntityAliasRef(componentRef, "")
	props := map[string]any{"component_kind": kind}
	return kp.makeObservedEvidence(assertionSystemComponentExists, props, componentSubjectRef)
}

func (kp *knowledgeEntityEventProjector) projectSystemComponentEvent(pe *projections.SystemComponentEvent) []ent.KnowledgeEvidenceRef {
	attrs := pe.Attributes
	return []ent.KnowledgeEvidenceRef{
		kp.makeSystemComponentExistsEvidence(attrs.Kind, attrs.ExternalRef, attrs.DisplayName, attrs.Description),
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemRelationshipEvent(pe *projections.SystemRelationshipEvent) []ent.KnowledgeEvidenceRef {
	attrs := pe.Attributes

	sourceExists := kp.makeSystemComponentExistsEvidence(attrs.SourceKind, attrs.SourceExternalRef, attrs.SourceDisplayName, "")
	sourceEntityRef := *sourceExists.SubjectAliasRef.SubjectEntityRef

	targetExists := kp.makeSystemComponentExistsEvidence(attrs.TargetKind, attrs.TargetExternalRef, attrs.TargetDisplayName, "")
	targetEntityRef := *targetExists.SubjectAliasRef.SubjectEntityRef

	relationshipAlias := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindRelationship,
		Provider:           kp.event.Provider,
		ProviderSubjectRef: kp.event.ProviderSubjectRef,
		Description:        attrs.Description,
		SubjectRelationshipRef: &ent.KnowledgeRelationshipRef{
			Kind:        attrs.Kind,
			Description: attrs.Description,
			EntityRefs:  [2]ent.KnowledgeEntityRef{sourceEntityRef, targetEntityRef},
		},
	}
	relationshipExists := kp.makeObservedEvidence(assertionSystemRelationshipExists, nil, relationshipAlias)

	return []ent.KnowledgeEvidenceRef{
		sourceExists,
		targetExists,
		relationshipExists,
	}
}
