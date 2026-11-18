package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionService) handleDeploymentEvent(ctx context.Context, e *projections.DeploymentEvent) ([]rez.ProjectedEntityRef, error) {
	serviceRef := e.Attributes.Service.KnowledgeEntityRef()
	evidence := []rez.KnowledgeEvidenceRef{{
		Kind:         projectionEvidenceKind(e.Event),
		Assertion:    projections.AssertionDeploymentService,
		EffectiveAt:  e.Event.OccurredAt,
		Authority:    schematypes.KnowledgeAuthorityObserved,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: e.Attributes.Service.DisplayName},
		Subject:      rez.KnowledgeSubjectRef{Entity: &serviceRef},
	}}
	if repository := e.Attributes.Repository; repository != nil {
		repositoryRef := repository.KnowledgeEntityRef()
		evidence = append(evidence, rez.KnowledgeEvidenceRef{
			Kind:         projectionEvidenceKind(e.Event),
			Assertion:    projections.AssertionDeploymentRepository,
			EffectiveAt:  e.Event.OccurredAt,
			Authority:    schematypes.KnowledgeAuthorityObserved,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: repository.DisplayName},
			Subject:      rez.KnowledgeSubjectRef{Entity: &repositoryRef},
		})
	}
	if ingestErr := s.knowledge.IngestEvidence(ctx, e.Event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("deployment involvement: %w", ingestErr)
	}
	if recordErr := s.changePoints.RecordChangePoint(ctx, e.Event, e.Event.OccurredAt); recordErr != nil {
		return nil, fmt.Errorf("record deployment: %w", recordErr)
	}
	return nil, nil
}
