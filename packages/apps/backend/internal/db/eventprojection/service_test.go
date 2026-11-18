package eventprojection

import (
	"context"
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
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
)

type ProjectionServiceSuite struct {
	test.Suite
}

func TestProjectionServiceSuite(t *testing.T) {
	suite.Run(t, &ProjectionServiceSuite{
		Suite: test.NewSuite(),
	})
}

func (s *ProjectionServiceSuite) ingestionService(tdb rez.Database) *db.KnowledgeGraphIngestionService {
	knowledge, knowledgeErr := db.NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)
	return knowledge
}

func (s *ProjectionServiceSuite) projectionService(tdb rez.Database) *ProjectionService {
	users, usersErr := db.NewUserService(tdb, mocks.NewMockOrganizationService(s.T()))
	s.Require().NoError(usersErr)

	// Observing an incident publishes nothing and touches no situation.
	incidents, incidentsErr := db.NewIncidentDetailsService(tdb, mocks.NewMockMessageQueue(s.T()), mocks.NewMockSituationService(s.T()))
	s.Require().NoError(incidentsErr)

	knowledge := s.ingestionService(tdb)

	knowledgeQuery, knowledgeQueryErr := db.NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(knowledgeQueryErr)

	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 1,
			},
		}, nil).
		Maybe()
	clock := test.NewClock(time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))
	alerts, alertsErr := db.NewAlertService(rez.DefaultConfig().Alerts, clock, tdb, jobService, db.NewSituationObservationService(jobService))
	s.Require().NoError(alertsErr)

	changePoints, changePointsErr := db.NewChangePointService(tdb)
	s.Require().NoError(changePointsErr)
	service, serviceErr := NewProjectionService(tdb, knowledge, knowledgeQuery, users, incidents, alerts, changePoints)
	s.Require().NoError(serviceErr)

	return service
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
		SetLastOccurredAt(occurredAt).
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
		Predicate:   knr.PredicateDependsOn,
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
	s.Equal(knr.PredicateDependsOn, projectedRelationship.Predicate)
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
	relationshipQuery.Where(knr.PredicateEQ(knr.PredicateDependsOn))
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

func (s *ProjectionServiceSuite) TestSkippedMembershipLeavesExistingMemberUnchanged() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	ingester := db.NewProviderEventIngester(tdb, nil, service, s.ingestionService(tdb))
	suffix := uuid.NewString()
	membershipRef := "membership:" + suffix
	attributes := projections.TeamMembershipEventAttributes{
		Team: projections.TeamMembershipTeamAttributes{
			ProviderResourceRef: rez.ProviderResourceRef{Provider: "slack", ProviderNamespace: "workspace", ResourceRef: "group-" + suffix},
			Name:                "Platform",
			Slug:                "platform-" + suffix,
		},
		User: projections.TeamMembershipUserAttributes{
			ProviderResourceRef: rez.ProviderResourceRef{Provider: "slack", ProviderNamespace: "workspace", ResourceRef: "user-" + suffix},
			Name:                "Avery",
			Email:               "avery-" + suffix + "@example.com",
		},
		Role: "member",
	}
	at := time.Now().UTC()
	first := s.createNormalizedEvent(ctx, tdb, projections.KindTeamMembership, membershipRef, at, attributes)
	s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, first.ID, false))

	// The same membership resource now names another user, which its alias cannot identify.
	other := attributes
	other.User = projections.TeamMembershipUserAttributes{
		ProviderResourceRef: rez.ProviderResourceRef{Provider: "slack", ProviderNamespace: "workspace", ResourceRef: "other-" + suffix},
		Name:                "Blake",
		Email:               "blake-" + suffix + "@example.com",
	}
	second := s.createNormalizedEvent(ctx, tdb, projections.KindTeamMembership, membershipRef, at.Add(time.Minute), other)
	s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, second.ID, false))

	queryReceipt := tdb.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(second.ID))
	receipt, receiptErr := queryReceipt.Only(ctx)
	s.Require().NoError(receiptErr)
	s.Require().Len(receipt.SkippedRefs, 1)
	s.Equal(schematypes.SkippedEvidenceReasonTopologyMismatch, receipt.SkippedRefs[0].Reason)

	member, memberErr := tdb.Client(ctx).User.Query().Where(user.Email(attributes.User.Email)).Only(ctx)
	s.Require().NoError(memberErr)
	s.Equal("Avery", member.Name)
	otherExists, otherErr := tdb.Client(ctx).User.Query().Where(user.Email(other.User.Email)).Exist(ctx)
	s.Require().NoError(otherErr)
	s.False(otherExists)
	membershipCount, membershipCountErr := tdb.Client(ctx).TeamMembership.Query().Count(ctx)
	s.Require().NoError(membershipCountErr)
	s.Equal(1, membershipCount)
}

// listingObservationProcessor models a poll that reports the same statement at each sync time.
type listingObservationProcessor struct{ snapshot bool }

func (p listingObservationProcessor) ProcessProviderEvent(_ context.Context, event rez.ProviderEvent) (ent.NormalizedEvents, error) {
	attributes := projections.SystemComponentEventAttributes{
		Category: kne.CategoryContainer, Kind: "service", DisplayName: "checkout-api", Listing: !p.snapshot,
	}
	if !p.snapshot {
		attributes.ValidFor = 3 * time.Hour
	}
	encoded, encodeErr := projections.EncodeAttributes(attributes)
	if encodeErr != nil {
		return nil, encodeErr
	}
	return ent.NormalizedEvents{{Provider: event.Provider, ProviderNamespace: event.ProviderNamespace,
		ProviderEventSource: event.ProviderEventSource, ProviderEventRef: event.ProviderEventRef,
		ProviderResourceRef: "service:checkout-api", Kind: projections.KindSystemComponent,
		OccurredAt: event.ReceivedAt, ReceivedAt: event.ReceivedAt, Attributes: encoded}}, nil
}

func (s *ProjectionServiceSuite) TestUnchangedListingRemainsFresh() {
	ctx, tdb := s.SetupTestDatabase()
	ingester := db.NewProviderEventIngester(tdb, map[string]rez.ProviderEventProcessor{"test": listingObservationProcessor{}}, s.projectionService(tdb), s.ingestionService(tdb))
	var queued []river.JobArgs
	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().InsertMany(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
			for _, param := range params {
				queued = append(queued, param.Args)
			}
			return make([]*rivertype.JobInsertResult, len(params)), nil
		})
	pipeline, pipelineErr := db.NewProviderEventPipelineService(tdb, jobService, ingester)
	s.Require().NoError(pipelineErr)
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	for _, observedAt := range []time.Time{at, at.Add(2 * time.Hour)} {
		event := rez.ProviderEvent{Provider: "test", ProviderNamespace: "listing-tests", ProviderEventSource: "services",
			ProviderEventRef: "unchanged-listing", ReceivedAt: observedAt}
		s.Require().NoError(pipeline.HandleProcessEventJob(ctx, &jobs.ProcessProviderEventArgs{Event: event}))
		for _, args := range queued {
			switch job := args.(type) {
			case jobs.ProjectNormalizedEvent:
				s.Require().NoError(pipeline.HandleEventProjectionJob(ctx, &river.Job[jobs.ProjectNormalizedEvent]{JobRow: &rivertype.JobRow{Attempt: 1}, Args: job}))
			case jobs.ReobserveNormalizedEvent:
				s.Require().NoError(pipeline.HandleEventReobservationJob(ctx, job))
			}
		}
		queued = nil
	}
	entity, queryErr := tdb.Client(ctx).KnowledgeEntity.Query().Only(ctx)
	s.Require().NoError(queryErr)
	s.False(knowledgegraph.IsStale(entity.ExpiresAt, at.Add(4*time.Hour)), "a repeated listing extends the observation's validity")
}

func (s *ProjectionServiceSuite) TestListingReobservations() {
	for _, scenario := range []struct {
		name                                    string
		beforeProjection, repeatFirst, snapshot bool
	}{
		{name: "project then repeat"},
		{name: "repeat job before projection", beforeProjection: true, repeatFirst: true},
		{name: "projection before repeat job", beforeProjection: true},
		{name: "snapshot replay", snapshot: true},
	} {
		s.Run(scenario.name, func() {
			ctx, tdb := s.SetupTestDatabase()
			ingester := db.NewProviderEventIngester(tdb, map[string]rez.ProviderEventProcessor{"test": listingObservationProcessor{snapshot: scenario.snapshot}}, s.projectionService(tdb), s.ingestionService(tdb))
			at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
			event := rez.ProviderEvent{Provider: "test", ProviderNamespace: "listing-tests", ProviderEventSource: "services", ProviderEventRef: "same-listing", ReceivedAt: at}
			first, firstErr := ingester.ProcessProviderEvent(ctx, event)
			s.Require().NoError(firstErr)
			s.Require().Len(first.NewEventIDs, 1)
			eventID := first.NewEventIDs[0]
			if !scenario.beforeProjection {
				s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, eventID, false))
			}
			event.ReceivedAt = at.Add(2 * time.Hour)
			repeated, repeatErr := ingester.ProcessProviderEvent(ctx, event)
			s.Require().NoError(repeatErr)
			s.Empty(repeated.NewEventIDs)
			s.Require().Len(repeated.Reobservations, 1)
			repeat := repeated.Reobservations[0]
			s.Equal(eventID, repeat.EventID)
			runRepeat := func() { s.Require().NoError(ingester.ReobserveNormalizedEvent(ctx, repeat.EventID, repeat.At)) }
			if scenario.repeatFirst {
				runRepeat()
			}
			s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, eventID, false))
			runRepeat()
			runRepeat()
			s.Require().NoError(ingester.ReobserveNormalizedEvent(ctx, eventID, at.Add(time.Hour)))
			entity, entityErr := tdb.Client(ctx).KnowledgeEntity.Query().Only(ctx)
			s.Require().NoError(entityErr)
			evidence, evidenceErr := tdb.Client(ctx).KnowledgeEvidence.Query().Only(ctx)
			s.Require().NoError(evidenceErr)
			stored, storedErr := tdb.Client(ctx).NormalizedEvent.Query().Only(ctx)
			s.Require().NoError(storedErr)
			s.True(stored.LastOccurredAt.Equal(repeat.At))
			s.Equal("checkout-api", entity.State.DisplayName)
			if scenario.snapshot {
				s.Nil(entity.ExpiresAt)
				s.Nil(evidence.ReobservedAt)
				s.True(entity.LastObservedAt.Equal(at))
			} else {
				s.Require().NotNil(entity.ExpiresAt)
				s.True(entity.ExpiresAt.Equal(at.Add(5 * time.Hour)))
				s.Require().NotNil(evidence.ReobservedAt)
				s.True(evidence.ReobservedAt.Equal(repeat.At))
			}
		})
	}
}

func (s *ProjectionServiceSuite) TestListingRepeatRestoresWithdrawalButRedeliveryDoesNot() {
	ctx, tdb := s.SetupTestDatabase()
	ingester := db.NewProviderEventIngester(tdb, map[string]rez.ProviderEventProcessor{"test": listingObservationProcessor{}}, s.projectionService(tdb), s.ingestionService(tdb))
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	event := rez.ProviderEvent{Provider: "test", ProviderNamespace: "listing-tests", ProviderEventSource: "services", ProviderEventRef: "same-listing", ReceivedAt: at}
	result, processErr := ingester.ProcessProviderEvent(ctx, event)
	s.Require().NoError(processErr)
	s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, result.NewEventIDs[0], false))
	knowledge := s.ingestionService(tdb)
	deleteAt := func(when time.Time) {
		deletion := s.createNormalizedEvent(ctx, tdb, projections.KindSystemComponent, "deleted-checkout", when, projections.SystemComponentEventAttributes{Category: kne.CategoryContainer, Kind: "service", DisplayName: "checkout-api"})
		s.Require().NoError(knowledge.IngestEvidence(ctx, deletion, rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityDeclared, Kind: ke.KindDeleted,
			Assertion: knowledgeAssertionSystemComponentExists, EffectiveAt: when,
			Subject: rez.KnowledgeSubjectRef{Entity: &rez.KnowledgeEntityRef{Category: kne.CategoryContainer, Kind: "service",
				ProviderResourceRef: rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "listing-tests", ResourceRef: "service:checkout-api"}}}}))
	}
	deleteAt(at.Add(time.Hour))
	redelivery, redeliveryErr := ingester.ProcessProviderEvent(ctx, event)
	s.Require().NoError(redeliveryErr)
	s.Empty(redelivery.NewEventIDs)
	s.Empty(redelivery.Reobservations, "an original-time redelivery queues no work")
	entity, entityErr := tdb.Client(ctx).KnowledgeEntity.Query().Only(ctx)
	s.Require().NoError(entityErr)
	s.NotNil(entity.WithdrawnAt)
	event.ReceivedAt = at.Add(2 * time.Hour)
	repeated, repeatErr := ingester.ProcessProviderEvent(ctx, event)
	s.Require().NoError(repeatErr)
	s.Require().Len(repeated.Reobservations, 1)
	repeat := repeated.Reobservations[0]
	s.Require().NoError(ingester.ReobserveNormalizedEvent(ctx, repeat.EventID, repeat.At))
	entity, entityErr = tdb.Client(ctx).KnowledgeEntity.Query().Only(ctx)
	s.Require().NoError(entityErr)
	s.Nil(entity.WithdrawnAt)
	deleteAt(at.Add(3 * time.Hour))
	s.Require().NoError(ingester.ReobserveNormalizedEvent(ctx, repeat.EventID, repeat.At))
	entity, entityErr = tdb.Client(ctx).KnowledgeEntity.Query().Only(ctx)
	s.Require().NoError(entityErr)
	s.NotNil(entity.WithdrawnAt)
}
