package eventprojection

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/internal/db"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
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
	suite.Run(t, &ProjectionServiceSuite{
		Suite: test.NewSuite(),
	})
}

func (s *ProjectionServiceSuite) projectionService(tdb rez.Database) *ProjectionService {
	users, usersErr := db.NewUserService(tdb, mocks.NewMockOrganizationService(s.T()))
	s.Require().NoError(usersErr)

	messageService := mocks.NewMockMessageQueue(s.T())

	retrospectives, retrospectivesErr := db.NewRetrospectiveService(tdb)
	s.Require().NoError(retrospectivesErr)

	incidents, incidentsErr := db.NewIncidentService(tdb, messageService, nil, retrospectives)
	s.Require().NoError(incidentsErr)

	knowledge, knowledgeErr := db.NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)

	knowledgeQuery, knowledgeQueryErr := db.NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(knowledgeQueryErr)

	analysisService, analysisServiceErr := db.NewSystemAnalysisService(tdb, knowledgeQuery)
	s.Require().NoError(analysisServiceErr)

	jobService := mocks.NewMockJobService(s.T())
	agentService := mocks.NewMockAiAgentSessionService(s.T())
	agentService.EXPECT().
		CreateAgentSession(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, params rez.CreateAiAgentSessionParams) (*ent.AgentSession, error) {
			return s.persistProjectionAgentSession(ctx, tdb, params)
		}).
		Maybe()
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 1,
			},
		}, nil).
		Maybe()
	investigationService := db.NewInvestigationService(tdb, agentService, jobService)
	situations, situationsErr := db.NewSituationService(tdb, investigationService, analysisService, knowledgeQuery)
	s.Require().NoError(situationsErr)

	alerts, alertsErr := db.NewAlertService(tdb, situations, knowledge)
	s.Require().NoError(alertsErr)

	service, serviceErr := NewProjectionService(tdb, knowledge, knowledgeQuery, users, incidents, alerts)
	s.Require().NoError(serviceErr)

	return service
}

func (s *ProjectionServiceSuite) persistProjectionAgentSession(ctx context.Context, tdb rez.Database, params rez.CreateAiAgentSessionParams) (*ent.AgentSession, error) {
	input, marshalErr := json.Marshal(params.Input)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal projection agent input: %w", marshalErr)
	}
	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	createSession := tdb.Client(ctx).AgentSession.Create().
		SetAgentName(params.AgentName).
		SetInput(input).
		SetScopes(params.PermissionScopes).
		SetMetadata(metadata)
	return createSession.Save(ctx)
}

func runProjection(ctx context.Context, service rez.EventProjectionService, event *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	projector, ok := service.GetEventProjectorFunc(event)
	if !ok {
		return nil, fmt.Errorf("unsupported event kind %q", event.Kind)
	}
	return projector(ctx, event)
}

func (s *ProjectionServiceSuite) createNormalizedEvent(ctx context.Context, tdb rez.Database, kind string, providerResourceRef string, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
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
	ctx, tdb := s.SetupTestDatabase()
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
		ctx,
		tdb,
		projections.KindSystemRelationship,
		"api-uses-database",
		now,
		relationship,
	)
	_, projectionErr := runProjection(ctx, service, event)
	s.Require().NoError(projectionErr)

	client := tdb.Client(ctx)
	sourceAliasQuery := client.KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider(sourceRef.Provider),
			ksa.ProviderNamespace(sourceRef.ProviderNamespace),
			ksa.ProviderResourceRef(sourceRef.ResourceRef),
		).
		WithEntity()
	sourceAlias, sourceAliasErr := sourceAliasQuery.Only(ctx)
	s.Require().NoError(sourceAliasErr)

	targetAliasQuery := client.KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider(targetRef.Provider),
			ksa.ProviderNamespace(targetRef.ProviderNamespace),
			ksa.ProviderResourceRef(targetRef.ResourceRef),
		).
		WithEntity()
	targetAlias, targetAliasErr := targetAliasQuery.Only(ctx)
	s.Require().NoError(targetAliasErr)
	s.Require().NotNil(sourceAlias.Edges.Entity)
	s.Require().NotNil(targetAlias.Edges.Entity)
	s.NotEqual(sourceAlias.Edges.Entity.ID, targetAlias.Edges.Entity.ID)
	projectedRelationship, relationshipErr := client.KnowledgeRelationship.Query().Only(ctx)
	s.Require().NoError(relationshipErr)
	s.Equal(knr.PredicateUses, projectedRelationship.Predicate)
	s.Equal(sourceAlias.Edges.Entity.ID, projectedRelationship.SourceEntityID, "API is the source of uses")
	s.Equal(targetAlias.Edges.Entity.ID, projectedRelationship.TargetEntityID, "database is the target of uses")

	evidenceQuery := client.KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID)).
		WithSubjectAlias()
	evidence, evidenceErr := evidenceQuery.All(ctx)
	s.Require().NoError(evidenceErr)
	s.Require().Len(evidence, 3)
	var entityAliasIDs []uuid.UUID
	var relationshipAliasIDs []uuid.UUID
	for _, item := range evidence {
		s.Equal(event.ID, item.EventID)
		s.Require().NotNil(item.Edges.SubjectAlias)
		alias := item.Edges.SubjectAlias
		if alias.RelationshipID != nil {
			s.Equal(projectedRelationship.ID, *alias.RelationshipID)
			s.Equal(knowledgeAssertionSystemRelationshipExists, item.Assertion)
			relationshipAliasIDs = append(relationshipAliasIDs, alias.ID)
		} else {
			entityAliasIDs = append(entityAliasIDs, alias.ID)
		}
	}
	s.ElementsMatch([]uuid.UUID{sourceAlias.ID, targetAlias.ID}, entityAliasIDs)
	s.Len(relationshipAliasIDs, 1)

	entityQuery := tdb.Client(ctx).KnowledgeEntity.Query()
	entityQuery.Where(kne.CategoryEQ(kne.CategoryContainer), kne.KindIn("service", "database"))
	entityCount, entityCountErr := entityQuery.Count(ctx)
	s.Require().NoError(entityCountErr)

	s.Equal(2, entityCount)
	relationshipQuery := tdb.Client(ctx).KnowledgeRelationship.Query()
	relationshipQuery.Where(knr.PredicateEQ(knr.PredicateUses))
	relationshipCount, relationshipCountErr := relationshipQuery.Count(ctx)
	s.Require().NoError(relationshipCountErr)

	s.Equal(1, relationshipCount)
	knowledgeEvidenceCount, knowledgeEvidenceCountErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(knowledgeEvidenceCountErr)

	s.Equal(3, knowledgeEvidenceCount)
}

func (s *ProjectionServiceSuite) TestProjectsTeamMembershipIntoDomainAndGraph() {
	ctx, tdb := s.SetupTestDatabase()
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
		ctx,
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
	createdUser, createdUserErr := userQuery.Only(ctx)
	s.Require().NoError(createdUserErr)

	teamQuery := tdb.Client(ctx).Team.Query()
	teamQuery.Where(team.Slug(attributes.Team.Slug))
	createdTeam, createdTeamErr := teamQuery.Only(ctx)
	s.Require().NoError(createdTeamErr)
	s.NotNil(createdUser.KnowledgeEntityID)
	s.NotNil(createdTeam.KnowledgeEntityID)
	membershipQuery := tdb.Client(ctx).TeamMembership.Query()
	membershipQuery.Where(teammembership.TeamID(createdTeam.ID), teammembership.UserID(createdUser.ID))
	membershipCount, membershipCountErr := membershipQuery.Count(ctx)
	s.Require().NoError(membershipCountErr)

	s.Equal(1, membershipCount)
	relationshipQuery := tdb.Client(ctx).KnowledgeRelationship.Query()
	relationshipQuery.Where(
		knr.PredicateEQ(knr.PredicateMemberOf),
		knr.SourceEntityID(*createdUser.KnowledgeEntityID),
		knr.TargetEntityID(*createdTeam.KnowledgeEntityID),
	)
	relationshipCount, relationshipCountErr := relationshipQuery.Count(ctx)
	s.Require().NoError(relationshipCountErr)

	s.Equal(1, relationshipCount)
	knowledgeEvidenceCount, knowledgeEvidenceCountErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(knowledgeEvidenceCountErr)

	s.Equal(3, knowledgeEvidenceCount)
}
