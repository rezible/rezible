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
		return nil, fmt.Errorf("unsupported event kind %q", event.Kind)
	}
	return projector(ctx, event)
}

func (s *ProjectionServiceSuite) createNormalizedEvent(tdb rez.Database, kind string, providerResourceRef string, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)

	create := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("projection-tests").
		SetProviderResourceRef(providerResourceRef).
		SetProviderEventSource("projection").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetKind(kind).
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

	sourceRef := rez.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "projection-tests",
		ResourceRef:       "api",
	}
	targetRef := rez.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "projection-tests",
		ResourceRef:       "database",
	}
	relationship := projections.SystemRelationshipEventAttributes{
		Predicate:   knr.PredicateUses,
		DisplayName: "uses",
		Source: projections.EntityObservation{
			Ref:         sourceRef,
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: "API",
		},
		Target: projections.EntityObservation{
			Ref:         targetRef,
			Category:    kne.CategoryContainer,
			Kind:        "database",
			DisplayName: "Database",
		},
	}
	event := s.createNormalizedEvent(
		tdb,
		projections.KindSystemRelationship,
		"api-uses-database",
		now,
		relationship,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	entityQuery := tdb.Client(ctx).KnowledgeEntity.Query()
	entityQuery.Where(kne.CategoryEQ(kne.CategoryContainer), kne.KindIn("service", "database"))
	s.Equal(2, entityQuery.CountX(ctx))
	relationshipQuery := tdb.Client(ctx).KnowledgeRelationship.Query()
	relationshipQuery.Where(knr.PredicateEQ(knr.PredicateUses))
	s.Equal(1, relationshipQuery.CountX(ctx))
	s.Equal(3, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}

func (s *ProjectionServiceSuite) TestProjectsTeamMembershipIntoDomainAndGraph() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.projectionService(tdb)
	suffix := uuid.NewString()
	teamRef := rez.ProviderResourceRef{
		Provider:          "slack",
		ProviderNamespace: "workspace",
		ResourceRef:       "group-" + suffix,
	}
	userRef := rez.ProviderResourceRef{
		Provider:          "slack",
		ProviderNamespace: "workspace",
		ResourceRef:       "user-" + suffix,
	}
	attributes := projections.TeamMembershipEventAttributes{
		Team: projections.TeamMembershipTeamAttributes{
			ProviderResourceRef: teamRef,
			Name:                "Platform",
			Slug:                "platform-" + suffix,
		},
		User: projections.TeamMembershipUserAttributes{
			ProviderResourceRef: userRef,
			Name:                "Avery",
			Email:               suffix + "@example.com",
			ChatId:              "user-" + suffix,
		},
		Role: "member",
	}
	event := s.createNormalizedEvent(
		tdb,
		projections.KindTeamMembership,
		"membership:group-"+suffix+":user-"+suffix,
		time.Now().UTC(),
		attributes,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	userQuery := tdb.Client(ctx).User.Query()
	userQuery.Where(user.Email(attributes.User.Email))
	createdUser := userQuery.OnlyX(ctx)
	teamQuery := tdb.Client(ctx).Team.Query()
	teamQuery.Where(team.Slug(attributes.Team.Slug))
	createdTeam := teamQuery.OnlyX(ctx)
	s.NotNil(createdUser.KnowledgeEntityID)
	s.NotNil(createdTeam.KnowledgeEntityID)
	membershipQuery := tdb.Client(ctx).TeamMembership.Query()
	membershipQuery.Where(teammembership.TeamID(createdTeam.ID), teammembership.UserID(createdUser.ID))
	s.Equal(1, membershipQuery.CountX(ctx))
	relationshipQuery := tdb.Client(ctx).KnowledgeRelationship.Query()
	relationshipQuery.Where(
		knr.PredicateEQ(knr.PredicateMemberOf),
		knr.SourceEntityID(*createdUser.KnowledgeEntityID),
		knr.TargetEntityID(*createdTeam.KnowledgeEntityID),
	)
	s.Equal(1, relationshipQuery.CountX(ctx))
	s.Equal(3, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}
