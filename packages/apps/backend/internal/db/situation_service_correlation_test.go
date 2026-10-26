package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
)

// processNotified processes the captured signal notifications, as the process-situation-signal job does.
func (s *SituationServiceSuite) processNotified(ctx context.Context, h *situationServiceFixture) {
	notified := h.notified
	h.notified = nil
	for _, signalID := range notified {
		s.Require().NoError(h.situations.ProcessSignal(ctx, signalID))
	}
}

// resolve records the resolution, `end` minutes after situationTestStart, of the definition's alert that
// started `start` minutes after it.
func (s *SituationServiceSuite) resolve(ctx context.Context, tdb rez.Database, h *situationServiceFixture, definition string, start, end int) {
	endedAt := situationTestStart.Add(time.Duration(end) * time.Minute)
	params := rez.RecordAlertInstanceParams{
		Event: s.testEvent(ctx, tdb, definition, endedAt),
		Definition: rez.AlertDefinitionValues{
			KnowledgeEntityID: s.entityID(ctx, h, s.entityRef(kne.CategorySignal, definition)),
			Title:             definition,
		},
		Instance: rez.AlertInstanceValues{
			InstanceID: definition,
			Severity:   schematypes.SignalSeverityCritical,
			StartedAt:  situationTestStart.Add(time.Duration(start) * time.Minute),
			EndedAt:    &endedAt,
		},
	}
	_, recordErr := h.alerts.RecordAlertInstance(ctx, params)
	s.Require().NoError(recordErr)
}

// placed returns the situations holding each signal, loaded as GetSituation does; nil when not placed.
func (s *SituationServiceSuite) placed(ctx context.Context, tdb rez.Database, h *situationServiceFixture, signalIDs ...uuid.UUID) []*ent.Situation {
	placed := make([]*ent.Situation, len(signalIDs))
	for i, signalID := range signalIDs {
		queryMembership := tdb.Client(ctx).SituationSignal.Query().
			Where(sitsig.KnowledgeEntityID(signalID))
		membership, queryErr := queryMembership.Only(ctx)
		if ent.IsNotFound(queryErr) {
			continue
		}
		s.Require().NoError(queryErr)
		situation, getErr := h.situations.GetSituation(ctx, sit.ID(membership.SituationID))
		s.Require().NoError(getErr)
		placed[i] = situation
	}
	return placed
}

func (s *SituationServiceSuite) TestCorrelationBySharedEntity() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	first := s.signal(ctx, tdb, h, "checkout-errors", 5, checkout)
	second := s.signal(ctx, tdb, h, "checkout-latency", 6, checkout)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, first, second)
	s.Require().NotNil(placed[0])
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID, "one candidate holds both")
	s.Nil(placed[0].RaisedAt)
	s.Equal("checkout-errors", placed[0].Title)
	s.True(situationTestStart.Add(5 * time.Minute).Equal(placed[0].OpenedAt))
	members := s.members(placed[0])
	s.Equal(sitsig.MatchKindSeed, members[first].MatchKind)
	s.Equal(sitsig.MatchKindSharedEntity, members[second].MatchKind)

	// A second run for a signal that passed the first membership check finds it placed under the lock.
	signals, loadErr := h.situations.loadSignals(ctx, []uuid.UUID{second}, rez.LoadSituationSignalsOptions{})
	s.Require().NoError(loadErr)
	s.Require().NoError(h.situations.correlateSignal(ctx, signals[second]))
	s.Require().NoError(h.situations.ProcessSignal(ctx, second))
	memberships, countErr := tdb.Client(ctx).SituationSignal.Query().Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(2, memberships)
}

func (s *SituationServiceSuite) TestCorrelationAcrossInstallationsThroughSharedService() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkoutLinking := projections.LinkingAttributes{
		projections.LinkingAttributeServiceName: "checkout-api",
	}
	checkoutFromA := s.entityRef(kne.CategoryContainer, "installation-a:checkout-api")
	checkoutFromA.LinkingAttributes = checkoutLinking
	checkoutFromB := s.entityRef(kne.CategoryContainer, "installation-b:checkout-api")
	checkoutFromB.LinkingAttributes = checkoutLinking

	first := s.signal(ctx, tdb, h, "checkout-errors", 5, checkoutFromA)
	second := s.signal(ctx, tdb, h, "checkout-latency", 6, checkoutFromB)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, first, second)
	s.Require().NotNil(placed[0])
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID)
	s.Equal(sitsig.MatchKindSharedEntity, s.members(placed[1])[second].MatchKind)
}

func (s *SituationServiceSuite) TestCorrelationByDependency() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	search := s.entityRef(kne.CategoryContainer, "search")
	s.ingestRelationship(ctx, tdb, h, checkout, knr.PredicateCalls, search)
	calls, callsErr := tdb.Client(ctx).KnowledgeRelationship.Query().Where(knr.PredicateEQ(knr.PredicateCalls)).Only(ctx)
	s.Require().NoError(callsErr)

	checkoutErrors := s.signal(ctx, tdb, h, "checkout-errors", 5, checkout)
	s.processNotified(ctx, h)
	searchLatency := s.signal(ctx, tdb, h, "search-latency", 6, search)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, checkoutErrors, searchLatency)
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID)
	member := s.members(placed[1])[searchLatency]
	s.Equal(sitsig.MatchKindDependency, member.MatchKind)
	s.Equal(&calls.ID, member.ViaRelationshipID)
}

func (s *SituationServiceSuite) TestAttachingThroughAMergedAwayRelationshipRetries() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	client := tdb.Client(ctx)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	search := s.entityRef(kne.CategoryContainer, "search")
	checkoutErrors := s.signal(ctx, tdb, h, "checkout-errors", 5, checkout)
	s.processNotified(ctx, h)
	situation := s.placed(ctx, tdb, h, checkoutErrors)[0]
	s.Require().NotNil(situation)
	searchLatency := s.signal(ctx, tdb, h, "search-latency", 6, search)

	// Correlation read a relationship that an entity merge then folded into another and deleted; the
	// membership's foreign key refuses it.
	mergedAway := uuid.New()
	attach := AttachSituationSignalsParams{
		SituationID:       situation.ID,
		GroupTitle:        "Search",
		SignalEntityIDs:   []uuid.UUID{searchLatency},
		MatchKind:         sitsig.MatchKindDependency,
		ViaRelationshipID: &mergedAway,
	}
	attachErr := h.situations.AttachSituationSignals(ctx, attach)

	pgErr, isPgErr := errors.AsType[*pgconn.PgError](attachErr)
	s.Require().True(isPgErr, "the foreign key refuses the membership")
	s.Equal("23503", pgErr.Code)
	s.True(jobs.IsRetryableError(h.situations.signalJobError(attachErr)), "the correlation job runs again on the merged graph")
	s.False(client.SituationSignal.Query().Where(sitsig.KnowledgeEntityID(searchLatency)).ExistX(ctx))
}

func (s *SituationServiceSuite) TestSignalJobRetriesOnlyWhenAMergeDeletedASubject() {
	_, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	knowledgeFK := &pgconn.PgError{Code: "23503", ConstraintName: ent.KnowledgeForeignKeyConstraints[0]}
	otherFK := &pgconn.PgError{Code: "23503", ConstraintName: "discussion_threads_users_user"}

	s.True(jobs.IsRetryableError(h.situations.signalJobError(fmt.Errorf("correlate: %w", errs.ErrKnowledgeSubjectMerged))))
	s.True(jobs.IsRetryableError(h.situations.signalJobError(fmt.Errorf("attach: %w", knowledgeFK))))

	otherErr := h.situations.signalJobError(fmt.Errorf("attach: %w", otherFK))
	s.False(jobs.IsRetryableError(otherErr))
	s.ErrorIs(otherErr, otherFK)
}

func (s *SituationServiceSuite) TestAnEntityMergeMovesAViaRelationshipToTheKeptOne() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	client := tdb.Client(ctx)
	service := func(resource, name string) rez.KnowledgeEntityRef {
		ref := s.entityRef(kne.CategoryContainer, resource)
		ref.Kind = knowledgegraph.EntityKindService
		ref.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: name}
		return ref
	}
	// `checkout-api` is seen first, so a merge keeps it and folds checkout's call to search into its own.
	checkoutAPI := service("installation-a:checkout-api", "checkout-api")
	checkout := service("installation-b:checkout", "checkout")
	search := s.entityRef(kne.CategoryContainer, "search")
	s.ingestRelationship(ctx, tdb, h, checkoutAPI, knr.PredicateCalls, search)
	s.ingestRelationship(ctx, tdb, h, checkout, knr.PredicateCalls, search)
	keptID := s.entityID(ctx, h, checkoutAPI)
	otherID := s.entityID(ctx, h, checkout)
	queryCalls := client.KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateCalls))
	callsBySource := make(map[uuid.UUID]uuid.UUID)
	for _, calls := range queryCalls.AllX(ctx) {
		callsBySource[calls.SourceEntityID] = calls.ID
	}
	keptCalls, foldedCalls := callsBySource[keptID], callsBySource[otherID]

	checkoutErrors := s.signal(ctx, tdb, h, "checkout-errors", 5, checkoutAPI)
	s.processNotified(ctx, h)
	situation := s.placed(ctx, tdb, h, checkoutErrors)[0]
	s.Require().NotNil(situation)
	searchLatency := s.signal(ctx, tdb, h, "search-latency", 6, search)
	attach := AttachSituationSignalsParams{
		SituationID:       situation.ID,
		GroupTitle:        "Search",
		SignalEntityIDs:   []uuid.UUID{searchLatency},
		MatchKind:         sitsig.MatchKindDependency,
		ViaRelationshipID: &foldedCalls,
	}
	s.Require().NoError(h.situations.AttachSituationSignals(ctx, attach))

	proposal := client.KnowledgeProposal.Query().OnlyX(ctx)
	merge := rez.MergeEntitiesParams{ProposalID: proposal.ID, EntityIDs: [2]uuid.UUID{keptID, otherID}}
	result, mergeErr := h.knowledge.MergeEntities(ctx, merge)

	s.Require().NoError(mergeErr)
	s.Equal(keptID, result.SurvivorID)
	s.False(client.KnowledgeRelationship.Query().Where(knr.ID(foldedCalls)).ExistX(ctx))
	membership := client.SituationSignal.Query().Where(sitsig.KnowledgeEntityID(searchLatency)).OnlyX(ctx)
	s.Equal(&keptCalls, membership.ViaRelationshipID)
}

func (s *SituationServiceSuite) TestMembershipKeepsTheCauseOfOtherConstraintErrors() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkoutErrors := s.signal(ctx, tdb, h, "checkout-errors", 5)
	s.processNotified(ctx, h)
	situation := s.placed(ctx, tdb, h, checkoutErrors)[0]
	s.Require().NotNil(situation)
	searchLatency := s.signal(ctx, tdb, h, "search-latency", 6)
	signals, loadErr := h.situations.LoadSignals(ctx, []uuid.UUID{searchLatency}, rez.LoadSituationSignalsOptions{})
	s.Require().NoError(loadErr)

	// A group that does not exist violates a foreign key, not the one-situation-per-signal index.
	membership := situationMembership{groupID: uuid.New(), matchKind: sitsig.MatchKindManual}
	createErr := h.situations.createMemberships(ctx, situation.ID, membership, signals, situationTestStart)

	s.Require().Error(createErr)
	s.NotErrorIs(createErr, errs.ErrConflict)
	pgErr, isPgErr := errors.AsType[*pgconn.PgError](createErr)
	s.Require().True(isPgErr)
	s.Equal("23503", pgErr.Code)
}

func (s *SituationServiceSuite) TestCorrelationEligibility() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	quick := s.signal(ctx, tdb, h, "checkout-errors", 28)
	s.resolve(ctx, tdb, h, "checkout-errors", 28, 29)
	old := s.signal(ctx, tdb, h, "payment-errors", -120)
	s.resolve(ctx, tdb, h, "payment-errors", -120, -100)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, quick, old)
	s.NotNil(placed[0], "a signal resolved before processing is placed while in its grace")
	s.Nil(placed[1], "a signal that finished long ago is not placed")
}

func (s *SituationServiceSuite) TestNewCandidateRecurrence() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	earlier := s.signal(ctx, tdb, h, "checkout-errors", 0, checkout)
	s.resolve(ctx, tdb, h, "checkout-errors", 0, 1)
	s.processNotified(ctx, h)
	closed := s.placed(ctx, tdb, h, earlier)[0]
	s.Require().NotNil(closed)
	_, closeErr := h.situations.CloseSituation(ctx, closed.ID, rez.CloseSituationParams{})
	s.Require().NoError(closeErr)

	later := s.signal(ctx, tdb, h, "checkout-errors", 20, checkout)
	s.processNotified(ctx, h)
	recurrence := s.placed(ctx, tdb, h, later)[0]
	s.Require().NotNil(recurrence)
	s.NotEqual(closed.ID, recurrence.ID, "a closed situation is not joined")
	s.Require().Len(recurrence.Edges.Links, 1)
	s.Equal(sitlink.KindRecurrenceOf, recurrence.Edges.Links[0].Kind)
	s.Equal(closed.ID, recurrence.Edges.Links[0].LinkedSituationID)
}

func (s *SituationServiceSuite) TestNonSeedingSignalsJoinButDoNotStart() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	info := s.signalWithSeverity(ctx, tdb, h, "checkout-info", schematypes.SignalSeverityInfo, 5, checkout)
	s.processNotified(ctx, h)
	s.Nil(s.placed(ctx, tdb, h, info)[0], "an info signal does not start a candidate")

	critical := s.signal(ctx, tdb, h, "checkout-errors", 6, checkout)
	s.signalWithSeverity(ctx, tdb, h, "checkout-info", schematypes.SignalSeverityInfo, 5, checkout)
	s.processNotified(ctx, h)
	placed := s.placed(ctx, tdb, h, critical, info)
	s.Require().NotNil(placed[1], "an info signal joins")
	s.Equal(placed[0].ID, placed[1].ID)
}

func (s *SituationServiceSuite) TestProcessingAPlacedSignalRefreshesItsSituation() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	payments := s.entityRef(kne.CategoryContainer, "payments")

	signal := s.signal(ctx, tdb, h, "checkout-errors", 5, checkout)
	s.processNotified(ctx, h)
	s.signal(ctx, tdb, h, "checkout-errors", 5, payments)
	s.processNotified(ctx, h)

	situation := s.placed(ctx, tdb, h, signal)[0]
	s.Require().NotNil(situation)
	s.Equal(map[uuid.UUID]bool{s.entityID(ctx, h, checkout): true, s.entityID(ctx, h, payments): true}, s.entities(situation))
	situations, countErr := tdb.Client(ctx).Situation.Query().Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, situations)
}
