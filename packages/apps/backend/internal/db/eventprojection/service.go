package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

type ProjectionService struct {
	db        rez.Database
	users     rez.UserService
	incidents rez.IncidentService
	knowledge rez.KnowledgeGraphService

	projFns map[projections.SubjectKind]rez.EventProjectorFunc
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
		projFns:   map[projections.SubjectKind]rez.EventProjectorFunc{},
	}
	s.registerProjectorFuncs()
	return s, nil
}

func (s *ProjectionService) registerProjectorFuncs() {
	s.projFns = map[projections.SubjectKind]rez.EventProjectorFunc{
		projections.SubjectKindUser:               makeProjector(projections.DecodeUserEvent, s.handleUserEvent),
		projections.SubjectKindTeam:               makeProjector(projections.DecodeTeamEvent, s.handleTeamEvent),
		projections.SubjectKindTeamMembership:     makeProjector(projections.DecodeTeamMembershipEvent, s.handleTeamMembershipEvent),
		projections.SubjectKindSystemComponent:    makeProjector(projections.DecodeSystemComponentEvent, s.handleSystemComponentEvent),
		projections.SubjectKindSystemRelationship: makeProjector(projections.DecodeSystemRelationshipEvent, s.handleSystemRelationshipEvent),
		projections.SubjectKindCodeForge:          makeProjector(projections.DecodeCodeForgeEvent, s.handleCodeForgeEvent),
		projections.SubjectKindCodeChange:         makeProjector(projections.DecodeCodeChangeEvent, s.handleCodeChangeEvent),
		projections.SubjectKindIncident:           makeProjector(projections.DecodeIncidentEvent, s.handleIncidentEvent),
		projections.SubjectKindAlertInstance:      makeProjector(projections.DecodeAlertInstanceEvent, s.handleAlertInstanceEvent),
	}
}

func (s *ProjectionService) GetEventProjectorFunc(ev *ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	fn, ok := s.projFns[projections.SubjectKind(ev.SubjectKind)]
	return fn, ok
}

func projectionEvidenceKind(event *ent.NormalizedEvent) ke.Kind {
	if event.Kind == ne.KindDeleted {
		return ke.KindDeleted
	}
	return ke.KindObserved
}
