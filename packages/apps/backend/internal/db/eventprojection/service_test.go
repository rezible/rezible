package eventprojection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/internal/db"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type ProjectionServiceSuite struct {
	test.Suite
}

func TestProjectionServiceSuite(t *testing.T) {
	suite.Run(t, &ProjectionServiceSuite{Suite: test.NewSuite()})
}

func (s *ProjectionServiceSuite) projectionService() *ProjectionService {
	users, _ := db.NewUserService(s.Database(), mocks.NewMockOrganizationService(s.T()))

	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().AddEventHandlers(mock.Anything).Return(nil).Once()

	incidents, _ := db.NewIncidentService(s.Database(), messageService)

	knowledge, _ := db.NewKnowledgeGraphService(s.Database())

	service, err := NewProjectionService(s.Database(), users, incidents, knowledge)
	s.Require().NoError(err)
	return service
}

func runProjection(ctx context.Context, service rez.EventProjectionService, event *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	projector, ok := service.GetEventProjectorFunc(event)
	if !ok {
		return nil, fmt.Errorf("unsupported subject kind %q", event.SubjectKind)
	}
	return projector(ctx, event)
}

func (s *ProjectionServiceSuite) createNormalizedEvent(subjectKind projections.SubjectKind, providerSubjectRef string, kind ne.Kind, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)

	create := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("projection").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetProviderSubjectRef(providerSubjectRef).
		SetKind(kind).
		SetSubjectKind(subjectKind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt.Add(time.Minute)).
		SetAttributes(encodedAttributes)

	event, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return event
}
