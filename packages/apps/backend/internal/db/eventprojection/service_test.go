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

func (s *ProjectionServiceSuite) projectionService(tdb rez.Database) *ProjectionService {
	users, _ := db.NewUserService(tdb, mocks.NewMockOrganizationService(s.T()))

	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().AddHandlers(mock.Anything).Return(nil).Once()

	incidents, _ := db.NewIncidentService(tdb, messageService)

	knowledge, _ := db.NewKnowledgeGraphService(tdb)

	service, err := NewProjectionService(tdb, users, incidents, knowledge)
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

func (s *ProjectionServiceSuite) createNormalizedEvent(tdb rez.Database, subjectKind projections.SubjectKind, providerSubjectRef string, kind ne.Kind, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)

	create := tdb.Client(ctx).NormalizedEvent.Create().
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
	tdb := s.CreateTestDatabase()
	service := s.projectionService(tdb)
	now := time.Now().UTC()

	for _, component := range []projections.SystemComponentSubjectAttributes{
		{ExternalRef: "api", Kind: kne.KindContainer, Subkind: "service", DisplayName: "API"},
		{ExternalRef: "database", Kind: kne.KindContainer, Subkind: "database", DisplayName: "Database"},
	} {
		event := s.createNormalizedEvent(
			tdb,
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
		Kind:              knr.KindInteractsWith,
		Subkind:           "uses",
		DisplayName:       "uses",
		SourceExternalRef: "api",
		SourceKind:        kne.KindContainer,
		SourceSubkind:     "service",
		SourceDisplayName: "API",
		TargetExternalRef: "database",
		TargetKind:        kne.KindContainer,
		TargetSubkind:     "database",
		TargetDisplayName: "Database",
	}
	event := s.createNormalizedEvent(
		tdb,
		projections.SubjectKindSystemRelationship,
		relationship.ExternalRef,
		ne.KindObserved,
		now,
		relationship,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.KindEQ(kne.KindContainer), kne.SubkindIn("service", "database")).
		CountX(ctx))
	s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.KindEQ(knr.KindInteractsWith), knr.Subkind("uses")).
		CountX(ctx))
}

func (s *ProjectionServiceSuite) TestProjectsTeamMembershipIntoDomainAndGraph() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.projectionService(tdb)
	suffix := uuid.NewString()
	attributes := projections.TeamMembershipSubjectAttributes{
		Team: projections.TeamSubjectAttributes{
			ExternalRef: "slack:group-" + suffix,
			Name:        "Platform",
			Slug:        "platform-" + suffix,
		},
		User: projections.UserSubjectAttributes{
			ExternalRef: "slack:user-" + suffix,
			Name:        "Avery",
			Email:       suffix + "@example.com",
			ChatId:      "user-" + suffix,
		},
		Role: "member",
	}
	event := s.createNormalizedEvent(
		tdb,
		projections.SubjectKindTeamMembership,
		"slack:group-"+suffix+":user-"+suffix,
		ne.KindObserved,
		time.Now().UTC(),
		attributes,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	createdUser := tdb.Client(ctx).User.Query().Where(user.Email(attributes.User.Email)).OnlyX(ctx)
	createdTeam := tdb.Client(ctx).Team.Query().Where(team.Slug(attributes.Team.Slug)).OnlyX(ctx)
	s.NotNil(createdUser.KnowledgeEntityID)
	s.NotNil(createdTeam.KnowledgeEntityID)
	s.Equal(1, tdb.Client(ctx).TeamMembership.Query().
		Where(teammembership.TeamID(createdTeam.ID), teammembership.UserID(createdUser.ID)).
		CountX(ctx))
	s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().
		Where(
			knr.KindEQ(knr.KindParticipatesIn),
			knr.Subkind(knowledgeRelationshipSubkindMemberOf),
			knr.SourceEntityID(*createdUser.KnowledgeEntityID),
			knr.TargetEntityID(*createdTeam.KnowledgeEntityID),
		).
		CountX(ctx))
}
