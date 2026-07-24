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
	// TODO
	return nil, nil
}
