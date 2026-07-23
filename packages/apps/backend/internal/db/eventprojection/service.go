package eventprojection

import (
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
}

func (s *ProjectionService) GetEventProjectorFunc(subjectKind string) (rez.EventProjectorFunc, bool) {
	switch projections.SubjectKind(subjectKind) {
	case projections.SubjectKindCodeForge:
		return s.handleCodeForgeEventProjection, true
	case projections.SubjectKindCodeChange:
		return s.handleCodeChangeEventProjection, true
	case projections.SubjectKindSystemComponent:
		return s.handleSystemComponentEventProjection, true
	case projections.SubjectKindSystemRelationship:
		return s.handleSystemRelationshipEventProjection, true
	case projections.SubjectKindUser:
		return s.handleUserEventProjection, true
	case projections.SubjectKindIncident:
		return s.handleIncidentEventProjection, true
	case projections.SubjectKindIncidentImpact:
		return s.handleIncidentImpactEventProjection, true
	case projections.SubjectKindAlertInstance:
		return s.handleAlertEventProjection, true
	default:
		return nil, false
	}
}

func projectionEvidenceKind(event *ent.NormalizedEvent) ke.EvidenceKind {
	if event.Kind == ne.KindDeleted {
		return ke.EvidenceKindDeleted
	}
	return ke.EvidenceKindObserved
}

func NewProjectionService(db rez.Database, users rez.UserService, incidents rez.IncidentService) (*ProjectionService, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	if users == nil {
		return nil, fmt.Errorf("user service is required")
	}
	if incidents == nil {
		return nil, fmt.Errorf("incident service is required")
	}
	return &ProjectionService{db: db, users: users, incidents: incidents}, nil
}
