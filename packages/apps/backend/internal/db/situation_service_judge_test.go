package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/situations"
)

// stubJudge answers as the model judge would, checking that no transaction is held while it runs.
type stubJudge struct {
	s       *SituationServiceSuite
	answer  rezai.SituationJudgeOutput
	fail    error
	inputs  []rezai.SituationJudgeInput
	running func()
}

func (j *stubJudge) Run(ctx context.Context, input rezai.SituationJudgeInput) (rezai.SituationJudgeOutput, error) {
	j.s.Nil(ent.TxFromContext(ctx), "no transaction is held while the model runs")
	j.inputs = append(j.inputs, input)
	if j.running != nil {
		j.running()
	}
	return j.answer, j.fail
}

// persistentCandidate is a candidate whose one alert, about a checkout service, has been active long enough
// to persist, so its outcome needs a decision. The collection window has passed.
func (s *SituationServiceSuite) persistentCandidate(ctx context.Context, tdb rez.Database, h *situationServiceFixture, judge *stubJudge) *ent.Situation {
	h.situations.judge = judge
	signal := s.signal(ctx, tdb, h, "checkout-errors", 0, s.entityRef(kne.CategoryContainer, "checkout"))
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	h.clock.Advance(situations.CollectionWindow)
	return candidate
}

func (s *SituationServiceSuite) countJudgments(ctx context.Context, tdb rez.Database, situationID uuid.UUID) int {
	queryJudgments := tdb.Client(ctx).SituationJudgment.Query().
		Where(sitjudg.SituationID(situationID))
	count, countErr := queryJudgments.Count(ctx)
	s.Require().NoError(countErr)
	return count
}

func (s *SituationServiceSuite) TestModelJudgeDecidesOnceForUnchangedFacts() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	judge := &stubJudge{s: s, answer: rezai.SituationJudgeOutput{Decision: "hold", Reasons: []schematypes.SituationRaiseReason{}, Explanation: "Only one alert."}}
	candidate := s.persistentCandidate(ctx, tdb, h, judge)

	held, _ := s.evaluate(ctx, h, candidate.ID)
	s.evaluate(ctx, h, candidate.ID)
	s.Nil(held.RaisedAt)
	s.Require().Len(judge.inputs, 1, "unchanged facts are judged once")
	s.Equal(situations.JudgeModel, held.Edges.LatestJudgment.Judge)
	s.Equal(sitjudg.OutcomeNeedsDecision, held.Edges.LatestJudgment.Outcome)
	facts := judge.inputs[0].Facts
	s.Require().Len(facts.Entities, 1, "the judge sees the situation's entities")
	s.Equal([]uuid.UUID{facts.Entities[0].ID}, facts.Signals[0].EntityIDs)
	s.Equal(1, s.countJudgments(ctx, tdb, candidate.ID))
}

func (s *SituationServiceSuite) TestConcurrentEvaluationsCallJudgeTwiceButRecordOnce() {
	ctx, tdb := s.SetupTestDatabase()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	h := s.newFixture(tdb)
	judge := &stubJudge{s: s, answer: rezai.SituationJudgeOutput{Decision: "hold", Explanation: "Only one alert."}}
	candidate := s.persistentCandidate(ctx, tdb, h, judge)
	entered := make(chan struct{})
	release := make(chan struct{})
	judge.running = func() {
		if len(judge.inputs) == 1 {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
			}
		}
	}

	// Pause the first model call before starting the second evaluation. This orders access to the
	// shared stub's inputs without sleeps, while keeping both evaluations in flight.
	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- h.situations.EvaluateSituation(ctx, candidate.ID)
	}()
	s.T().Cleanup(func() { cancel(); <-done })
	select {
	case <-entered:
	case <-ctx.Done():
		s.Require().FailNow("the first evaluation did not reach the judge", "%v", ctx.Err())
	}

	// Capture reserves no in-flight judgment, so unchanged facts still trigger a second model call.
	held, _ := s.evaluate(ctx, h, candidate.ID)
	s.Require().Len(judge.inputs, 2)
	s.Require().NotNil(held.Edges.LatestJudgment)
	judgmentID := held.Edges.LatestJudgment.ID
	close(release)
	s.Require().NoError(<-done)

	// The second evaluation committed first; apply discards the first call's now-redundant answer.
	final, getErr := h.situations.GetSituation(ctx, candidate.ID)
	s.Require().NoError(getErr)
	s.Nil(final.RaisedAt)
	s.Equal(judgmentID, *final.LatestJudgmentID)
	s.Equal(1, s.countJudgments(ctx, tdb, candidate.ID))
}

func (s *SituationServiceSuite) TestModelJudgeRaises() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	answer := rezai.SituationJudgeOutput{
		Decision:    "raise",
		Reasons:     []schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonPersistence},
		Explanation: "Checkout has failed for half an hour.",
	}
	candidate := s.persistentCandidate(ctx, tdb, h, &stubJudge{s: s, answer: answer})

	raised, _ := s.evaluate(ctx, h, candidate.ID)
	s.NotNil(raised.RaisedAt)
	s.NotNil(raised.Edges.Investigation)
	s.Equal(situations.JudgeModel, raised.Edges.LatestJudgment.Judge)
	s.Equal(answer.Reasons, raised.Edges.LatestJudgment.CitedReasons)
}

func (s *SituationServiceSuite) TestModelJudgeRejectsMalformedAndInvalidAnswers() {
	answers := map[string]*stubJudge{
		"malformed": {fail: fmt.Errorf("%w: reasons is not an array", rezai.ErrWorkflowInvalidOutput)},
		"invalid": {answer: rezai.SituationJudgeOutput{
			Decision:    "raise",
			Reasons:     []schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonBreadth},
			Explanation: "Many alerts.",
		}},
	}
	for name, judge := range answers {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			h := s.newFixture(tdb)
			judge.s = s
			candidate := s.persistentCandidate(ctx, tdb, h, judge)

			rejected, _ := s.evaluate(ctx, h, candidate.ID)
			s.Nil(rejected.RaisedAt)
			s.Equal(situations.JudgeModelRejected, rejected.Edges.LatestJudgment.Judge)
			s.Equal(sitjudg.DecisionHold, rejected.Edges.LatestJudgment.Decision)
		})
	}
}

func (s *SituationServiceSuite) TestModelJudgeRetriesWhenUnavailable() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	judge := &stubJudge{s: s, fail: errors.New("model overloaded")}
	candidate := s.persistentCandidate(ctx, tdb, h, judge)

	unavailable, deadline := s.evaluate(ctx, h, candidate.ID)
	s.Nil(unavailable.RaisedAt)
	s.Equal(situations.JudgeUnavailable, unavailable.Edges.LatestJudgment.Judge)
	s.Require().NotNil(deadline)
	s.True(h.clock.Now().Add(situations.JudgeRetryAfter).Equal(*deadline), "the retry is scheduled as a deadline")

	h.clock.Advance(situations.JudgeRetryAfter - time.Second)
	s.evaluate(ctx, h, candidate.ID)
	s.Len(judge.inputs, 1, "not retried before JudgeRetryAfter")

	h.clock.Set(*deadline)
	judge.fail = nil
	judge.answer = rezai.SituationJudgeOutput{Decision: "hold", Explanation: "Only one alert."}
	judged, _ := s.evaluate(ctx, h, candidate.ID)
	s.Len(judge.inputs, 2)
	s.Equal(situations.JudgeModel, judged.Edges.LatestJudgment.Judge)
}

func (s *SituationServiceSuite) TestModelJudgeAnswerIsDiscardedWhenFactsChange() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	judge := &stubJudge{s: s, answer: rezai.SituationJudgeOutput{
		Decision:    "raise",
		Reasons:     []schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonPersistence},
		Explanation: "Checkout has failed for half an hour.",
	}}
	candidate := s.persistentCandidate(ctx, tdb, h, judge)
	judge.running = func() {
		s.resolve(ctx, tdb, h, "checkout-errors", 0, 31)
	}

	discarded, _ := s.evaluate(ctx, h, candidate.ID)
	s.Nil(discarded.RaisedAt)
	s.Nil(discarded.Edges.LatestJudgment)
	s.Contains(h.evaluations, jobs.EvaluateSituation{SituationID: candidate.ID}, "an evaluation is requested now")
}
