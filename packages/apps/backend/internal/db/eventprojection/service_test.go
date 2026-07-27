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
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	"github.com/rezible/rezible/ent/user"
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

func (s *ProjectionServiceSuite) TestProjectsSystemTopologyRelationship() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()
	now := time.Now().UTC()

	for _, component := range []projections.SystemComponentSubjectAttributes{
		{ExternalRef: "api", Kind: "service", DisplayName: "API"},
		{ExternalRef: "database", Kind: "database", DisplayName: "Database"},
	} {
		event := s.createNormalizedEvent(
			projections.SubjectKindSystemComponent,
			component.ExternalRef,
			ne.KindObserved,
			now,
			component,
		)
		_, projectionErr := runProjection(ctx, service, event)
		s.Require().NoError(projectionErr)
	}

	relationship := projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       "api-uses-database",
		Kind:              "uses",
		DisplayName:       "uses",
		SourceExternalRef: "api",
		SourceKind:        "service",
		SourceDisplayName: "API",
		TargetExternalRef: "database",
		TargetKind:        "database",
		TargetDisplayName: "Database",
	}
	event := s.createNormalizedEvent(
		projections.SubjectKindSystemRelationship,
		relationship.ExternalRef,
		ne.KindObserved,
		now,
		relationship,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	s.Equal(2, s.Client(ctx).KnowledgeEntity.Query().Where(kne.Kind(knowledgeEntityKindSystemComponent)).CountX(ctx))
	s.Equal(1, s.Client(ctx).KnowledgeRelationship.Query().Where(knr.Kind("uses")).CountX(ctx))
}

func (s *ProjectionServiceSuite) TestProjectsTeamMembershipIntoDomainAndGraph() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()
	suffix := uuid.NewString()
	attributes := projections.TeamMembershipSubjectAttributes{
		Team: projections.TeamSubjectAttributes{
			ExternalRef: "slack:group-" + suffix,
			Name:        "Platform",
			Slug:        "platform-" + suffix,
		},
		User: projections.UserSubjectAttributes{
			Name:   "Avery",
			Email:  suffix + "@example.com",
			ChatId: "user-" + suffix,
		},
		UserExternalRef: "slack:user-" + suffix,
		Role:            "member",
	}
	event := s.createNormalizedEvent(
		projections.SubjectKindTeamMembership,
		"slack:group-"+suffix+":user-"+suffix,
		ne.KindObserved,
		time.Now().UTC(),
		attributes,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	createdUser := s.Client(ctx).User.Query().Where(user.Email(attributes.User.Email)).OnlyX(ctx)
	createdTeam := s.Client(ctx).Team.Query().Where(team.Slug(attributes.Team.Slug)).OnlyX(ctx)
	s.NotNil(createdUser.KnowledgeEntityID)
	s.NotNil(createdTeam.KnowledgeEntityID)
	s.Equal(1, s.Client(ctx).TeamMembership.Query().
		Where(teammembership.TeamID(createdTeam.ID), teammembership.UserID(createdUser.ID)).
		CountX(ctx))
	s.Equal(1, s.Client(ctx).KnowledgeRelationship.Query().
		Where(
			knr.Kind(knowledgeRelationshipKindMemberOf),
			knr.SourceEntityID(*createdUser.KnowledgeEntityID),
			knr.TargetEntityID(*createdTeam.KnowledgeEntityID),
		).
		CountX(ctx))
}
