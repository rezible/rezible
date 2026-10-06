package situations

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	inc "github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

var evaluationNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// at is a time `minutes` after evaluationNow.
func at(minutes int) time.Time {
	return evaluationNow.Add(time.Duration(minutes) * time.Minute)
}

// member is a seeding, default-attention alert signal attached at `attached`, finished at `finished` when
// given, and otherwise active for `activeMinutes`.
func member(attached int, finished *int, activeMinutes int) Member {
	signal := Signal{
		EntityID:        uuid.New(),
		Severity:        schematypes.SignalSeverityCritical,
		SignalAttention: ssa.LevelDefault,
		Alert:           &schematypes.SituationAlertFacts{},
	}
	if finished != nil {
		signal.FinishedAt = new(at(*finished))
	} else {
		signal.Alert.Active = true
		signal.Alert.ActiveSeconds = int64(activeMinutes * 60)
	}
	return Member{Signal: signal, AttachedAt: at(attached)}
}

func TestCloseableAt(t *testing.T) {
	resolved := LinkedIncident{ResponseState: inc.ResponseStateResolved, ResolvedAt: new(at(-5))}
	open := LinkedIncident{ResponseState: inc.ResponseStateStarted}

	cases := []struct {
		name       string
		input      EvaluationInput
		want       *time.Time
		wantReason sit.CloseReason
	}{
		{
			name:  "an unresolved incident holds it open",
			input: EvaluationInput{Members: []Member{member(-60, new(-50), 0)}, Incidents: []LinkedIncident{open}},
		},
		{
			name:  "a raised situation with unfinished evidence stays open",
			input: EvaluationInput{Raised: true, Members: []Member{member(-60, nil, 60)}},
		},
		{
			name:       "a candidate with unfinished evidence expires after its maximum age from creation",
			input:      EvaluationInput{CreatedAt: at(-60), Members: []Member{member(-60, nil, 600)}},
			want:       new(at(-60).Add(CandidateMaxAge)),
			wantReason: sit.CloseReasonExpired,
		},
		{
			name:       "a muted candidate is dismissed",
			input:      EvaluationInput{CreatedAt: at(-60), Muted: true, Members: []Member{member(-60, nil, 60)}},
			want:       new(at(-60).Add(CandidateMaxAge)),
			wantReason: sit.CloseReasonDismissed,
		},
		{
			name:       "finished evidence closes after the quiet period from the latest finish",
			input:      EvaluationInput{Raised: true, Members: []Member{member(-60, new(-50), 0), member(-55, new(-20), 0)}},
			want:       new(at(-20).Add(SituationQuietPeriod)),
			wantReason: sit.CloseReasonStabilized,
		},
		{
			name:       "a later attachment defers closure",
			input:      EvaluationInput{Raised: true, Members: []Member{member(-60, new(-50), 0), member(-10, new(-40), 0)}},
			want:       new(at(-10).Add(SituationQuietPeriod)),
			wantReason: sit.CloseReasonStabilized,
		},
		{
			name: "a later incident resolution defers closure",
			input: EvaluationInput{
				Raised:    true,
				Members:   []Member{member(-60, new(-50), 0)},
				Incidents: []LinkedIncident{resolved},
			},
			want:       new(at(-5).Add(SituationQuietPeriod)),
			wantReason: sit.CloseReasonStabilized,
		},
		{
			name:       "a hold defers closure",
			input:      EvaluationInput{HoldUntil: new(at(120)), Members: []Member{member(-60, new(-50), 0)}},
			want:       new(at(120)),
			wantReason: sit.CloseReasonExpired,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input.Now = evaluationNow
			closeableAt := tc.input.CloseableAt()
			if tc.want == nil {
				assert.Nil(t, closeableAt)
				return
			}
			if assert.NotNil(t, closeableAt) {
				assert.Equal(t, *tc.want, *closeableAt)
			}
			assert.Equal(t, tc.wantReason, tc.input.CloseReason())
		})
	}
}

func TestNextDeadline(t *testing.T) {
	cases := []struct {
		name  string
		input EvaluationInput
		want  *time.Time
	}{
		{
			name:  "closure of a raised situation",
			input: EvaluationInput{Raised: true, CreatedAt: at(-60), Members: []Member{member(-60, new(-10), 0)}},
			want:  new(at(-10).Add(SituationQuietPeriod)),
		},
		{
			name:  "the end of a candidate's collection window",
			input: EvaluationInput{CreatedAt: at(0), Members: []Member{member(0, nil, 0)}},
			want:  new(at(0).Add(CollectionWindow)),
		},
		{
			name:  "an alert reaching the novelty active time",
			input: EvaluationInput{CreatedAt: at(-10), Members: []Member{member(-10, nil, 2)}},
			want:  new(at(3)),
		},
		{
			name:  "an alert reaching persistence without a baseline",
			input: EvaluationInput{CreatedAt: at(-10), Members: []Member{member(-10, nil, 20)}},
			want:  new(at(10)),
		},
		{
			name:  "a muted candidate waits only for closure",
			input: EvaluationInput{CreatedAt: at(-10), Muted: true, Members: []Member{member(-10, nil, 20)}},
			want:  new(at(-10).Add(CandidateMaxAge)),
		},
		{
			name: "the retry of an unavailable judge",
			input: EvaluationInput{
				CreatedAt: at(-60),
				Members:   []Member{member(-60, new(-10), 0)},
				Latest:    &LatestJudgment{JudgedAt: at(-1), Judge: JudgeUnavailable},
			},
			want: new(at(-1).Add(JudgeRetryAfter)),
		},
		{
			name:  "nothing for a raised situation with unfinished evidence",
			input: EvaluationInput{Raised: true, CreatedAt: at(-10), Members: []Member{member(-10, nil, 20)}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input.Now = evaluationNow
			assert.Equal(t, tc.want, tc.input.NextDeadline())
		})
	}
}
