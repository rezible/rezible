package db

import (
	"context"
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
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
)

// processNotified processes the captured observation notifications, as the process-situation-observation job
// does.
func (s *SituationServiceSuite) processNotified(ctx context.Context, h *situationServiceFixture) {
	notified := h.notified
	h.notified = nil
	for _, episodeID := range notified {
		s.Require().NoError(h.situations.ProcessObservation(ctx, rez.SituationObservationRef{AlertEpisodeID: episodeID}))
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

// placed returns the situation observing each episode, loaded as GetSituation does; nil when not placed.
func (s *SituationServiceSuite) placed(ctx context.Context, tdb rez.Database, h *situationServiceFixture, episodeIDs ...uuid.UUID) []*ent.Situation {
	placed := make([]*ent.Situation, len(episodeIDs))
	for i, episodeID := range episodeIDs {
		querySituation := tdb.Client(ctx).Situation.Query().
			Where(sit.HasObservationsWith(sitobs.AlertEpisodeID(episodeID)))
		situationID, queryErr := querySituation.OnlyID(ctx)
		if ent.IsNotFound(queryErr) {
			continue
		}
		s.Require().NoError(queryErr)
		situation, getErr := h.situations.GetSituation(ctx, sit.ID(situationID))
		s.Require().NoError(getErr)
		placed[i] = situation
	}
	return placed
}

func (s *SituationServiceSuite) TestCorrelationBySharedEntity() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	first := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5, checkout)
	second := s.alertEpisode(ctx, tdb, h, "checkout-latency", 6, checkout)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, first, second)
	s.Require().NotNil(placed[0])
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID, "one candidate holds both")
	s.Nil(placed[0].RaisedAt)
	s.Equal("checkout-errors", placed[0].Title)
	s.True(situationTestStart.Add(5 * time.Minute).Equal(placed[0].OpenedAt))
	s.Len(s.members(placed[0]), 2)

	// A second run for an observation that passed the first membership check finds it placed under the lock.
	observations, loadErr := h.situations.LoadObservations(ctx, episodeRefs(second), rez.LoadSituationObservationsOptions{})
	s.Require().NoError(loadErr)
	s.Require().NoError(h.situations.correlateObservation(ctx, observations[rez.SituationObservationRef{AlertEpisodeID: second}]))
	s.Require().NoError(h.situations.ProcessObservation(ctx, rez.SituationObservationRef{AlertEpisodeID: second}))
	memberships, countErr := tdb.Client(ctx).SituationObservation.Query().Count(ctx)
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

	first := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5, checkoutFromA)
	second := s.alertEpisode(ctx, tdb, h, "checkout-latency", 6, checkoutFromB)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, first, second)
	s.Require().NotNil(placed[0])
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID)
}

func (s *SituationServiceSuite) TestCorrelationByDependency() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	search := s.entityRef(kne.CategoryContainer, "search")
	s.ingestRelationship(ctx, tdb, h, checkout, knr.PredicateCalls, search)

	checkoutErrors := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5, checkout)
	s.processNotified(ctx, h)
	searchLatency := s.alertEpisode(ctx, tdb, h, "search-latency", 6, search)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, checkoutErrors, searchLatency)
	s.Require().NotNil(placed[1])
	s.Equal(placed[0].ID, placed[1].ID)
}

func (s *SituationServiceSuite) TestObservationJobRetriesOnlyWhenAMergeDeletedASubject() {
	_, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	knowledgeFK := &pgconn.PgError{Code: "23503", ConstraintName: ent.KnowledgeForeignKeyConstraints[0]}
	otherFK := &pgconn.PgError{Code: "23503", ConstraintName: "discussion_threads_users_user"}

	s.True(jobs.IsRetryableError(h.situations.observationJobError(fmt.Errorf("correlate: %w", errs.ErrKnowledgeSubjectMerged))))
	s.True(jobs.IsRetryableError(h.situations.observationJobError(fmt.Errorf("attach: %w", knowledgeFK))))

	otherErr := h.situations.observationJobError(fmt.Errorf("attach: %w", otherFK))
	s.False(jobs.IsRetryableError(otherErr))
	s.ErrorIs(otherErr, otherFK)
}

func (s *SituationServiceSuite) TestCorrelationEligibility() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	quick := s.alertEpisode(ctx, tdb, h, "checkout-errors", 28)
	s.resolve(ctx, tdb, h, "checkout-errors", 28, 29)
	old := s.alertEpisode(ctx, tdb, h, "payment-errors", -120)
	s.resolve(ctx, tdb, h, "payment-errors", -120, -100)
	s.processNotified(ctx, h)

	placed := s.placed(ctx, tdb, h, quick, old)
	s.NotNil(placed[0], "an observation resolved before processing is placed while in its grace")
	s.Nil(placed[1], "an observation that finished long ago is not placed")
}

func (s *SituationServiceSuite) TestAClosedSituationIsNotJoined() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	earlier := s.alertEpisode(ctx, tdb, h, "checkout-errors", 0, checkout)
	s.resolve(ctx, tdb, h, "checkout-errors", 0, 1)
	s.processNotified(ctx, h)
	closed := s.placed(ctx, tdb, h, earlier)[0]
	s.Require().NotNil(closed)
	_, closeErr := h.situations.CloseSituation(ctx, closed.ID, rez.CloseSituationParams{})
	s.Require().NoError(closeErr)

	later := s.alertEpisode(ctx, tdb, h, "checkout-errors", 20, checkout)
	s.processNotified(ctx, h)
	recurrence := s.placed(ctx, tdb, h, later)[0]
	s.Require().NotNil(recurrence)
	s.NotEqual(closed.ID, recurrence.ID, "a closed situation is not joined")
}

func (s *SituationServiceSuite) TestNonSeedingObservationsJoinButDoNotStart() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	info := s.alertEpisodeWithSeverity(ctx, tdb, h, "checkout-info", schematypes.SignalSeverityInfo, 5, checkout)
	s.processNotified(ctx, h)
	s.Nil(s.placed(ctx, tdb, h, info)[0], "an info observation does not start a candidate")

	critical := s.alertEpisode(ctx, tdb, h, "checkout-errors", 6, checkout)
	s.alertEpisodeWithSeverity(ctx, tdb, h, "checkout-info", schematypes.SignalSeverityInfo, 5, checkout)
	s.processNotified(ctx, h)
	placed := s.placed(ctx, tdb, h, critical, info)
	s.Require().NotNil(placed[1], "an info observation joins")
	s.Equal(placed[0].ID, placed[1].ID)
}

func (s *SituationServiceSuite) TestProcessingAPlacedObservationRefreshesItsSituation() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	payments := s.entityRef(kne.CategoryContainer, "payments")

	episode := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5, checkout)
	s.processNotified(ctx, h)
	s.alertEpisode(ctx, tdb, h, "checkout-errors", 5, payments)
	s.processNotified(ctx, h)

	situation := s.placed(ctx, tdb, h, episode)[0]
	s.Require().NotNil(situation)
	s.Equal(map[uuid.UUID]bool{s.entityID(ctx, h, checkout): true, s.entityID(ctx, h, payments): true}, s.entities(situation))
	situations, countErr := tdb.Client(ctx).Situation.Query().Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, situations)
}
