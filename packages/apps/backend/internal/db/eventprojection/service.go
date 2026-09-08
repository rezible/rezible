package eventprojection

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/rezible/rezible/pkg/projections"
)

type ProjectionService struct {
	db        rez.Database
	users     rez.UserService
	incidents rez.IncidentService
	knowledge rez.KnowledgeGraphService
	alerts    rez.AlertService

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

func NewProjectionService(db rez.Database, users rez.UserService, incidents rez.IncidentService, knowledge rez.KnowledgeGraphService, alerts rez.AlertService) (*ProjectionService, error) {
	s := &ProjectionService{
		db:        db,
		users:     users,
		incidents: incidents,
		knowledge: knowledge,
		alerts:    alerts,
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
