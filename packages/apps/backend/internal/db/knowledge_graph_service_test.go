package db

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
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

func (s *KnowledgeGraphServiceSuite) TestOneEventCanStoreMultipleAssertionsForAnAlias() {
	ctx := s.SeedTenantContext()

	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "component:multi",
		Kind:        "service",
		DisplayName: "Multi",
	}
	event := s.createEvent(projections.SubjectKindSystemComponent, "component:multi", time.Now(), attrs)

	makeEvidence := func(assertion string) ent.KnowledgeEvidenceRef {
		aliasRef := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
		aliasRef.SubjectEntityRef = &ent.KnowledgeEntityRef{
			Kind:        "foo",
			Reference:   "multi",
			DisplayName: "Multi",
		}
		return ent.KnowledgeEvidenceRef{
			Kind:            ke.EvidenceKindObserved,
			Assertion:       assertion,
			EffectiveAt:     event.OccurredAt,
			SubjectAliasRef: aliasRef,
		}
	}

	ingestErr := s.knowledgeService().IngestEvidence(ctx, event, makeEvidence("first"), makeEvidence("second"))
	s.Require().NoError(ingestErr)

	count, countErr := s.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(2, count)
}

func (s *KnowledgeGraphServiceSuite) TestAliasCannotBeReassignedToAnotherEntity() {
	ctx := s.SeedTenantContext()

	subjRef := "component:shared"
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: subjRef,
		Kind:        "service",
		DisplayName: "Shared",
	}
	event := s.createEvent(projections.SubjectKindSystemComponent, subjRef, time.Now(), attrs)

	firstAlias := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
	firstAlias.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:      "system_component",
		Reference: "component:first",
	}

	svc := s.knowledgeService()
	firstErr := svc.IngestEvidence(ctx, event, ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       "first",
		EffectiveAt:     event.OccurredAt,
		SubjectAliasRef: firstAlias,
	})
	s.Require().NoError(firstErr)

	secondAlias := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
	secondAlias.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:      "system_component",
		Reference: "component:second",
	}
	secondErr := svc.IngestEvidence(ctx, event, ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       "second",
		EffectiveAt:     event.OccurredAt,
		SubjectAliasRef: secondAlias,
	})
	s.Require().Error(secondErr)
	s.True(errors.Is(secondErr, rez.ErrConflict))

	queryEntities := s.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ReferenceIn(firstAlias.SubjectEntityRef.Reference, secondAlias.SubjectEntityRef.Reference))
	entityCount, countErr := queryEntities.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, entityCount)
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
