package db

import (
	"context"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inc "github.com/rezible/rezible/ent/incident"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	inver "github.com/rezible/rezible/ent/investigationevidencerevision"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/situations"
)

// evaluate runs the situation's evaluation, as its job does. It returns the situation and the deadline the
// evaluation scheduled, if any.
func (s *SituationServiceSuite) evaluate(ctx context.Context, h *situationServiceFixture, id uuid.UUID) (*ent.Situation, *time.Time) {
	h.evaluations = nil
	s.Require().NoError(h.situations.EvaluateSituation(ctx, id))
	situation, getErr := h.situations.GetSituation(ctx, sit.ID(id))
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
	episode := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5)
	s.resolve(ctx, tdb, h, "checkout-errors", 5, 10)
	created := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {episode}})
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
	watched := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5)
	muted := s.alertEpisode(ctx, tdb, h, "payment-errors", 5)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {watched}})
	mutedCandidate := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {muted}})
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
	incidents, management, declare := s.newIncidents(ctx, tdb, h)

	// An unresolved incident holds a situation open until its resolution plus the quiet period.
	episode := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5)
	s.resolve(ctx, tdb, h, "checkout-errors", 5, 10)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {episode}})
	incident := declare(linked.ID)
	_, deadline := s.evaluate(ctx, h, linked.ID)
	s.Nil(deadline, "nothing can close it while the incident is unresolved")

	resolvedAt := h.clock.Now().Add(time.Hour)
	h.evaluations = nil
	resolution := rez.AddIncidentMilestoneParams{IncidentID: incident.ID, Kind: im.KindResolution, Timestamp: resolvedAt}
	_, resolveErr := management.AddMilestone(ctx, resolution)
	s.Require().NoError(resolveErr)
	s.Contains(h.evaluations, jobs.EvaluateSituation{SituationID: linked.ID}, "a milestone requests evaluation")
	resolved, getErr := incidents.GetIncident(ctx, inc.ID(incident.ID))
	s.Require().NoError(getErr)
	s.NotNil(resolved.Edges.Retrospective, "resolving creates the retrospective")
	_, deadline = s.evaluate(ctx, h, linked.ID)
	s.Require().NotNil(deadline)
	s.True(resolvedAt.Add(situations.SituationQuietPeriod).Equal(*deadline))

	// A hold defers closure; a later attachment defers it again.
	first := s.alertEpisode(ctx, tdb, h, "cart-errors", 5)
	s.resolve(ctx, tdb, h, "cart-errors", 5, 10)
	held := s.create(ctx, h, map[string][]uuid.UUID{"Cart": {first}})
	h.clock.Advance(situations.CollectionWindow)
	holdUntil := h.clock.Now().Add(2 * time.Hour)
	_, holdErr := h.situations.SetSituationHold(ctx, held.ID, &rez.SituationHold{Until: &holdUntil})
	s.Require().NoError(holdErr)
	_, deadline = s.evaluate(ctx, h, held.ID)
	s.Require().NotNil(deadline)
	s.True(holdUntil.Equal(*deadline))

	h.clock.Advance(3 * time.Hour)
	second := s.alertEpisode(ctx, tdb, h, "cart-latency", 5)
	s.resolve(ctx, tdb, h, "cart-latency", 5, 10)
	attach := AttachSituationObservationsParams{
		SituationID:  held.ID,
		Observations: episodeRefs(second),
	}
	s.Require().NoError(h.situations.AttachSituationObservations(ctx, attach))
	open, deadline := s.evaluate(ctx, h, held.ID)
	s.Nil(open.ClosedAt)
	s.Require().NotNil(deadline)
	s.True(h.clock.Now().Add(situations.SituationQuietPeriod).Equal(*deadline))
}

func (s *SituationServiceSuite) TestEvaluationRaisesCandidates() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	// A second alert source raises on breadth once the collection window has passed.
	first := s.alertEpisode(ctx, tdb, h, "checkout-errors", 29)
	second := s.alertEpisode(ctx, tdb, h, "checkout-latency", 30)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first, second}})
	collecting, deadline := s.evaluate(ctx, h, candidate.ID)
	s.Nil(collecting.RaisedAt, "only a hard reason raises while collecting")
	s.Empty(collecting.Edges.Judgments)
	s.Require().NotNil(deadline)
	s.True(candidate.CreatedAt.Add(situations.CollectionWindow).Equal(*deadline))

	h.clock.Set(*deadline)
	raised, _ := s.evaluate(ctx, h, candidate.ID)
	s.NotNil(raised.RaisedAt)
	s.NotNil(raised.InvestigationID)
	s.Require().NotEmpty(raised.Edges.Judgments)
	s.Equal(sitjudg.OutcomeNeedsDecision, raised.Edges.Judgments[0].Outcome)
	s.Equal(sitjudg.DecisionRaise, raised.Edges.Judgments[0].Decision)
	s.Equal([]schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonBreadth}, raised.Edges.Judgments[0].CitedReasons)
	s.Equal(situations.JudgeRules, raised.Edges.Judgments[0].Judge)
	s.Equal(sitact.ActionRaised, s.lastAction(raised).Action)
	s.Nil(s.lastAction(raised).UserID)

	// A hard reason raises before the collection window ends.
	var wide []uuid.UUID
	for _, definition := range []string{"cart-a", "cart-b", "cart-c", "cart-d"} {
		wide = append(wide, s.alertEpisode(ctx, tdb, h, definition, 30))
	}
	broad := s.create(ctx, h, map[string][]uuid.UUID{"Cart": wide})
	hardRaised, _ := s.evaluate(ctx, h, broad.ID)
	s.NotNil(hardRaised.RaisedAt)
	s.Equal(sitjudg.OutcomeRaise, hardRaised.Edges.Judgments[0].Outcome)

	// Unchanged facts are judged once.
	lone := s.alertEpisode(ctx, tdb, h, "search-errors", 30)
	held := s.create(ctx, h, map[string][]uuid.UUID{"Search": {lone}})
	h.clock.Advance(situations.CollectionWindow)
	judged, _ := s.evaluate(ctx, h, held.ID)
	rejudged, _ := s.evaluate(ctx, h, held.ID)
	s.Nil(rejudged.RaisedAt)
	s.Equal(sitjudg.DecisionHold, rejudged.Edges.Judgments[0].Decision)
	s.Equal(judged.Edges.Judgments[0].ID, rejudged.Edges.Judgments[0].ID)
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
	_, _, declare := s.newIncidents(ctx, tdb, h)
	episode := s.alertEpisode(ctx, tdb, h, "payment-errors", 25)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {episode}})
	declare(linked.ID)
	linkRaised, getErr := h.situations.GetSituation(ctx, sit.ID(linked.ID))
	s.Require().NoError(getErr)
	s.Equal(sitact.ActionRaised, s.lastAction(linkRaised).Action)
	s.Nil(s.lastAction(linkRaised).UserID)
	var candidates []*ent.Situation
	for _, name := range []string{"a", "b", "c", "d"} {
		first := s.alertEpisode(ctx, tdb, h, name+"-errors", 25, checkout)
		second := s.alertEpisode(ctx, tdb, h, name+"-latency", 25, checkout)
		candidates = append(candidates, s.create(ctx, h, map[string][]uuid.UUID{name: {first, second}}))
	}

	h.clock.Advance(situations.CollectionWindow)
	for i, candidate := range candidates {
		raised, _ := s.evaluate(ctx, h, candidate.ID)
		s.NotNil(raised.RaisedAt)
		if i < situations.AutoInvestigationLimit {
			s.NotNil(raised.InvestigationID)
		} else {
			s.Nil(raised.InvestigationID, "past the limit it raises without an investigation")
		}
	}
}

func (s *SituationServiceSuite) TestEvaluationRefreshesInvestigations() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	episode := s.alertEpisode(ctx, tdb, h, "checkout-errors", 5)
	params := rez.CreateSituationParams{
		Title:        "Checkout errors",
		Observations: episodeRefs(episode),
		Raise:        &rez.RaiseSituationParams{StartInvestigation: true},
	}
	created, createErr := h.situations.CreateSituation(ctx, params)
	s.Require().NoError(createErr)
	revisions := func() int {
		queryRevisions := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(*created.InvestigationID))
		count, countErr := queryRevisions.Count(ctx)
		s.Require().NoError(countErr)
		return count
	}
	remind := func() {
		s.alertEpisode(ctx, tdb, h, "checkout-errors", 5)
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
	_, _, declare := s.newIncidents(ctx, tdb, h)

	// An episode in a raised situation opened before the lookback, which is not counted.
	longAgo := -100 * 24 * 60
	old := s.alertEpisode(ctx, tdb, h, "checkout-errors", longAgo)
	s.resolve(ctx, tdb, h, "checkout-errors", longAgo, longAgo+1)
	forgotten := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {old}})
	_, raiseErr := h.situations.RaiseSituation(ctx, forgotten.ID, rez.RaiseSituationParams{})
	s.Require().NoError(raiseErr)

	// Two earlier episodes of the definition in one incident-linked situation, which counts once.
	first := s.alertEpisode(ctx, tdb, h, "checkout-errors", 0)
	s.resolve(ctx, tdb, h, "checkout-errors", 0, 1)
	second := s.alertEpisode(ctx, tdb, h, "checkout-errors", 10)
	s.resolve(ctx, tdb, h, "checkout-errors", 10, 11)
	linked := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first, second}})
	declare(linked.ID)

	// A third in a situation muted as not noteworthy.
	third := s.alertEpisode(ctx, tdb, h, "checkout-errors", 20)
	s.resolve(ctx, tdb, h, "checkout-errors", 20, 21)
	muted := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {third}})
	_, muteErr := h.situations.SetSituationMute(ctx, muted.ID, &rez.SituationMute{Reason: sit.MuteReasonNotNoteworthy})
	s.Require().NoError(muteErr)

	latest := s.alertEpisode(ctx, tdb, h, "checkout-errors", 30)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {latest}})
	h.clock.Advance(situations.CollectionWindow)
	judged, _ := s.evaluate(ctx, h, candidate.ID)
	s.NotNil(judged.RaisedAt)
	s.Equal([]schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonPastIncident}, judged.Edges.Judgments[0].CitedReasons)

	judgment, judgmentErr := tdb.Client(ctx).SituationJudgment.Get(ctx, judged.Edges.Judgments[0].ID)
	s.Require().NoError(judgmentErr)
	s.Require().Len(judgment.Facts.Observations, 1)
	want := schematypes.SituationPriorOutcomeFacts{Raised: 1, MutedNotNoteworthy: 1, IncidentLinked: 1}
	s.Equal(&want, judgment.Facts.Observations[0].PriorOutcomes)
}

func (s *SituationServiceSuite) TestEvaluationRefreshesTheTargetInvestigationAfterAMerge() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	investigated := func(definition string) (*ent.Situation, uuid.UUID) {
		episode := s.alertEpisode(ctx, tdb, h, definition, 5)
		params := rez.CreateSituationParams{
			Title:        definition,
			Observations: episodeRefs(episode),
			Raise:        &rez.RaiseSituationParams{StartInvestigation: true},
		}
		created, createErr := h.situations.CreateSituation(ctx, params)
		s.Require().NoError(createErr)
		return created, episode
	}
	target, _ := investigated("checkout-errors")
	source, moved := investigated("payment-errors")
	revisions := func() int {
		queryRevisions := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
			Where(inver.InvestigationID(*target.InvestigationID))
		count, countErr := queryRevisions.Count(ctx)
		s.Require().NoError(countErr)
		return count
	}

	mergeParams := rez.MergeSituationsParams{SourceID: source.ID, TargetID: target.ID, Explanation: "same outage"}
	_, mergeErr := h.situations.MergeSituations(ctx, mergeParams)
	s.Require().NoError(mergeErr)
	merged, _ := s.evaluate(ctx, h, target.ID)
	s.Equal(1, revisions(), "the moved evidence reaches the target's investigation")

	observations, loadErr := h.alerts.LoadObservations(ctx, episodeRefs(moved), rez.LoadSituationObservationsOptions{})
	s.Require().NoError(loadErr)
	s.Require().Len(observations, 1)
	s.Equal(observations[0].Revision, s.members(merged)[moved].ObservedRevision)

	s.evaluate(ctx, h, target.ID)
	s.Equal(1, revisions(), "evaluating again records nothing more")
}
