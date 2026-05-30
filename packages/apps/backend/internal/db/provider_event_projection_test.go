package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/incident"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/rezible/rezible/testkit"
	"github.com/stretchr/testify/suite"
)

type ProjectedEntityLinkingSuite struct {
	testkit.Suite
}

func TestProjectedEntityLinkingSuite(t *testing.T) {
	suite.Run(t, &ProjectedEntityLinkingSuite{Suite: testkit.NewSuite()})
}

func (s *ProjectedEntityLinkingSuite) TestIncidentProjectionReusesKnowledgeEntityAndIncidentAcrossEvents() {
	ctx := s.SeedTenantContext()
	subjectRef := "fake:incident:" + uuid.NewString()

	ev1 := s.createIncidentEvent(ctx, subjectRef, "Initial API outage", "Requests are failing")
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev1))

	ev2 := s.createIncidentEvent(ctx, subjectRef, "Updated API outage", "Requests are still failing")
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev2))

	entity := s.entityForAlias(ctx, "fake", subjectRef)

	incidents, err := s.Client().Incident.Query().
		Where(incident.KnowledgeEntityID(entity.ID)).
		All(ctx)
	s.Require().NoError(err)
	s.Len(incidents, 1)
	s.Require().NotNil(incidents[0].KnowledgeEntityID)
	s.Equal(entity.ID, *incidents[0].KnowledgeEntityID)
	s.Equal("Updated API outage", incidents[0].Title)

	evidence := s.entityEvidence(ctx, entity.ID)
	s.Len(evidence, 2)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
	s.Equal(knev.EvidenceKindChanged, evidence[1].EvidenceKind)
}

func (s *ProjectedEntityLinkingSuite) TestIncidentProjectionReprocessingEventDoesNotDuplicateEvidence() {
	ctx := s.SeedTenantContext()
	ev := s.createIncidentEvent(ctx, "fake:incident:"+uuid.NewString(), "API outage", "Requests are failing")

	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev))
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev))

	entity := s.entityForAlias(ctx, "fake", ev.ProviderSubjectRef)
	incidents, err := s.Client().Incident.Query().
		Where(incident.KnowledgeEntityID(entity.ID)).
		All(ctx)
	s.Require().NoError(err)
	s.Len(incidents, 1)
	s.Require().NotNil(incidents[0].KnowledgeEntityID)

	evidence := s.entityEvidence(ctx, *incidents[0].KnowledgeEntityID)
	s.Len(evidence, 1)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
}

func (s *ProjectedEntityLinkingSuite) TestIncidentProjectionUnchangedRefreshWritesObservedEvidence() {
	ctx := s.SeedTenantContext()
	subjectRef := "fake:incident:" + uuid.NewString()

	ev1 := s.createIncidentEvent(ctx, subjectRef, "API outage", "Requests are failing")
	ev2 := s.createIncidentEvent(ctx, subjectRef, "API outage", "Requests are failing")
	ev2.OccurredAt = ev1.OccurredAt
	ev2.ReceivedAt = ev1.ReceivedAt

	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev1))
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev2))

	entity := s.entityForAlias(ctx, "fake", subjectRef)
	incidents, err := s.Client().Incident.Query().
		Where(incident.KnowledgeEntityID(entity.ID)).
		All(ctx)
	s.Require().NoError(err)
	s.Len(incidents, 1)
	s.Require().NotNil(incidents[0].KnowledgeEntityID)

	evidence := s.entityEvidence(ctx, *incidents[0].KnowledgeEntityID)
	s.Len(evidence, 2)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
	s.Equal(knev.EvidenceKindObserved, evidence[1].EvidenceKind)
}

func (s *ProjectedEntityLinkingSuite) TestIncidentProjectionPreservesFirstObservedAndOpenedAt() {
	ctx := s.SeedTenantContext()
	subjectRef := "fake:incident:" + uuid.NewString()

	ev1 := s.createIncidentEvent(ctx, subjectRef, "API outage", "Requests are failing")
	ev1.OccurredAt = time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	ev1.ReceivedAt = ev1.OccurredAt
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev1))

	ev2 := s.createIncidentEvent(ctx, subjectRef, "API outage updated", "Requests are still failing")
	ev2.OccurredAt = time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)
	ev2.ReceivedAt = ev2.OccurredAt
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev2))

	entity := s.entityForAlias(ctx, "fake", subjectRef)
	s.Require().NotNil(entity.FirstObservedAt)
	s.Require().NotNil(entity.LastObservedAt)
	s.True(ev1.OccurredAt.Equal(*entity.FirstObservedAt))
	s.True(ev2.OccurredAt.Equal(*entity.LastObservedAt))

	inc, err := s.Client().Incident.Query().
		Where(incident.KnowledgeEntityID(entity.ID)).
		Only(ctx)
	s.Require().NoError(err)
	s.True(ev1.OccurredAt.Equal(inc.OpenedAt))
}

func (s *ProjectedEntityLinkingSuite) TestIncidentEvidencePropertiesUseEventAttributes() {
	ctx := s.SeedTenantContext()
	subjectRef := "fake:incident:" + uuid.NewString()

	ev1 := s.createIncidentEvent(ctx, subjectRef, "Initial API outage", "First summary")
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev1))

	ev2 := s.createIncidentEvent(ctx, subjectRef, "Updated API outage", "Second summary")
	s.Require().NoError(HandleIncidentEventProjection(ctx, s.Client(), ev2))

	entity := s.entityForAlias(ctx, "fake", subjectRef)
	evidence := s.entityEvidence(ctx, entity.ID)
	s.Require().Len(evidence, 2)
	s.Equal("First summary", evidence[0].Properties["summary"])
	s.Equal("Second summary", evidence[1].Properties["summary"])
}

func (s *ProjectedEntityLinkingSuite) TestAlertProjectionReusesKnowledgeEntityAndAlert() {
	ctx := s.SeedTenantContext()
	subjectRef := "fake:alert:" + uuid.NewString()

	ev1 := s.createAlertEvent(ctx, subjectRef, "Latency alert", "p95 latency is high")
	s.Require().NoError(HandleAlertEventProjection(ctx, s.Client(), ev1))

	ev2 := s.createAlertEvent(ctx, subjectRef, "Latency alert updated", "p99 latency is high")
	s.Require().NoError(HandleAlertEventProjection(ctx, s.Client(), ev2))

	entity := s.entityForAlias(ctx, "fake", subjectRef)
	alerts, err := s.Client().Alert.Query().
		Where(alert.KnowledgeEntityID(entity.ID)).
		All(ctx)
	s.Require().NoError(err)
	s.Len(alerts, 1)
	s.Require().NotNil(alerts[0].KnowledgeEntityID)
	s.Equal("Latency alert updated", alerts[0].Title)
	s.Equal(knowledgeKindAlert, entity.Kind)

	evidence := s.entityEvidence(ctx, *alerts[0].KnowledgeEntityID)
	s.Len(evidence, 2)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
	s.Equal(knev.EvidenceKindChanged, evidence[1].EvidenceKind)
}

func (s *ProjectedEntityLinkingSuite) TestUserProjectionLinksKnowledgeEntityAndFallsBackByEmail() {
	ctx := s.SeedTenantContext()
	subjectRef := "slack:U" + uuid.NewString()
	email := "user-" + uuid.NewString() + "@example.com"

	existingUser, createErr := s.Client().User.Create().
		SetEmail(email).
		SetName("Existing User").
		Save(ctx)
	s.Require().NoError(createErr)

	ev := s.createUserEvent(ctx, subjectRef, email, "Updated User")
	s.Require().NoError(HandleUserEventProjection(ctx, s.Client(), ev))

	entity := s.entityForAlias(ctx, "fake", subjectRef)
	updatedUser, userErr := s.Client().User.Get(ctx, existingUser.ID)
	s.Require().NoError(userErr)
	s.Require().NotNil(updatedUser.KnowledgeEntityID)
	s.Equal(entity.ID, *updatedUser.KnowledgeEntityID)
	s.Equal("Updated User", updatedUser.Name)

	evidence := s.entityEvidence(ctx, entity.ID)
	s.Len(evidence, 1)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
}

func (s *ProjectedEntityLinkingSuite) TestUserProjectionFailsWhenLinkedUserConflictsWithEmailUser() {
	ctx := s.SeedTenantContext()
	subjectRef := "slack:U" + uuid.NewString()

	linkedEmail := "linked-" + uuid.NewString() + "@example.com"
	firstEvent := s.createUserEvent(ctx, subjectRef, linkedEmail, "Linked User")
	s.Require().NoError(HandleUserEventProjection(ctx, s.Client(), firstEvent))

	conflictEmail := "conflict-" + uuid.NewString() + "@example.com"
	_, createErr := s.Client().User.Create().
		SetEmail(conflictEmail).
		SetName("Conflict User").
		Save(ctx)
	s.Require().NoError(createErr)

	conflictEvent := s.createUserEvent(ctx, subjectRef, conflictEmail, "Conflict User")
	s.Error(HandleUserEventProjection(ctx, s.Client(), conflictEvent))
}

func (s *ProjectedEntityLinkingSuite) TestProjectionStatusOnlyCreatedForApplicableHandlers() {
	ctx := s.SeedTenantContext()

	providerEvents := &ProviderEventService{
		db:  s.Client(),
		reg: projections.NewEventProjectionHandlerRegistry(),
	}
	RegisterEventProcessors(providerEvents.reg)

	ev := s.createIncidentEvent(ctx, "fake:incident:"+uuid.NewString(), "API outage", "Requests are failing")

	res := providerEvents.projectNormalizedEvent(ctx, ev)
	s.Empty(res.handlerErrors)

	statuses, err := s.Client().NormalizedEventProjectionStatus.Query().All(ctx)
	s.Require().NoError(err)
	s.Len(statuses, 1)
	s.Equal("incidents", statuses[0].HandlerName)
}

func (s *ProjectedEntityLinkingSuite) TestKnowledgeAliasConflictFailsProjection() {
	ctx := s.SeedTenantContext()

	entityA := s.createKnowledgeEntity(ctx, "service", "Service A")
	entityB := s.createKnowledgeEntity(ctx, "service", "Service B")
	aliasA := EntityAliasRef{Provider: "fake", ProviderSubjectRef: "fake:service:" + uuid.NewString()}
	aliasB := EntityAliasRef{Provider: "fake", ProviderSubjectRef: "fake:service:" + uuid.NewString()}
	s.createKnowledgeAlias(ctx, entityA.ID, aliasA)
	s.createKnowledgeAlias(ctx, entityB.ID, aliasB)

	ev := s.createSystemComponentEvent(ctx, "fake:service:"+uuid.NewString())
	projector := newKnowledgeEntityEventProjector(ev, s.Client())
	_, err := projector.saveProjectedEntity(ctx, ProjectedKnowledgeEntity{
		Kind:        "service",
		Assertion:   assertionSystemComponentExists,
		DisplayName: "Conflicting service",
		Attributes:  map[string]any{"external_ref": "conflict"},
		Aliases:     []EntityAliasRef{aliasA, aliasB},
	})
	s.Error(err)
}

func (s *ProjectedEntityLinkingSuite) createIncidentEvent(ctx context.Context, subjectRef string, title string, summary string) *ent.NormalizedEvent {
	attrs := projections.IncidentSubjectAttributes{
		Title:       title,
		Summary:     summary,
		SeverityRef: "SEV-1",
		TypeRef:     "outage",
	}
	return s.createNormalizedEvent(ctx, projections.SubjectKindIncident, subjectRef, attrs)
}

func (s *ProjectedEntityLinkingSuite) createAlertEvent(ctx context.Context, subjectRef string, title string, description string) *ent.NormalizedEvent {
	attrs := projections.AlertSubjectAttributes{
		Title:       title,
		Description: description,
		Definition:  "threshold > 1",
	}
	return s.createNormalizedEvent(ctx, projections.SubjectKindAlert, subjectRef, attrs)
}

func (s *ProjectedEntityLinkingSuite) createUserEvent(ctx context.Context, subjectRef string, email string, name string) *ent.NormalizedEvent {
	attrs := projections.UserSubjectAttributes{
		Name:     name,
		Email:    email,
		ChatId:   subjectRef,
		Timezone: "Australia/Perth",
	}
	return s.createNormalizedEvent(ctx, projections.SubjectKindUser, subjectRef, attrs)
}

func (s *ProjectedEntityLinkingSuite) createSystemComponentEvent(ctx context.Context, subjectRef string) *ent.NormalizedEvent {
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: subjectRef,
		Kind:        "service",
		DisplayName: "Service",
	}
	return s.createNormalizedEvent(ctx, projections.SubjectKindSystemComponent, subjectRef, attrs)
}

func (s *ProjectedEntityLinkingSuite) createNormalizedEvent(ctx context.Context, subjectKind projections.SubjectKind, subjectRef string, attrs any) *ent.NormalizedEvent {
	encodedAttrs, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	ev, err := s.Client().NormalizedEvent.Create().
		SetActivityKind(ne.ActivityKindObserved).
		SetProvider("fake").
		SetProviderSource("test").
		SetProviderEventRef("fake:event:" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetSubjectKind(subjectKind.String()).
		SetAttributes(encodedAttrs).
		SetOccurredAt(now).
		SetReceivedAt(now).
		Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *ProjectedEntityLinkingSuite) entityEvidence(ctx context.Context, entityID uuid.UUID) []*ent.KnowledgeEvidence {
	evidence, err := s.Client().KnowledgeEvidence.Query().
		Where(knev.EntityID(entityID)).
		Order(ent.Asc(knev.FieldObservedAt)).
		All(ctx)
	s.Require().NoError(err)
	return evidence
}

func (s *ProjectedEntityLinkingSuite) entityForAlias(ctx context.Context, provider string, subjectRef string) *ent.KnowledgeEntity {
	alias, err := s.Client().KnowledgeEntityAlias.Query().
		Where(knea.Provider(provider), knea.ProviderSubjectRef(subjectRef)).
		WithEntity().
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(alias.Edges.Entity)
	return alias.Edges.Entity
}

func (s *ProjectedEntityLinkingSuite) createKnowledgeEntity(ctx context.Context, kind string, displayName string) *ent.KnowledgeEntity {
	entity, err := s.Client().KnowledgeEntity.Create().
		SetKind(kind).
		SetDisplayName(displayName).
		SetProperties(map[string]any{}).
		Save(ctx)
	s.Require().NoError(err)
	return entity
}

func (s *ProjectedEntityLinkingSuite) createKnowledgeAlias(ctx context.Context, entityID uuid.UUID, ref EntityAliasRef) *ent.KnowledgeEntityAlias {
	alias, err := s.Client().KnowledgeEntityAlias.Create().
		SetEntityID(entityID).
		SetProvider(ref.Provider).
		SetProviderSubjectRef(ref.ProviderSubjectRef).
		Save(ctx)
	s.Require().NoError(err)
	return alias
}
