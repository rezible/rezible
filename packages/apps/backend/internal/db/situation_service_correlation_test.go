package db

import (
	"context"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/pkg/situations"
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
		situation, getErr := h.situations.GetSituation(ctx, membership.SituationID)
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
	signals, loadErr := h.situations.loadSignals(ctx, []uuid.UUID{second}, situations.LoadSignalsOptions{})
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
