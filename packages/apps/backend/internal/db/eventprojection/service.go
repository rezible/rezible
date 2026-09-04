package eventprojection

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

type ProjectionService struct {
	db        rez.Database
	users     rez.UserService
	incidents rez.IncidentService
	knowledge rez.KnowledgeGraphService

	projFns map[string]rez.EventProjectorFunc
}

func makeProjector[E any](decodeFn func(*ent.NormalizedEvent) (E, error), projFn func(context.Context, E) ([]rez.ProjectedEntityRef, error)) rez.EventProjectorFunc {
	return func(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		proj, decodeErr := decodeFn(ev)
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid event: %w", decodeErr)
		}
		return projFn(ctx, proj)
	}
}

func NewProjectionService(db rez.Database, users rez.UserService, incidents rez.IncidentService, knowledge rez.KnowledgeGraphService) (*ProjectionService, error) {
	s := &ProjectionService{
		db:        db,
		users:     users,
		incidents: incidents,
		knowledge: knowledge,
		projFns:   map[string]rez.EventProjectorFunc{},
	}
	s.registerProjectorFuncs()
	return s, nil
}

func (s *ProjectionService) registerProjectorFuncs() {
	s.projFns = map[string]rez.EventProjectorFunc{
		projections.KindUser:               makeProjector(projections.DecodeUserEvent, s.handleUserEvent),
		projections.KindTeam:               makeProjector(projections.DecodeTeamEvent, s.handleTeamEvent),
		projections.KindTeamMembership:     makeProjector(projections.DecodeTeamMembershipEvent, s.handleTeamMembershipEvent),
		projections.KindSystemComponent:    makeProjector(projections.DecodeSystemComponentEvent, s.handleSystemComponentEvent),
		projections.KindSystemRelationship: makeProjector(projections.DecodeSystemRelationshipEvent, s.handleSystemRelationshipEvent),
		projections.KindCodeForge:          makeProjector(projections.DecodeCodeForgeEvent, s.handleCodeForgeEvent),
		projections.KindCodeChange:         makeProjector(projections.DecodeCodeChangeEvent, s.handleCodeChangeEvent),
		projections.KindIncident:           makeProjector(projections.DecodeIncidentEvent, s.handleIncidentEvent),
		projections.KindAlertInstance:      makeProjector(projections.DecodeAlertInstanceEvent, s.handleAlertInstanceEvent),
	}
}

func (s *ProjectionService) GetEventProjectorFunc(ev *ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	fn, ok := s.projFns[ev.Kind]
	return fn, ok
}

func projectionEvidenceKind(event *ent.NormalizedEvent) ke.Kind {
	if event.Kind == projections.KindTeam {
		var attrs projections.TeamEventAttributes
		if json.Unmarshal(event.Attributes, &attrs) == nil && attrs.Deleted {
			return ke.KindDeleted
		}
	}
	return ke.KindObserved
}

func (s *ProjectionService) ingestSubjectEvidence(ctx context.Context, event *ent.NormalizedEvent, evidence ent.KnowledgeEvidenceRef, supportingEvidence ...ent.KnowledgeEvidenceRef) (*ent.KnowledgeSubjectAlias, error) {
	refs := append(supportingEvidence, evidence)
	if err := s.knowledge.IngestEvidence(ctx, event, refs...); err != nil {
		return nil, err
	}
	var ref rez.ProviderResourceRef
	if evidence.SubjectEntity != nil {
		ref = evidence.SubjectEntity.ProviderResourceRef
	} else if evidence.SubjectRelationship != nil {
		ref = evidence.SubjectRelationship.ProviderResourceRef
	}
	query := s.db.Client(ctx).KnowledgeSubjectAlias.Query()
	query.Where(
		ksa.Provider(ref.Provider),
		ksa.ProviderNamespace(ref.ProviderNamespace),
		ksa.ProviderResourceRef(ref.ResourceRef),
	)
	query.WithEntity()
	query.WithRelationship()
	return query.Only(ctx)
}
