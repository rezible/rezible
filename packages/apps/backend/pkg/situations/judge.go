package situations

import (
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
)

const (
	// JudgeModel is the judge of decisions made by the judge_situation_candidate model workflow.
	JudgeModel = "llm:judge_situation_candidate"
	// JudgeModelRejected is the judge of holds recorded because the model's answer was malformed or invalid,
	// so they are never presented as the model's own assessment.
	JudgeModelRejected = JudgeModel + ":rejected"
	// JudgeUnavailable is the judge of holds recorded because the model failed or timed out. Nothing was
	// decided on their facts, and the same facts are judged again after JudgeRetryAfter.
	JudgeUnavailable = "unavailable"
)

// ErrMalformedAnswer describes a model answer that did not match the expected form.
var ErrMalformedAnswer = errors.New("the answer did not match the expected form")

// ValidateAnswer checks the model judge's answer: a decision of hold or raise, a non-empty explanation, and
// cited reasons that are known, distinct and met, with at least one for a raise.
func (r Reasons) ValidateAnswer(answer Judgment) error {
	if answer.Decision != sitjudg.DecisionHold && answer.Decision != sitjudg.DecisionRaise {
		return fmt.Errorf("unknown decision %q", answer.Decision)
	}
	if strings.TrimSpace(answer.Explanation) == "" {
		return errors.New("no explanation")
	}
	cited := mapset.NewSet[schematypes.SituationRaiseReason]()
	for _, reason := range answer.Cited {
		index := r.index(reason)
		switch {
		case index < 0:
			return fmt.Errorf("it cites unknown reason %q", reason)
		case !cited.Add(reason):
			return fmt.Errorf("it cites %q twice", reason)
		case !r[index].Met:
			return fmt.Errorf("it cites %q, which is not met", reason)
		}
	}
	if answer.Decision == sitjudg.DecisionRaise && cited.IsEmpty() {
		return errors.New("a raise cites no reason")
	}
	return nil
}

func (r Reasons) index(reason schematypes.SituationRaiseReason) int {
	for i, result := range r {
		if result.Reason == reason {
			return i
		}
	}
	return -1
}

// JudgeByModel records the model judge's answer when it is valid, and a rejected hold otherwise.
func (r Reasons) JudgeByModel(answer Judgment) Judgment {
	if invalidErr := r.ValidateAnswer(answer); invalidErr != nil {
		return RejectedJudgment(invalidErr)
	}
	answer.Judge = JudgeModel
	answer.Explanation = strings.TrimSpace(answer.Explanation)
	if answer.Cited == nil {
		answer.Cited = []schematypes.SituationRaiseReason{}
	}
	return answer
}

// RejectedJudgment holds because the model judge's answer had the problem.
func RejectedJudgment(problem error) Judgment {
	return Judgment{
		Decision:    sitjudg.DecisionHold,
		Cited:       []schematypes.SituationRaiseReason{},
		Explanation: fmt.Sprintf("Held: the model judge's answer was rejected (%s).", problem),
		Judge:       JudgeModelRejected,
	}
}

// UnavailableJudgment holds because the model judge failed or timed out.
func UnavailableJudgment() Judgment {
	return Judgment{
		Decision:    sitjudg.DecisionHold,
		Cited:       []schematypes.SituationRaiseReason{},
		Explanation: "Held: the model judge was unavailable, so nothing was decided on these facts. They will be judged again.",
		Judge:       JudgeUnavailable,
	}
}

// JudgeFacts cuts the facts to the model judge's limits and names every list it cut: it keeps the
// JudgeMaxSignals most recently attached signals and JudgeMaxEntities entities, matching first, then by ID,
// and restricts relationships and signal entity references to the kept entities. It also names each alert
// whose facts sample its active instances. Lists are never nil. Reasons and stored judgments use the full
// facts.
func JudgeFacts(facts schematypes.SituationFacts) (schematypes.SituationFacts, []string) {
	truncated := []string{}
	signals := facts.Signals
	if count := len(signals); count > JudgeMaxSignals {
		truncated = append(truncated, fmt.Sprintf("signals: kept the %d most recently attached of %d", JudgeMaxSignals, count))
		signals = signals[count-JudgeMaxSignals:]
	}
	entities := facts.Entities
	if count := len(entities); count > JudgeMaxEntities {
		truncated = append(truncated, fmt.Sprintf("entities: kept %d of %d, matching entities first", JudgeMaxEntities, count))
		entities = entities[:JudgeMaxEntities]
	}
	kept := mapset.NewSet[uuid.UUID]()
	for _, entity := range entities {
		kept.Add(entity.ID)
	}

	relationships := make([]schematypes.SituationRelationshipFacts, 0, len(facts.Relationships))
	for _, relationship := range facts.Relationships {
		if kept.Contains(relationship.SourceID) && kept.Contains(relationship.TargetID) {
			relationships = append(relationships, relationship)
		}
	}
	if len(relationships) < len(facts.Relationships) {
		truncated = append(truncated, fmt.Sprintf("relationships: kept the %d of %d between kept entities", len(relationships), len(facts.Relationships)))
	}

	restricted := false
	var sampled []string
	keptSignals := make([]schematypes.SituationSignalFacts, 0, len(signals))
	for _, signal := range signals {
		if alert := signal.Alert; alert != nil && len(alert.ActiveInstances) < alert.ActiveInstanceCount {
			sampled = append(sampled, fmt.Sprintf("%s alert active_instances: kept %d of %d, by instance key",
				signal.Ref, len(alert.ActiveInstances), alert.ActiveInstanceCount))
		}
		entityIDs := make([]uuid.UUID, 0, len(signal.EntityIDs))
		for _, entityID := range signal.EntityIDs {
			if kept.Contains(entityID) {
				entityIDs = append(entityIDs, entityID)
			}
		}
		restricted = restricted || len(entityIDs) < len(signal.EntityIDs)
		signal.EntityIDs = entityIDs
		keptSignals = append(keptSignals, signal)
	}
	if restricted {
		truncated = append(truncated, "signal entity_ids: restricted to kept entities")
	}
	truncated = append(truncated, sampled...)

	facts.Signals = keptSignals
	facts.Entities = append([]schematypes.SituationEntityFacts{}, entities...)
	facts.Relationships = relationships
	facts.LinkedIncidents = append([]schematypes.SituationIncidentFacts{}, facts.LinkedIncidents...)
	facts.Earlier = append([]schematypes.SituationEarlierFacts{}, facts.Earlier...)
	return facts, truncated
}
