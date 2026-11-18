package db

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

type ChangePointServiceSuite struct {
	test.Suite
}

func TestChangePointServiceSuite(t *testing.T) {
	suite.Run(t, &ChangePointServiceSuite{Suite: test.NewSuite()})
}

func (s *ChangePointServiceSuite) TestRecordChangePoint() {
	at := time.Date(2026, 10, 7, 14, 0, 0, 0, time.UTC)
	for name, reverse := range map[string]bool{"started first": false, "succeeded first": true} {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			graph := newTestGraph(s.Require(), tdb)
			service, serviceErr := NewChangePointService(tdb)
			s.Require().NoError(serviceErr)
			checkout := graph.entity(kne.CategoryContainer, "service", "test", "checkout")
			repository := graph.entity(kne.CategoryCode, "repository", "test", "acme/checkout")
			involvement := func(ref rez.KnowledgeEntityRef, role string) rez.KnowledgeEvidenceRef {
				evidence := graph.observeEntity(ref, ref.ProviderResourceRef.ResourceRef)
				evidence.Assertion = role
				return evidence
			}
			deployment := func(offset time.Duration, attributes string, withRepository bool) *ent.NormalizedEvent {
				evidence := []rez.KnowledgeEvidenceRef{involvement(checkout, projections.AssertionDeploymentService)}
				if withRepository {
					evidence = append(evidence, involvement(repository, projections.AssertionDeploymentRepository))
				}
				return graph.changePointEvent(ctx, "deployment:run-1", projections.KindDeployment, at.Add(offset), attributes, evidence...)
			}
			// rows returns the change point's involvement rows by role.
			rows := func() map[string]*ent.ChangePointInvolvement {
				involved, involvedErr := tdb.Client(ctx).ChangePointInvolvement.Query().All(ctx)
				s.Require().NoError(involvedErr)
				byRole := make(map[string]*ent.ChangePointInvolvement, len(involved))
				for _, row := range involved {
					byRole[row.Role] = row
				}
				s.Require().Len(byRole, len(involved), "one row per role")
				return byRole
			}
			rowIDs := func(byRole map[string]*ent.ChangePointInvolvement) map[string]uuid.UUID {
				ids := make(map[string]uuid.UUID, len(byRole))
				for role, row := range byRole {
					ids[role] = row.ID
				}
				return ids
			}
			started := deployment(0, `{"status":"started","version":"old"}`, true)
			succeeded := deployment(time.Minute, `{"status":"succeeded"}`, true)
			events := []*ent.NormalizedEvent{started, succeeded}
			if reverse {
				slices.Reverse(events)
			}
			for _, event := range events {
				s.Require().NoError(service.RecordChangePoint(ctx, event, event.OccurredAt))
			}

			point, queryErr := tdb.Client(ctx).ChangePoint.Query().Only(ctx)
			s.Require().NoError(queryErr)
			s.Equal(succeeded.ID, point.LatestEventID)
			s.True(succeeded.OccurredAt.Equal(point.OccurredAt))
			involved := rows()
			s.Require().Len(involved, 2)
			for role, row := range involved {
				evidence, evidenceErr := tdb.Client(ctx).KnowledgeEvidence.Get(ctx, row.EvidenceID)
				s.Require().NoError(evidenceErr)
				s.Equal(succeeded.ID, evidence.EventID, "%s cites the latest event's evidence", role)
				s.Equal(evidence.SubjectAliasID, row.SubjectAliasID)
			}

			for _, event := range []*ent.NormalizedEvent{started, succeeded} {
				s.Require().NoError(service.RecordChangePoint(ctx, event, event.OccurredAt))
				after, getErr := tdb.Client(ctx).ChangePoint.Get(ctx, point.ID)
				s.Require().NoError(getErr)
				s.Equal(point.LatestEventID, after.LatestEventID, "an older event and reprocessing change nothing")
				s.True(point.UpdatedAt.Equal(after.UpdatedAt))
				s.Equal(rowIDs(involved), rowIDs(rows()), "the involvement rows are unchanged")
			}

			withoutRepository := deployment(2*time.Minute, `{"status":"failed"}`, false)
			s.Require().NoError(service.RecordChangePoint(ctx, withoutRepository, withoutRepository.OccurredAt))
			replaced := rows()
			s.Require().Len(replaced, 1, "a later event naming no repository leaves no repository row")
			s.Contains(replaced, projections.AssertionDeploymentService)
		})
	}
}

func (s *ChangePointServiceSuite) TestListChangePoints() {
	ctx, tdb := s.SetupTestDatabase()
	graph := newTestGraph(s.Require(), tdb)
	service, serviceErr := NewChangePointService(tdb)
	s.Require().NoError(serviceErr)
	at := time.Date(2026, 10, 7, 14, 0, 0, 0, time.UTC)
	checkout := graph.entity(kne.CategoryContainer, "service", "test", "checkout")
	checkoutOtherAlias := checkout
	checkout.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout"}
	checkoutOtherAlias.LinkingAttributes = checkout.LinkingAttributes
	checkoutOtherAlias.ProviderResourceRef.ResourceRef = "checkout-other"
	search := graph.entity(kne.CategoryContainer, "service", "test", "search")
	involvement := func(ref rez.KnowledgeEntityRef, role string) rez.KnowledgeEvidenceRef {
		evidence := graph.observeEntity(ref, ref.ProviderResourceRef.ResourceRef)
		evidence.Assertion = role
		return evidence
	}
	record := func(ref string, offset time.Duration, attributes string, evidence ...rez.KnowledgeEvidenceRef) *ent.ChangePoint {
		event := graph.changePointEvent(ctx, ref, projections.KindDeployment, at.Add(offset), attributes, evidence...)
		s.Require().NoError(service.RecordChangePoint(ctx, event, event.OccurredAt))
		points, pointsErr := tdb.Client(ctx).ChangePoint.Query().All(ctx)
		s.Require().NoError(pointsErr)
		for _, point := range points {
			if point.ProviderResourceRef == ref {
				return point
			}
		}
		s.FailNow("change point missing")
		return nil
	}
	first := record("first", 0, `{"status":"succeeded","environment":{"name":"production"},"repository":{"linking_attributes":{"repository.full_name":"acme/checkout"}},"merge":{"into_default_branch":true,"merge_commit_sha":"abc"}}`,
		involvement(checkout, projections.AssertionDeploymentService), involvement(checkoutOtherAlias, projections.AssertionDeploymentRepository))
	second := record("second", time.Minute, `{"status":"failed","environment":{"name":"staging"},"repository":{"linking_attributes":{"repository.full_name":null}}}`, involvement(checkout, projections.AssertionDeploymentService))
	third := record("third", time.Minute, `{"status":"started"}`, involvement(checkout, projections.AssertionDeploymentRepository))
	record("replaced", 0, `{}`, involvement(checkout, projections.AssertionDeploymentService))
	record("replaced", 2*time.Minute, `{}`, involvement(search, projections.AssertionDeploymentService))
	params := rez.ListChangePointsParams{EntityID: graph.entityID(ctx, checkout), Kind: projections.KindDeployment, Limit: 20}
	list := func(params rez.ListChangePointsParams) *rez.ChangePoints {
		points, listErr := service.ListChangePoints(ctx, params)
		s.Require().NoError(listErr)
		return points
	}
	ids := func(points *rez.ChangePoints) []uuid.UUID {
		result := make([]uuid.UUID, 0, len(points.ChangePoints))
		for _, point := range points.ChangePoints {
			result = append(result, point.ChangePoint.ID)
		}
		return result
	}
	tied := []uuid.UUID{second.ID, third.ID}
	slices.SortFunc(tied, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	all := list(params)
	s.Equal(append(tied, first.ID), ids(all), "latest involvement only, no duplicate for multiple aliases")
	for _, point := range all.ChangePoints {
		s.Require().NotNil(point.Involvement)
		s.Equal(params.EntityID, *point.Involvement.Edges.SubjectAlias.EntityID)
		if point.ChangePoint.ID == first.ID {
			s.Len(point.Involved, 2)
			lowest := slices.MinFunc(point.Involved, func(a, b *ent.ChangePointInvolvement) int {
				return strings.Compare(a.EvidenceID.String(), b.EvidenceID.String())
			})
			s.Equal(lowest.EvidenceID, point.Involvement.EvidenceID, "lowest evidence ID wins")
		}
	}
	createMerged := tdb.Client(ctx).KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service").
		SetMergedIntoID(params.EntityID)
	merged := createMerged.SaveX(ctx)
	byMerged := params
	byMerged.EntityID = merged.ID
	s.Equal(ids(all), ids(list(byMerged)), "an entity merged into checkout lists checkout's change points")
	params.Limit = 2
	limited := list(params)
	s.True(limited.Cut)
	s.Equal(tied, ids(limited))
	params.Limit = 3
	s.False(list(params).Cut)
	params.Limit = 20
	for _, fromInclusive := range []bool{false, true} {
		for _, toInclusive := range []bool{false, true} {
			bounded := params
			bounded.From, bounded.To = at, at.Add(time.Minute)
			bounded.FromInclusive, bounded.ToInclusive = fromInclusive, toInclusive
			expected := []uuid.UUID{}
			if toInclusive {
				expected = append(expected, tied...)
			}
			if fromInclusive {
				expected = append(expected, first.ID)
			}
			s.Equal(expected, ids(list(bounded)))
		}
	}
	params.Roles = []string{projections.AssertionDeploymentRepository}
	byRole := list(params)
	s.Len(byRole.ChangePoints, 2)
	for _, point := range byRole.ChangePoints {
		s.Equal(projections.AssertionDeploymentRepository, point.Involvement.Role)
	}
	params.Roles = nil
	for _, filter := range []rez.ChangePointAttribute{
		{Path: projections.DeploymentStatus, Value: "succeeded"},
		{Path: projections.DeploymentEnvironmentName, Value: "production"},
		{Path: projections.DeploymentRepositoryFullName, Value: "acme/checkout"},
		{Path: projections.CodeChangeIntoDefaultBranch, Value: "true"},
		{Path: projections.CodeChangeMergeCommitSha, Value: "abc"},
	} {
		filtered := params
		filtered.AttributeEquals = []rez.ChangePointAttribute{filter}
		s.Equal([]uuid.UUID{first.ID}, ids(list(filtered)))
	}
	params.AttributeAbsent = []projections.AttributePath{projections.DeploymentRepositoryFullName}
	s.Equal(tied, ids(list(params)), "both missing paths and JSON null are absent")
	otherCtx, _ := s.SetupTestDatabase()
	other, otherErr := service.ListChangePoints(otherCtx, params)
	s.Require().NoError(otherErr)
	s.Empty(other.ChangePoints, "another tenant lists none")

	for _, invalid := range []rez.ListChangePointsParams{
		{Kind: projections.KindDeployment, EntityID: params.EntityID},
		{Kind: projections.KindDeployment, EntityID: params.EntityID, Limit: 101},
		{EntityID: params.EntityID, Limit: 1},
		{Kind: "push", EntityID: params.EntityID, Limit: 1},
		{Kind: projections.KindDeployment, EntityID: params.EntityID, Limit: 1,
			Roles: []string{projections.AssertionChangeRepository}},
		{Kind: projections.KindDeployment, EntityID: params.EntityID, Limit: 1,
			AttributeEquals: []rez.ChangePointAttribute{{Path: "status' OR 'a", Value: "x"}}},
	} {
		_, invalidErr := service.ListChangePoints(ctx, invalid)
		s.ErrorIs(invalidErr, errs.ErrInvalidInput)
	}
}
