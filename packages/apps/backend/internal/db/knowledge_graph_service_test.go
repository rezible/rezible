package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/db/eventprojection"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type KnowledgeGraphServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphServiceSuite) knowledgeService() *KnowledgeGraphService {
	return &KnowledgeGraphService{db: s.Database()}
}

func (s *KnowledgeGraphServiceSuite) projectionService() *eventprojection.ProjectionService {
	users, userErr := NewUserService(s.Database(), nil)
	s.Require().NoError(userErr)
	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().AddEventHandlers(mock.Anything).Return(nil).Once()
	incidents, incidentErr := NewIncidentService(s.Database(), messageService)
	s.Require().NoError(incidentErr)
	service, serviceErr := eventprojection.NewProjectionService(s.Database(), users, incidents)
	s.Require().NoError(serviceErr)
	return service
}

func (s *KnowledgeGraphServiceSuite) createEvent(kind projections.SubjectKind, subjectRef string, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)
	event, createErr := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("knowledge-graph-tests").
		SetProviderEventRef(uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(kind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encodedAttributes).
		Save(ctx)
	s.Require().NoError(createErr)
	return event
}

func (s *KnowledgeGraphServiceSuite) TestViewReturnsEvidenceAndRecordsCitations() {
	ctx := s.SeedTenantContext()
	projector := s.projectionService()
	event := s.createEvent(projections.SubjectKindAlertInstance, "alert:latency", time.Now().Add(-time.Minute), projections.AlertInstanceSubjectAttributes{
		ExternalRef: "latency", Title: "Latency high", Description: "p95 is high",
		RelatedEntities: []projections.RelatedEntityRef{{ExternalRef: "service:api", Kind: "service", DisplayName: "API"}},
	})
	projectEvent, ok := projector.GetEventProjectorFunc(event.SubjectKind)
	s.Require().True(ok)
	_, projectErr := projectEvent(ctx, event)
	s.Require().NoError(projectErr)

	alert, alertErr := s.Client(ctx).Alert.Query().Only(ctx)
	s.Require().NoError(alertErr)
	s.NotNil(alert.KnowledgeEntityID)
	instance, instanceErr := s.Client(ctx).AlertInstance.Query().Only(ctx)
	s.Require().NoError(instanceErr)
	s.Require().NotNil(instance.KnowledgeEntityID)
	service := s.knowledgeService()
	result, contextErr := service.GetView(ctx, *instance.KnowledgeEntityID, rez.GetKnowledgeGraphViewParams{Depth: 2})
	s.Require().NoError(contextErr)
	s.Len(result.Entities, 3)
	s.Len(result.Relationships, 2)
	s.Len(result.Evidence, 5)
	for _, evidence := range result.Evidence {
		s.NotEmpty(evidence.Assertion)
		s.NotNil(evidence.Edges.Event)
	}

	user, userErr := s.Client(ctx).User.Create().SetName("Investigator").SetEmail(uuid.NewString() + "@example.test").Save(ctx)
	s.Require().NoError(userErr)
	session, sessionErr := s.Client(ctx).AgentSession.Create().SetAgentName("alerts").SetOwnerUserID(user.ID).Save(ctx)
	s.Require().NoError(sessionErr)
	turnID := uuid.New()
	_, turnErr := s.Client(ctx).AgentTurn.Create().
		SetID(turnID).
		SetAgentSessionID(session.ID).
		SetRiverJobID(1).
		SetInput([]byte("{}")).
		SetStatus(agentturn.StatusRunning).
		Save(ctx)
	s.Require().NoError(turnErr)
	citations := make([]rez.KnowledgeCitation, len(result.Evidence))
	for i, evidence := range result.Evidence {
		citations[i] = rez.KnowledgeCitation{EvidenceID: evidence.ID, Summary: "Supports the investigation"}
	}
	s.Require().NoError(service.RecordTurnKnowledgeCitations(ctx, turnID, citations))
	s.Require().NoError(service.RecordTurnKnowledgeCitations(ctx, turnID, citations))
	citationCount, citationErr := s.Client(ctx).AgentTurnKnowledgeCitation.Query().Count(ctx)
	s.Require().NoError(citationErr)
	s.Equal(len(citations), citationCount)
}

func (s *KnowledgeGraphServiceSuite) TestViewHonorsTraversalDepth() {
	ctx := s.SeedTenantContext()
	root, rootErr := s.Client(ctx).KnowledgeEntity.Create().SetKind("service").SetReference("root").SetDisplayName("Root").Save(ctx)
	s.Require().NoError(rootErr)
	middle, middleErr := s.Client(ctx).KnowledgeEntity.Create().SetKind("service").SetReference("middle").SetDisplayName("Middle").Save(ctx)
	s.Require().NoError(middleErr)
	leaf, leafErr := s.Client(ctx).KnowledgeEntity.Create().SetKind("database").SetReference("leaf").SetDisplayName("Leaf").Save(ctx)
	s.Require().NoError(leafErr)
	_, relationshipErr := s.Client(ctx).KnowledgeRelationship.Create().SetKind("calls").SetSourceEntityID(root.ID).SetTargetEntityID(middle.ID).Save(ctx)
	s.Require().NoError(relationshipErr)
	_, relationshipErr = s.Client(ctx).KnowledgeRelationship.Create().SetKind("uses").SetSourceEntityID(middle.ID).SetTargetEntityID(leaf.ID).Save(ctx)
	s.Require().NoError(relationshipErr)

	service := s.knowledgeService()
	one, oneErr := service.GetView(ctx, root.ID, rez.GetKnowledgeGraphViewParams{Depth: 1})
	s.Require().NoError(oneErr)
	s.Len(one.Entities, 2)
	s.Len(one.Relationships, 1)
	two, twoErr := service.GetView(ctx, root.ID, rez.GetKnowledgeGraphViewParams{Depth: 2})
	s.Require().NoError(twoErr)
	s.Len(two.Entities, 3)
	s.Len(two.Relationships, 2)
}

func (s *KnowledgeGraphServiceSuite) TestHistoricalReferencesReturnStateAtReferencedTime() {
	ctx := s.SeedTenantContext()
	observedAt := time.Now().Add(-2 * time.Hour)
	changedAt := observedAt.Add(time.Hour)
	entity, entityErr := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("service").
		SetReference("checkout").
		SetDisplayName("Checkout API v2").
		Save(ctx)
	s.Require().NoError(entityErr)
	alias, aliasErr := s.Client(ctx).KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderSubjectRef("checkout").
		SetDescription("checkout service").
		SetEntityID(entity.ID).
		Save(ctx)
	s.Require().NoError(aliasErr)
	oldEvent := s.createEvent(projections.SubjectKindSystemComponent, "checkout", observedAt, map[string]any{})
	_, evidenceErr := s.Client(ctx).KnowledgeEvidence.Create().
		SetEventID(oldEvent.ID).
		SetAliasID(alias.ID).
		SetAssertion("service_observed").
		SetEvidenceKind(ke.EvidenceKindObserved).
		SetEffectiveAt(observedAt).
		SetProperties(map[string]any{}).
		SetSubjectState(map[string]any{
			"subject_kind": "entity",
			"entity": ent.KnowledgeEntityRef{
				Kind: "service", Reference: "checkout", DisplayName: "Checkout API v1",
			},
		}).
		Save(ctx)
	s.Require().NoError(evidenceErr)
	newEvent := s.createEvent(projections.SubjectKindSystemComponent, "checkout", changedAt, map[string]any{})
	_, evidenceErr = s.Client(ctx).KnowledgeEvidence.Create().
		SetEventID(newEvent.ID).
		SetAliasID(alias.ID).
		SetAssertion("service_changed").
		SetEvidenceKind(ke.EvidenceKindChanged).
		SetEffectiveAt(changedAt).
		SetProperties(map[string]any{}).
		SetSubjectState(map[string]any{
			"subject_kind": "entity",
			"entity": ent.KnowledgeEntityRef{
				Kind: "service", Reference: "checkout", DisplayName: "Checkout API v2",
			},
		}).
		Save(ctx)
	s.Require().NoError(evidenceErr)

	historical, historicalErr := s.knowledgeService().GetEntityAt(ctx, entity.ID, observedAt.Add(30*time.Minute))
	s.Require().NoError(historicalErr)
	s.Equal("Checkout API v1", historical.DisplayName)
	s.Equal("Checkout API v2", entity.DisplayName)

	database, databaseErr := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("database").
		SetReference("orders").
		SetDisplayName("Orders").
		Save(ctx)
	s.Require().NoError(databaseErr)
	relationship, relationshipErr := s.Client(ctx).KnowledgeRelationship.Create().
		SetKind("reads_from").
		SetSourceEntityID(entity.ID).
		SetTargetEntityID(database.ID).
		SetDescription("Reads and writes orders").
		Save(ctx)
	s.Require().NoError(relationshipErr)
	relationshipAlias, relationshipAliasErr := s.Client(ctx).KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindRelationship).
		SetProvider("test").
		SetProviderSubjectRef("checkout-orders").
		SetDescription("checkout to orders").
		SetRelationshipID(relationship.ID).
		Save(ctx)
	s.Require().NoError(relationshipAliasErr)
	relationshipEvent := s.createEvent(projections.SubjectKindSystemRelationship, "checkout-orders", observedAt, map[string]any{})
	_, relationshipEvidenceErr := s.Client(ctx).KnowledgeEvidence.Create().
		SetEventID(relationshipEvent.ID).
		SetAliasID(relationshipAlias.ID).
		SetAssertion("relationship_observed").
		SetEvidenceKind(ke.EvidenceKindObserved).
		SetEffectiveAt(observedAt).
		SetProperties(map[string]any{}).
		SetSubjectState(map[string]any{
			"subject_kind": "relationship",
			"relationship": ent.KnowledgeRelationshipRef{
				Kind: "reads_from", Description: "Reads orders",
				EntityRefs: [2]ent.KnowledgeEntityRef{
					{Kind: "service", Reference: "checkout"},
					{Kind: "database", Reference: "orders"},
				},
			},
		}).
		Save(ctx)
	s.Require().NoError(relationshipEvidenceErr)

	historicalRelationship, historicalRelationshipErr := s.knowledgeService().GetRelationshipAt(ctx, relationship.ID, observedAt.Add(30*time.Minute))
	s.Require().NoError(historicalRelationshipErr)
	s.Equal("Reads orders", historicalRelationship.Description)
	s.Equal("Reads and writes orders", relationship.Description)
}

func (s *KnowledgeGraphServiceSuite) TestViewReportsEntityLimit() {
	ctx := s.SeedTenantContext()
	root, rootErr := s.Client(ctx).KnowledgeEntity.Create().SetKind("service").SetReference("limit-root").SetDisplayName("Root").Save(ctx)
	s.Require().NoError(rootErr)
	for i := 0; i < maxKnowledgeViewEntities; i++ {
		target, targetErr := s.Client(ctx).KnowledgeEntity.Create().
			SetKind("service").
			SetReference(uuid.NewString()).
			Save(ctx)
		s.Require().NoError(targetErr)
		_, relationshipErr := s.Client(ctx).KnowledgeRelationship.Create().
			SetKind("calls").
			SetSourceEntityID(root.ID).
			SetTargetEntityID(target.ID).
			Save(ctx)
		s.Require().NoError(relationshipErr)
	}

	view, viewErr := s.knowledgeService().GetView(ctx, root.ID, rez.GetKnowledgeGraphViewParams{Depth: 1})
	s.Require().NoError(viewErr)
	s.True(view.Truncated)
	s.Len(view.Entities, maxKnowledgeViewEntities)
	s.Len(view.Warnings, 1)
}

func (s *KnowledgeGraphServiceSuite) TestViewDoesNotCrossTenantBoundary() {
	otherTenantContext := s.SystemContext()
	otherTenant, tenantErr := s.Client(otherTenantContext).Tenant.Create().Save(otherTenantContext)
	s.Require().NoError(tenantErr)
	otherTenantContext = execution.NewTenantContext(otherTenantContext, otherTenant.ID)
	otherEntity, entityErr := s.Client(otherTenantContext).KnowledgeEntity.Create().
		SetKind("service").
		SetReference("other-tenant-service").
		SetDisplayName("Other tenant service").
		Save(otherTenantContext)
	s.Require().NoError(entityErr)

	tenantID := s.SeedTenant.ID
	userID := uuid.New()
	ctx := execution.SetContext(s.T().Context(), execution.Context{
		ActorKind: execution.KindUser,
		Auth:      execution.Auth{TenantID: &tenantID, UserID: &userID},
	})
	_, viewErr := s.knowledgeService().GetView(ctx, otherEntity.ID, rez.GetKnowledgeGraphViewParams{Depth: 1})
	s.Require().Error(viewErr)
}
