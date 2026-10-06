package situations

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// activeAlert is a default-attention seeding alert from its own source, active for the given time, with
// three weeks of observed history and no earlier occurrences.
func activeAlert(ref string, active time.Duration) schematypes.SituationSignalFacts {
	return schematypes.SituationSignalFacts{
		Ref:            ref,
		EntityID:       uuid.New(),
		SourceEntityID: new(uuid.New()),
		Seeding:        true,
		Attention:      string(ssa.LevelDefault),
		Alert: &schematypes.SituationAlertFacts{
			Active:              true,
			ActiveGroupCount:    1,
			ActiveSeconds:       int64(active / time.Second),
			InstanceCount:       1,
			ActiveInstanceCount: 1,
			Baseline: schematypes.SituationBaselineFacts{
				HasSufficientHistory:   true,
				ObservedHistoryAgeDays: 21,
			},
		},
	}
}

// withAlert is the signal with a copy of its alert facts changed by edit.
func withAlert(signal schematypes.SituationSignalFacts, edit func(*schematypes.SituationAlertFacts)) schematypes.SituationSignalFacts {
	alert := *signal.Alert
	edit(&alert)
	signal.Alert = &alert
	return signal
}

func withAttention(signal schematypes.SituationSignalFacts, level ssa.Level) schematypes.SituationSignalFacts {
	signal.Attention = string(level)
	return signal
}

func withPastIncident(signal schematypes.SituationSignalFacts) schematypes.SituationSignalFacts {
	signal.PriorOutcomes = &schematypes.SituationPriorOutcomeFacts{IncidentLinked: 1}
	return signal
}

func reason(t *testing.T, reasons Reasons, name schematypes.SituationRaiseReason) schematypes.SituationReasonResult {
	t.Helper()
	for _, result := range reasons {
		if result.Reason == name {
			return result
		}
	}
	require.FailNow(t, "missing reason "+string(name))
	return schematypes.SituationReasonResult{}
}

func TestCheckReasons(t *testing.T) {
	type expectation struct {
		met, hard bool
	}
	seenTwice := func(alert *schematypes.SituationAlertFacts) { alert.Baseline.Occurrences = 2 }
	fiveMinuteMedian := func(alert *schematypes.SituationAlertFacts) { alert.Baseline.MedianDurationSeconds = new(int64(300)) }
	cases := []struct {
		name      string
		signals   []schematypes.SituationSignalFacts
		incidents int
		expect    map[schematypes.SituationRaiseReason]expectation
	}{
		{
			name:      "a linked incident is hard",
			signals:   []schematypes.SituationSignalFacts{activeAlert("s1", time.Minute)},
			incidents: 1,
			expect:    map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonLinkedIncident: {met: true, hard: true}},
		},
		{
			name:    "two sources are breadth",
			signals: []schematypes.SituationSignalFacts{activeAlert("s1", time.Minute), activeAlert("s2", time.Minute)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonBreadth: {met: true}},
		},
		{
			name: "four sources are hard breadth",
			signals: []schematypes.SituationSignalFacts{
				activeAlert("s1", time.Minute),
				activeAlert("s2", time.Minute),
				activeAlert("s3", time.Minute),
				activeAlert("s4", time.Minute),
			},
			expect: map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonBreadth: {met: true, hard: true}},
		},
		{
			name: "three active groups of one alert are breadth",
			signals: []schematypes.SituationSignalFacts{
				withAlert(activeAlert("s1", time.Minute), func(alert *schematypes.SituationAlertFacts) { alert.ActiveGroupCount = 3 }),
			},
			expect: map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonBreadth: {met: true}},
		},
		{
			name: "grouped instances count once",
			signals: []schematypes.SituationSignalFacts{
				withAlert(activeAlert("s1", time.Minute), func(alert *schematypes.SituationAlertFacts) {
					alert.ActiveInstanceCount = 6
					alert.ActiveGroupCount = 2
				}),
			},
			expect: map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonBreadth: {}},
		},
		{
			name:    "a rare alert active long enough is novel",
			signals: []schematypes.SituationSignalFacts{activeAlert("s1", 6*time.Minute)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonNovelty: {met: true}},
		},
		{
			name:    "a rare alert not yet active long enough is not novel",
			signals: []schematypes.SituationSignalFacts{activeAlert("s1", 4*time.Minute)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonNovelty: {}},
		},
		{
			name:    "a frequent alert is not novel",
			signals: []schematypes.SituationSignalFacts{withAlert(activeAlert("s1", time.Hour), seenTwice)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonNovelty: {}},
		},
		{
			name:    "persistence without a baseline at its fixed time",
			signals: []schematypes.SituationSignalFacts{activeAlert("s1", PersistenceNoBaseline)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonPersistence: {met: true}},
		},
		{
			name:    "persistence with a baseline at a multiple of the median",
			signals: []schematypes.SituationSignalFacts{withAlert(activeAlert("s1", 15*time.Minute), fiveMinuteMedian)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonPersistence: {met: true}},
		},
		{
			name:    "no persistence below the baseline threshold",
			signals: []schematypes.SituationSignalFacts{withAlert(activeAlert("s1", 14*time.Minute), fiveMinuteMedian)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonPersistence: {}},
		},
		{
			name:    "a source with an earlier incident",
			signals: []schematypes.SituationSignalFacts{withPastIncident(activeAlert("s1", time.Minute))},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonPastIncident: {met: true}},
		},
		{
			name:    "the watch-only gate leaves a lone watch-only source unmet",
			signals: []schematypes.SituationSignalFacts{withAttention(withPastIncident(activeAlert("s1", time.Minute)), ssa.LevelWatchOnly)},
			expect:  map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonPastIncident: {}},
		},
		{
			name: "two watch-only sources pass the gate",
			signals: []schematypes.SituationSignalFacts{
				withAttention(activeAlert("s1", time.Minute), ssa.LevelWatchOnly),
				withAttention(activeAlert("s2", time.Minute), ssa.LevelWatchOnly),
			},
			expect: map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonBreadth: {met: true}},
		},
		{
			name:      "the watch-only gate keeps a linked incident",
			signals:   []schematypes.SituationSignalFacts{withAttention(activeAlert("s1", time.Minute), ssa.LevelWatchOnly)},
			incidents: 1,
			expect:    map[schematypes.SituationRaiseReason]expectation{schematypes.SituationRaiseReasonLinkedIncident: {met: true, hard: true}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			facts := schematypes.SituationFacts{Signals: tc.signals}
			for range tc.incidents {
				facts.LinkedIncidents = append(facts.LinkedIncidents, schematypes.SituationIncidentFacts{ID: uuid.New()})
			}
			reasons := CheckReasons(facts)
			assert.Len(t, reasons, 5)
			for name, want := range tc.expect {
				result := reason(t, reasons, name)
				assert.Equal(t, want.met, result.Met, "%s met", name)
				assert.Equal(t, want.hard, result.Hard, "%s hard", name)
			}
		})
	}
}

func TestOutcomeAndRuleJudge(t *testing.T) {
	met := schematypes.SituationReasonResult{Reason: schematypes.SituationRaiseReasonBreadth, Met: true}
	hard := schematypes.SituationReasonResult{Reason: schematypes.SituationRaiseReasonLinkedIncident, Met: true, Hard: true}
	unmet := schematypes.SituationReasonResult{Reason: schematypes.SituationRaiseReasonNovelty}
	cases := []struct {
		name         string
		reasons      Reasons
		wantOutcome  sitjudg.Outcome
		wantDecision sitjudg.Decision
		wantCited    []schematypes.SituationRaiseReason
	}{
		{
			name:         "nothing met holds",
			reasons:      Reasons{unmet},
			wantOutcome:  sitjudg.OutcomeNoReason,
			wantDecision: sitjudg.DecisionHold,
			wantCited:    []schematypes.SituationRaiseReason{},
		},
		{
			name:         "a met reason needs a decision and the rule judge raises",
			reasons:      Reasons{met, unmet},
			wantOutcome:  sitjudg.OutcomeNeedsDecision,
			wantDecision: sitjudg.DecisionRaise,
			wantCited:    []schematypes.SituationRaiseReason{met.Reason},
		},
		{
			name:         "a hard reason raises citing every met reason",
			reasons:      Reasons{hard, met, unmet},
			wantOutcome:  sitjudg.OutcomeRaise,
			wantDecision: sitjudg.DecisionRaise,
			wantCited:    []schematypes.SituationRaiseReason{hard.Reason, met.Reason},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantOutcome, tc.reasons.Outcome())
			judgment := tc.reasons.JudgeByRules()
			assert.Equal(t, tc.wantDecision, judgment.Decision)
			assert.Equal(t, tc.wantCited, judgment.Cited)
		})
	}
}

func TestFingerprint(t *testing.T) {
	incident := schematypes.SituationIncidentFacts{
		ID:            uuid.New(),
		Title:         "Checkout outage",
		ResponseState: "started",
	}
	base := schematypes.SituationFacts{
		AsOf:            evaluationNow,
		Signals:         []schematypes.SituationSignalFacts{activeAlert("s1", time.Minute), activeAlert("s2", time.Minute)},
		LinkedIncidents: []schematypes.SituationIncidentFacts{incident},
		Entities:        []schematypes.SituationEntityFacts{{ID: uuid.New(), DisplayName: "checkout", Matching: true}},
	}
	fingerprint := func(facts schematypes.SituationFacts) string {
		value, fingerprintErr := CheckReasons(facts).Fingerprint(facts)
		require.NoError(t, fingerprintErr)
		return value
	}
	baseline := fingerprint(base)

	cases := []struct {
		name    string
		edit    func(facts *schematypes.SituationFacts)
		changed bool
	}{
		{"a later capture time", func(facts *schematypes.SituationFacts) { facts.AsOf = facts.AsOf.Add(time.Hour) }, false},
		{"growing active time below a threshold", func(facts *schematypes.SituationFacts) {
			facts.Signals[0] = withAlert(facts.Signals[0], func(alert *schematypes.SituationAlertFacts) { alert.ActiveSeconds = 120 })
		}, false},
		{"other display refs", func(facts *schematypes.SituationFacts) { facts.Signals[0].Ref = "s9" }, false},
		{"a signal title", func(facts *schematypes.SituationFacts) { facts.Signals[0].Title = "Checkout errors" }, false},
		{"an alert description and definition", func(facts *schematypes.SituationFacts) {
			facts.Signals[0] = withAlert(facts.Signals[0], func(alert *schematypes.SituationAlertFacts) {
				alert.Description = "Errors are high."
				alert.Definition = "errors > 5"
			})
		}, false},
		{"an incident title", func(facts *schematypes.SituationFacts) { facts.LinkedIncidents[0].Title = "Renamed" }, false},
		{"an incident response state", func(facts *schematypes.SituationFacts) { facts.LinkedIncidents[0].ResponseState = "resolved" }, true},
		{"a finished signal", func(facts *schematypes.SituationFacts) { facts.Signals[0].FinishedAt = new(evaluationNow) }, true},
		{"a changed attention level", func(facts *schematypes.SituationFacts) {
			facts.Signals[1] = withAttention(facts.Signals[1], ssa.LevelWatchOnly)
		}, true},
		{"another linked incident", func(facts *schematypes.SituationFacts) {
			facts.LinkedIncidents = append(facts.LinkedIncidents, schematypes.SituationIncidentFacts{ID: uuid.New()})
		}, true},
		{"a renamed entity", func(facts *schematypes.SituationFacts) { facts.Entities[0].DisplayName = "checkout-v2" }, true},
		{"a new relationship", func(facts *schematypes.SituationFacts) {
			facts.Relationships = []schematypes.SituationRelationshipFacts{{SourceID: facts.Entities[0].ID, Predicate: "calls", TargetID: uuid.New()}}
		}, true},
		{"an earlier situation's outcome", func(facts *schematypes.SituationFacts) {
			facts.Earlier = []schematypes.SituationEarlierFacts{{SituationID: uuid.New(), WasRaised: true}}
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			facts := base
			facts.Signals = append([]schematypes.SituationSignalFacts{}, base.Signals...)
			facts.LinkedIncidents = append([]schematypes.SituationIncidentFacts{}, base.LinkedIncidents...)
			facts.Entities = append([]schematypes.SituationEntityFacts{}, base.Entities...)
			tc.edit(&facts)
			assert.Equal(t, tc.changed, fingerprint(facts) != baseline)
		})
	}
}
