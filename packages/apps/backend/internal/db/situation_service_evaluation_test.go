package db

import (
	"context"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inc "github.com/rezible/rezible/ent/incident"
	inver "github.com/rezible/rezible/ent/investigationevidencerevision"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/situations"
)

// evaluate runs the situation's evaluation, as its job does. It returns the situation and the deadline the
// evaluation scheduled, if any.
func (s *SituationServiceSuite) evaluate(ctx context.Context, h *situationServiceFixture, id uuid.UUID) (*ent.Situation, *time.Time) {
	h.evaluations = nil
	s.Require().NoError(h.situations.EvaluateSituation(ctx, id))
	situation, getErr := h.situations.GetSituation(ctx, id)
	s.Require().NoError(getErr)
	var deadline *time.Time
	for _, evaluation := range h.evaluations {
		if evaluation.SituationID == id && evaluation.DueAt != nil {
			deadline = evaluation.DueAt
		}
	}
	return situation, deadline
}

func (s *SituationServiceSuite) lastAction(situation *ent.Situation) *ent.SituationAction {
	return situation.Edges.Actions[len(situation.Edges.Actions)-1]
}

func (s *SituationServiceSuite) TestEvaluationClosesFinishedSituations() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	// The alert finished at minute 15, before the situation was created at minute 30.
	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	s.resolve(ctx, tdb, h, "checkout-errors", 5, 10)
	created := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	_, raiseErr := h.situations.RaiseSituation(ctx, created.ID, rez.RaiseSituationParams{})
	s.Require().NoError(raiseErr)
	closesAt := created.CreatedAt.Add(situations.SituationQuietPeriod)

	open, deadline := s.evaluate(ctx, h, created.ID)
	s.Nil(open.ClosedAt)
	s.Require().NotNil(deadline)
	s.True(closesAt.Equal(*deadline), "the next deadline is the attachment plus the quiet period")

	h.clock.Set(closesAt.Add(time.Minute))
	closed, _ := s.evaluate(ctx, h, created.ID)
	s.Equal(sit.CloseReasonStabilized, *closed.CloseReason)
	s.True(closesAt.Equal(*closed.ClosedAt), "it closes at the computed time")
	s.Equal(sitact.ActionClosed, s.lastAction(closed).Action)
	s.True(h.clock.Now().Equal(s.lastAction(closed).At), "the action records the processing time")
	s.Nil(s.lastAction(closed).UserID)
}

func (s *SituationServiceSuite) TestEvaluationExpiresCandidates() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	watched := s.signal(ctx, tdb, h, "checkout-errors", 5)
	muted := s.signal(ctx, tdb, h, "payment-errors", 5)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {watched}}, watched)
	mutedCandidate := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {muted}}, muted)
	_, muteErr := h.situations.SetSituationMute(ctx, mutedCandidate.ID, &rez.SituationMute{Reason: sit.MuteReasonExpected})
	s.Require().NoError(muteErr)

	h.clock.Set(candidate.CreatedAt.Add(situations.CandidateMaxAge))
	expired, _ := s.evaluate(ctx, h, candidate.ID)
	s.Equal(sit.CloseReasonExpired, *expired.CloseReason, "a never-raised candidate expires with unfinished evidence")
	dismissed, _ := s.evaluate(ctx, h, mutedCandidate.ID)
	s.Equal(sit.CloseReasonDismissed, *dismissed.CloseReason)
}

func (s *SituationServiceSuite) TestEvaluationClosureWaitsForIncidentsHoldsAndAttachments() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	incidents, newIncident := s.newIncidents(ctx, tdb, h)

	// An unresolved incident holds a situation open until its resolution plus the quiet period.
	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	s.resolve(ctx, tdb, h, "checkout-errors", 5, 10)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	incident, linkErr := incidents.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		newIncident(m)
		m.AddSituationIDs(linked.ID)
	})
	s.Require().NoError(linkErr)
	_, deadline := s.evaluate(ctx, h, linked.ID)
	s.Nil(deadline, "nothing can close it while the incident is unresolved")

	resolvedAt := h.clock.Now().Add(time.Hour)
	h.evaluations = nil
	_, resolveErr := incidents.Set(ctx, incident.ID, func(m *ent.IncidentMutation) {
		m.SetResponseState(inc.ResponseStateResolved)
		m.SetResolvedAt(resolvedAt)
	})
	s.Require().NoError(resolveErr)
	s.Contains(h.evaluations, jobs.EvaluateSituation{SituationID: linked.ID}, "an incident write requests evaluation")
	_, deadline = s.evaluate(ctx, h, linked.ID)
	s.Require().NotNil(deadline)
	s.True(resolvedAt.Add(situations.SituationQuietPeriod).Equal(*deadline))

	// A hold defers closure; a later attachment defers it again.
	first := s.signal(ctx, tdb, h, "cart-errors", 5)
	s.resolve(ctx, tdb, h, "cart-errors", 5, 10)
	held := s.create(ctx, h, map[string][]uuid.UUID{"Cart": {first}}, first)
	h.clock.Advance(situations.CollectionWindow)
	holdUntil := h.clock.Now().Add(2 * time.Hour)
	_, holdErr := h.situations.SetSituationHold(ctx, held.ID, &rez.SituationHold{Until: &holdUntil})
	s.Require().NoError(holdErr)
	_, deadline = s.evaluate(ctx, h, held.ID)
	s.Require().NotNil(deadline)
	s.True(holdUntil.Equal(*deadline))

	h.clock.Advance(3 * time.Hour)
	second := s.signal(ctx, tdb, h, "cart-latency", 5)
	s.resolve(ctx, tdb, h, "cart-latency", 5, 10)
	attach := AttachSituationSignalsParams{
		SituationID:     held.ID,
		GroupTitle:      "Cart",
		SignalEntityIDs: []uuid.UUID{second},
		MatchKind:       sitsig.MatchKindManual,
	}
	s.Require().NoError(h.situations.AttachSituationSignals(ctx, attach))
	open, deadline := s.evaluate(ctx, h, held.ID)
	s.Nil(open.ClosedAt)
	s.Require().NotNil(deadline)
	s.True(h.clock.Now().Add(situations.SituationQuietPeriod).Equal(*deadline))
}

func (s *SituationServiceSuite) TestEvaluationRaisesCandidates() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	// A second alert source raises on breadth once the collection window has passed.
	first := s.signal(ctx, tdb, h, "checkout-errors", 29)
	second := s.signal(ctx, tdb, h, "checkout-latency", 30)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first, second}}, first)
	collecting, deadline := s.evaluate(ctx, h, candidate.ID)
	s.Nil(collecting.RaisedAt, "only a hard reason raises while collecting")
	s.Nil(collecting.Edges.LatestJudgment)
	s.Require().NotNil(deadline)
	s.True(candidate.CreatedAt.Add(situations.CollectionWindow).Equal(*deadline))

	h.clock.Set(*deadline)
	raised, _ := s.evaluate(ctx, h, candidate.ID)
	s.NotNil(raised.RaisedAt)
	s.NotNil(raised.Edges.Investigation)
	s.Require().NotNil(raised.Edges.LatestJudgment)
	s.Equal(sitjudg.OutcomeNeedsDecision, raised.Edges.LatestJudgment.Outcome)
	s.Equal(sitjudg.DecisionRaise, raised.Edges.LatestJudgment.Decision)
	s.Equal([]schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonBreadth}, raised.Edges.LatestJudgment.CitedReasons)
	s.Equal(situations.JudgeRules, raised.Edges.LatestJudgment.Judge)
	s.Equal(sitact.ActionRaised, s.lastAction(raised).Action)
	s.Nil(s.lastAction(raised).UserID)

	// A hard reason raises before the collection window ends.
	var wide []uuid.UUID
	for _, definition := range []string{"cart-a", "cart-b", "cart-c", "cart-d"} {
		wide = append(wide, s.signal(ctx, tdb, h, definition, 30))
	}
	broad := s.create(ctx, h, map[string][]uuid.UUID{"Cart": wide}, wide[0])
	hardRaised, _ := s.evaluate(ctx, h, broad.ID)
	s.NotNil(hardRaised.RaisedAt)
	s.Equal(sitjudg.OutcomeRaise, hardRaised.Edges.LatestJudgment.Outcome)

	// Unchanged facts are judged once.
	lone := s.signal(ctx, tdb, h, "search-errors", 30)
	held := s.create(ctx, h, map[string][]uuid.UUID{"Search": {lone}}, lone)
	h.clock.Advance(situations.CollectionWindow)
	judged, _ := s.evaluate(ctx, h, held.ID)
	rejudged, _ := s.evaluate(ctx, h, held.ID)
	s.Nil(rejudged.RaisedAt)
	s.Equal(sitjudg.DecisionHold, rejudged.Edges.LatestJudgment.Decision)
	s.Equal(judged.Edges.LatestJudgment.ID, rejudged.Edges.LatestJudgment.ID)
	queryJudgments := tdb.Client(ctx).SituationJudgment.Query().
		Where(sitjudg.SituationID(held.ID))
	judgments, countErr := queryJudgments.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, judgments)
}

func (s *SituationServiceSuite) TestEvaluationLimitsAutomaticInvestigations() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")

	// An incident link made without a user raises with no user, but is not an automatic raise.
	incidents, newIncident := s.newIncidents(ctx, tdb, h)
	signal := s.signal(ctx, tdb, h, "payment-errors", 25)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {signal}}, signal)
	_, linkErr := incidents.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		newIncident(m)
		m.AddSituationIDs(linked.ID)
	})
	s.Require().NoError(linkErr)
	linkRaised, getErr := h.situations.GetSituation(ctx, linked.ID)
	s.Require().NoError(getErr)
	s.Equal(sitact.ActionRaised, s.lastAction(linkRaised).Action)
	s.Nil(s.lastAction(linkRaised).UserID)
	var candidates []*ent.Situation
	for _, name := range []string{"a", "b", "c", "d"} {
		first := s.signal(ctx, tdb, h, name+"-errors", 25, checkout)
		second := s.signal(ctx, tdb, h, name+"-latency", 25, checkout)
		candidates = append(candidates, s.create(ctx, h, map[string][]uuid.UUID{name: {first, second}}, first))
	}

	h.clock.Advance(situations.CollectionWindow)
	for i, candidate := range candidates {
		raised, _ := s.evaluate(ctx, h, candidate.ID)
		s.NotNil(raised.RaisedAt)
		if i < situations.AutoInvestigationLimit {
			s.NotNil(raised.Edges.Investigation)
		} else {
			s.Nil(raised.Edges.Investigation, "past the limit it raises without an investigation")
		}
	}
}

func (s *SituationServiceSuite) TestEvaluationRefreshesInvestigations() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	params := rez.CreateSituationParams{
		Title:             "Checkout errors",
		SeedEntityID:      signal,
		ObservationGroups: []rez.SituationObservationGroupParams{{Title: "Checkout", SignalEntityIDs: []uuid.UUID{signal}}},
		Raise:             &rez.RaiseSituationParams{StartInvestigation: true},
	}
	created, createErr := h.situations.CreateSituation(ctx, params)
	s.Require().NoError(createErr)
	revisions := func() int {
		queryRevisions := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(created.Edges.Investigation.InvestigationID))
		count, countErr := queryRevisions.Count(ctx)
		s.Require().NoError(countErr)
		return count
	}
	remind := func() {
		s.signal(ctx, tdb, h, "checkout-errors", 5)
	}

	s.evaluate(ctx, h, created.ID)
	s.Equal(0, revisions(), "no evidence revision without a change")

	remind()
	s.evaluate(ctx, h, created.ID)
	s.evaluate(ctx, h, created.ID)
	s.Equal(1, revisions(), "one evidence revision per change")

	_, muteErr := h.situations.SetSituationMute(ctx, created.ID, &rez.SituationMute{Reason: sit.MuteReasonExpected})
	s.Require().NoError(muteErr)
	remind()
	remind()
	s.evaluate(ctx, h, created.ID)
	s.Equal(1, revisions(), "none while muted")

	_, unmuteErr := h.situations.SetSituationMute(ctx, created.ID, nil)
	s.Require().NoError(unmuteErr)
	s.evaluate(ctx, h, created.ID)
	s.Equal(2, revisions(), "one catch-up after unmute")
}

func (s *SituationServiceSuite) TestEvaluationCountsPriorOutcomesOfSources() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	incidents, newIncident := s.newIncidents(ctx, tdb, h)

	// An episode in a raised situation opened before the lookback, which is not counted.
	longAgo := -100 * 24 * 60
	old := s.signal(ctx, tdb, h, "checkout-errors", longAgo)
	s.resolve(ctx, tdb, h, "checkout-errors", longAgo, longAgo+1)
	forgotten := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {old}}, old)
	_, raiseErr := h.situations.RaiseSituation(ctx, forgotten.ID, rez.RaiseSituationParams{})
	s.Require().NoError(raiseErr)

	// Two earlier episodes of the definition in one incident-linked situation, which counts once.
	first := s.signal(ctx, tdb, h, "checkout-errors", 0)
	s.resolve(ctx, tdb, h, "checkout-errors", 0, 1)
	second := s.signal(ctx, tdb, h, "checkout-errors", 10)
	s.resolve(ctx, tdb, h, "checkout-errors", 10, 11)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first, second}}, first)
	_, linkErr := incidents.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		newIncident(m)
		m.AddSituationIDs(linked.ID)
	})
	s.Require().NoError(linkErr)

	// A third in a situation muted as not noteworthy.
	third := s.signal(ctx, tdb, h, "checkout-errors", 20)
	s.resolve(ctx, tdb, h, "checkout-errors", 20, 21)
	muted := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {third}}, third)
	_, muteErr := h.situations.SetSituationMute(ctx, muted.ID, &rez.SituationMute{Reason: sit.MuteReasonNotNoteworthy})
	s.Require().NoError(muteErr)

	latest := s.signal(ctx, tdb, h, "checkout-errors", 30)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {latest}}, latest)
	h.clock.Advance(situations.CollectionWindow)
	judged, _ := s.evaluate(ctx, h, candidate.ID)
	s.NotNil(judged.RaisedAt)
	s.Equal([]schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonPastIncident}, judged.Edges.LatestJudgment.CitedReasons)

	judgment, judgmentErr := tdb.Client(ctx).SituationJudgment.Get(ctx, judged.Edges.LatestJudgment.ID)
	s.Require().NoError(judgmentErr)
	s.Require().Len(judgment.Facts.Signals, 1)
	want := schematypes.SituationPriorOutcomeFacts{Raised: 1, MutedNotNoteworthy: 1, IncidentLinked: 1}
	s.Equal(&want, judgment.Facts.Signals[0].PriorOutcomes)
}

func (s *SituationServiceSuite) TestEvaluationRefreshesTheTargetInvestigationAfterAMerge() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	investigated := func(definition string) (*ent.Situation, uuid.UUID) {
		signal := s.signal(ctx, tdb, h, definition, 5)
		params := rez.CreateSituationParams{
			Title:             definition,
			SeedEntityID:      signal,
			ObservationGroups: []rez.SituationObservationGroupParams{{Title: definition, SignalEntityIDs: []uuid.UUID{signal}}},
			Raise:             &rez.RaiseSituationParams{StartInvestigation: true},
		}
		created, createErr := h.situations.CreateSituation(ctx, params)
		s.Require().NoError(createErr)
		return created, signal
	}
	target, _ := investigated("checkout-errors")
	source, moved := investigated("payment-errors")
	revisions := func() int {
		queryRevisions := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(target.Edges.Investigation.InvestigationID))
		count, countErr := queryRevisions.Count(ctx)
		s.Require().NoError(countErr)
		return count
	}

	mergeParams := rez.MergeSituationsParams{SourceID: source.ID, TargetID: target.ID, Explanation: "same outage"}
	_, mergeErr := h.situations.MergeSituations(ctx, mergeParams)
	s.Require().NoError(mergeErr)
	merged, _ := s.evaluate(ctx, h, target.ID)
	s.Equal(1, revisions(), "the moved evidence reaches the target's investigation")

	signals, loadErr := h.alerts.LoadSignals(ctx, []uuid.UUID{moved}, situations.LoadSignalsOptions{})
	s.Require().NoError(loadErr)
	s.Require().Len(signals, 1)
	s.Equal(signals[0].Revision, s.members(merged)[moved].ObservedRevision)

	s.evaluate(ctx, h, target.ID)
	s.Equal(1, revisions(), "evaluating again records nothing more")
}
