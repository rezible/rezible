package situations

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
)

func TestValidateAnswer(t *testing.T) {
	// Breadth is met; persistence is not.
	reasons := CheckReasons(schematypes.SituationFacts{
		Signals: []schematypes.SituationSignalFacts{activeAlert("s1", time.Minute), activeAlert("s2", time.Minute)},
	})
	breadth := schematypes.SituationRaiseReasonBreadth
	cases := []struct {
		name   string
		answer Judgment
		valid  bool
	}{
		{"a raise citing a met reason", Judgment{Decision: sitjudg.DecisionRaise, Cited: []schematypes.SituationRaiseReason{breadth}, Explanation: "Two sources."}, true},
		{"a hold citing nothing", Judgment{Decision: sitjudg.DecisionHold, Explanation: "Weak evidence."}, true},
		{"an unknown decision", Judgment{Decision: "escalate", Explanation: "Now."}, false},
		{"no explanation", Judgment{Decision: sitjudg.DecisionHold, Explanation: "  "}, false},
		{"an unknown reason", Judgment{Decision: sitjudg.DecisionRaise, Cited: []schematypes.SituationRaiseReason{"recent_change"}, Explanation: "A deploy."}, false},
		{"a reason cited twice", Judgment{Decision: sitjudg.DecisionRaise, Cited: []schematypes.SituationRaiseReason{breadth, breadth}, Explanation: "Two sources."}, false},
		{"an unmet reason", Judgment{Decision: sitjudg.DecisionRaise, Cited: []schematypes.SituationRaiseReason{schematypes.SituationRaiseReasonPersistence}, Explanation: "Long."}, false},
		{"a raise citing nothing", Judgment{Decision: sitjudg.DecisionRaise, Explanation: "Looks bad."}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.valid, reasons.ValidateAnswer(tc.answer) == nil)
			judged := reasons.JudgeByModel(tc.answer)
			if tc.valid {
				assert.Equal(t, JudgeModel, judged.Judge)
				assert.Equal(t, tc.answer.Decision, judged.Decision)
			} else {
				assert.Equal(t, JudgeModelRejected, judged.Judge)
				assert.Equal(t, sitjudg.DecisionHold, judged.Decision)
			}
		})
	}
}

// judgeFacts are facts with the given numbers of signals and entities, each signal about every entity, and a
// relationship between each consecutive pair of entities. Entities alternate matching, ordered as Facts
// orders them.
func judgeFacts(signals, entities int) schematypes.SituationFacts {
	judgeContext := JudgeContext{SignalEntityIDs: make(map[uuid.UUID][]uuid.UUID)}
	var entityIDs []uuid.UUID
	for i := range entities {
		id := uuid.New()
		entityIDs = append(entityIDs, id)
		judgeContext.Entities = append(judgeContext.Entities, schematypes.SituationEntityFacts{ID: id, Matching: i%2 == 0})
		if i > 0 {
			relationship := schematypes.SituationRelationshipFacts{SourceID: entityIDs[i-1], Predicate: "calls", TargetID: id}
			judgeContext.Relationships = append(judgeContext.Relationships, relationship)
		}
	}
	input := EvaluationInput{Now: evaluationNow, Context: &judgeContext}
	for i := range signals {
		attached := member(i, nil, 1)
		judgeContext.SignalEntityIDs[attached.Signal.EntityID] = entityIDs
		input.Members = append(input.Members, attached)
	}
	return input.Facts()
}

func TestJudgeFacts(t *testing.T) {
	t.Run("within the limits nothing is cut", func(t *testing.T) {
		facts := judgeFacts(2, 3)
		kept, truncated := JudgeFacts(facts)
		assert.Equal(t, facts.Signals, kept.Signals)
		assert.Equal(t, facts.Entities, kept.Entities)
		assert.Equal(t, facts.Relationships, kept.Relationships)
		assert.NotNil(t, truncated)
		assert.Empty(t, truncated)
		assert.NotNil(t, kept.LinkedIncidents, "lists are never nil")
		assert.NotNil(t, kept.Earlier, "lists are never nil")
	})

	t.Run("past the limits it keeps the latest signals and matching entities, and says so", func(t *testing.T) {
		facts := judgeFacts(JudgeMaxSignals+2, JudgeMaxEntities+4)
		kept, truncated := JudgeFacts(facts)
		require.Len(t, kept.Signals, JudgeMaxSignals)
		assert.Equal(t, facts.Signals[2].Ref, kept.Signals[0].Ref, "the most recently attached are kept")
		require.Len(t, kept.Entities, JudgeMaxEntities)
		assert.Equal(t, facts.Entities[:JudgeMaxEntities], kept.Entities, "matching entities first, then by ID")
		keptIDs := make(map[uuid.UUID]bool)
		for _, entity := range kept.Entities {
			keptIDs[entity.ID] = true
		}
		for _, relationship := range kept.Relationships {
			assert.True(t, keptIDs[relationship.SourceID] && keptIDs[relationship.TargetID])
		}
		assert.Less(t, len(kept.Relationships), len(facts.Relationships))
		assert.Len(t, kept.Signals[0].EntityIDs, JudgeMaxEntities)
		assert.Len(t, truncated, 4, "signals, entities, relationships and signal entity references: %v", truncated)

		again, againTruncated := JudgeFacts(facts)
		assert.Equal(t, kept, again, "truncation is deterministic")
		assert.Equal(t, truncated, againTruncated)
	})

	t.Run("a sampled alert is declared", func(t *testing.T) {
		facts := judgeFacts(1, 1)
		facts.Signals[0].Alert.ActiveInstanceCount = AlertFactsMaxInstances + 1
		_, truncated := JudgeFacts(facts)
		assert.Equal(t, []string{fmt.Sprintf("s1 alert active_instances: kept 0 of %d, by instance key", AlertFactsMaxInstances+1)}, truncated)
	})
}
