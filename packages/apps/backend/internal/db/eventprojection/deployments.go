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
	knowledgeAssertionDeploymentObserved           = "deployment_observed"
	knowledgeAssertionDeploymentServiceObserved    = "deployment_service_observed"
	knowledgeAssertionDeploymentImpactsService     = "deployment_impacts_service"
	knowledgeAssertionDeploymentRepositoryObserved = "deployment_repository_observed"
	knowledgeAssertionDeploymentTouchesRepository  = "deployment_touches_repository"
)

// handleDeploymentEvent records a report as graph evidence. The report's evidence is effective at the
// deployment's time and carries the whole deployment state, so the deployment entity's state is its latest
// report by deployment time, whatever order reports are processed in.
func (s *ProjectionService) handleDeploymentEvent(ctx context.Context, e *projections.DeploymentEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes
	evidenceKind := projectionEvidenceKind(event)

	deploymentResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	deploymentEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryEvent,
		Kind:                knowledgeEntityKindDeployment,
		ProviderResourceRef: deploymentResourceRef,
	}
	deploymentEvidence := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionDeploymentObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Service.DisplayName + " to " + attrs.Environment.DisplayName,
			Properties:  attrs.StateProperties(),
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &deploymentEntityRef},
	}

	serviceEntityRef := attrs.Service.KnowledgeEntityRef()
	serviceEvidence := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionDeploymentServiceObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Service.DisplayName,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &serviceEntityRef},
	}
	impactsRelationship := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateImpacts,
		ProviderResourceRef: projections.DerivedRelationshipRef(knr.PredicateImpacts, deploymentResourceRef, attrs.Service.Ref),
		Source:              deploymentEntityRef,
		Target:              serviceEntityRef,
	}
	impactsEvidence := rez.KnowledgeEvidenceRef{
		Kind:        evidenceKind,
		Assertion:   knowledgeAssertionDeploymentImpactsService,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Deploys " + attrs.Service.DisplayName,
		},
		Subject: rez.KnowledgeSubjectRef{Relationship: &impactsRelationship},
	}
	evidence := []rez.KnowledgeEvidenceRef{deploymentEvidence, serviceEvidence, impactsEvidence}

	if attrs.Repository != nil {
		repositoryEntityRef := attrs.Repository.KnowledgeEntityRef()
		repositoryEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionDeploymentRepositoryObserved,
			EffectiveAt: event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: attrs.Repository.DisplayName,
			},
			Subject: rez.KnowledgeSubjectRef{Entity: &repositoryEntityRef},
		}
		touchesRelationship := rez.KnowledgeRelationshipRef{
			Predicate:           knr.PredicateTouches,
			ProviderResourceRef: projections.DerivedRelationshipRef(knr.PredicateTouches, deploymentResourceRef, attrs.Repository.Ref),
			Source:              deploymentEntityRef,
			Target:              repositoryEntityRef,
		}
		touchesEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionDeploymentTouchesRepository,
			EffectiveAt: event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: "Deploys " + attrs.Repository.DisplayName,
			},
			Subject: rez.KnowledgeSubjectRef{Relationship: &touchesRelationship},
		}
		evidence = append(evidence, repositoryEvidence, touchesEvidence)
	}

	if ingestErr := s.knowledge.IngestEvidence(ctx, event, evidence...); ingestErr != nil {
		return nil, fmt.Errorf("deployment knowledge evidence: %w", ingestErr)
	}
	return nil, nil
}
