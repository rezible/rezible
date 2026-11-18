package eventprojection

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/pkg/knowledgegraph"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionCodeRepositoryObserved = "code_repository_exists"
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
		Kind:                knowledgegraph.EntityKindRepository,
		ProviderResourceRef: repositoryResourceRef,
		LinkingAttributes:   event.Attributes.LinkingAttributes,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionCodeRepositoryObserved,
		StateMode:   ke.StateModeReplace,
		EffectiveAt: event.Event.OccurredAt,
		Authority:   schematypes.KnowledgeAuthorityDeclared,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: event.Attributes.DisplayName,
			Properties:  properties,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &repositoryEntityRef},
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence); ingestErr != nil {
		return nil, fmt.Errorf("ingest code repository evidence: %w", ingestErr)
	}
	return nil, nil
}

func (s *ProjectionService) handleCodeChangeEvent(ctx context.Context, event *projections.CodeChangeEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	evidenceKind := projectionEvidenceKind(event.Event)

	repositoryRef := attributes.Repository.KnowledgeEntityRef()
	evidence := []rez.KnowledgeEvidenceRef{{
		Kind:         evidenceKind,
		Assertion:    projections.AssertionChangeRepository,
		EffectiveAt:  event.Event.OccurredAt,
		Authority:    schematypes.KnowledgeAuthorityObserved,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: attributes.Repository.DisplayName},
		Subject:      rez.KnowledgeSubjectRef{Entity: &repositoryRef},
	}}
	for _, related := range projections.SortEntityObservations(attributes.ImpactedEntities) {
		relatedRef := related.KnowledgeEntityRef()
		evidence = append(evidence, rez.KnowledgeEvidenceRef{
			Kind:         evidenceKind,
			Assertion:    projections.AssertionChangeImpacted,
			EffectiveAt:  event.Event.OccurredAt,
			Authority:    schematypes.KnowledgeAuthorityObserved,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: related.DisplayName},
			Subject:      rez.KnowledgeSubjectRef{Entity: &relatedRef},
		})
	}
	if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("ingest code change involvement: %w", ingestErr)
	}
	if attributes.Merge != nil {
		if recordErr := s.changePoints.RecordChangePoint(ctx, event.Event, attributes.Merge.MergedAt); recordErr != nil {
			return nil, fmt.Errorf("record merged change: %w", recordErr)
		}
	}
	return nil, nil
}
