package situations

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// JudgeRules is the judge of decisions made by the reasons alone.
const JudgeRules = "rules"

// Reasons are the five raise reasons checked against one snapshot of facts, met or not, in presentation
// order.
type Reasons []schematypes.SituationReasonResult

// CheckReasons checks every raise reason against the facts. It reads nothing else, so the same facts always
// give the same reasons.
func CheckReasons(facts schematypes.SituationFacts) Reasons {
	rules := raiseRules{facts: facts}
	for _, observation := range facts.Observations {
		if observation.Seeding {
			rules.seeding = append(rules.seeding, observation)
		}
	}
	results := Reasons{
		rules.linkedIncident(),
		rules.breadth(),
		rules.novelty(),
		rules.persistence(),
		rules.pastIncident(),
	}
	if !rules.watchOnlyAlone() {
		return results
	}
	// Except for a linked incident, a watch-only source alone never raises automatically.
	for i := range results {
		if results[i].Reason == schematypes.SituationRaiseReasonLinkedIncident {
			continue
		}
		results[i].Met = false
		results[i].Hard = false
		results[i].Detail += " Not counted: a watch-only source needs another seeding source."
	}
	return results
}

// raiseRules are the raise reasons over one snapshot of facts.
type raiseRules struct {
	facts schematypes.SituationFacts
	// seeding are the observations that count toward reasons.
	seeding []schematypes.SituationObservationFacts
}

// distinctSources counts the seeding observations' sources.
func (r raiseRules) distinctSources() int {
	sources := mapset.NewSet[uuid.UUID]()
	for _, observation := range r.seeding {
		sources.Add(observation.SourceID)
	}
	return sources.Cardinality()
}

// watchOnlyAlone is the watch-only gate: every seeding source is watch-only and there are fewer than two.
func (r raiseRules) watchOnlyAlone() bool {
	if len(r.seeding) == 0 {
		return false
	}
	for _, observation := range r.seeding {
		if observation.Attention != string(ssa.LevelWatchOnly) {
			return false
		}
	}
	return r.distinctSources() < 2
}

// defaultAlert is the alert facts of a default-attention seeding observation, or nil.
func (r raiseRules) defaultAlert(observation schematypes.SituationObservationFacts) *schematypes.SituationAlertFacts {
	if observation.Attention != string(ssa.LevelDefault) {
		return nil
	}
	return observation.Alert
}

func (r raiseRules) linkedIncident() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonLinkedIncident,
		Detail: "No incident is linked.",
	}
	if count := len(r.facts.LinkedIncidents); count > 0 {
		result.Met = true
		result.Hard = true
		result.Detail = fmt.Sprintf("Linked to %s.", plural(count, "incident"))
	}
	return result
}

func (r raiseRules) breadth() schematypes.SituationReasonResult {
	sources := r.distinctSources()
	// The widest is the default-attention seeding alert with the most distinct active groups.
	widestRef, widestCount := "", 0
	for _, observation := range r.seeding {
		if alert := r.defaultAlert(observation); alert != nil && alert.ActiveGroupCount > widestCount {
			widestRef, widestCount = observation.Ref, alert.ActiveGroupCount
		}
	}
	result := schematypes.SituationReasonResult{Reason: schematypes.SituationRaiseReasonBreadth}
	switch {
	case sources >= BreadthSources:
		result.Met = true
		result.Hard = sources >= HardBreadthSources
		result.Detail = fmt.Sprintf("Observations from %d different sources.", sources)
	case widestCount >= BreadthActiveCount:
		result.Met = true
		result.Detail = fmt.Sprintf("%s covers %d distinct active groups.", widestRef, widestCount)
	default:
		result.Detail = fmt.Sprintf("Observations from %s; %d needed, or one alert covering %d distinct active groups.",
			plural(sources, "source"), BreadthSources, BreadthActiveCount)
	}
	return result
}

func (r raiseRules) novelty() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonNovelty,
		Detail: "No active default-attention alert is rare in observed history.",
	}
	for _, observation := range r.seeding {
		alert := r.defaultAlert(observation)
		if alert == nil || !alert.Active || !alert.Baseline.HasSufficientHistory || alert.Baseline.Occurrences > NoveltyMaxOccurrences {
			continue
		}
		if active := time.Duration(alert.ActiveSeconds) * time.Second; active < NoveltyMinActive {
			result.Detail = fmt.Sprintf("%s is rare in observed history but has been active %s of %s.",
				observation.Ref, active, NoveltyMinActive)
			continue
		}
		result.Met = true
		result.Detail = fmt.Sprintf("%s is rare in observed history: its source fired %s in the last %d days.",
			observation.Ref, plural(alert.Baseline.Occurrences, "time"), NoveltyWindow/(24*time.Hour))
		return result
	}
	return result
}

func (r raiseRules) persistence() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonPersistence,
		Detail: "No default-attention seeding alert is active.",
	}
	for _, observation := range r.seeding {
		alert := r.defaultAlert(observation)
		if alert == nil || !alert.Active {
			continue
		}
		active := time.Duration(alert.ActiveSeconds) * time.Second
		threshold := persistenceThreshold(alert)
		if active >= threshold {
			result.Met = true
			result.Detail = fmt.Sprintf("%s has been active %s, at least %s.", observation.Ref, active, threshold)
			return result
		}
		result.Detail = fmt.Sprintf("%s has been active %s; it persists at %s.", observation.Ref, active, threshold)
	}
	return result
}

func (r raiseRules) pastIncident() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonPastIncident,
		Detail: fmt.Sprintf("No other situation with a seeding observation's source led to an incident in the last %d days.",
			PastIncidentLookback/(24*time.Hour)),
	}
	for _, observation := range r.seeding {
		if observation.PriorOutcomes == nil || observation.PriorOutcomes.IncidentLinked == 0 {
			continue
		}
		result.Met = true
		result.Detail = fmt.Sprintf("%s with %s's source led to an incident.",
			plural(observation.PriorOutcomes.IncidentLinked, "other situation"), observation.Ref)
		return result
	}
	return result
}

func plural(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

// Outcome is what the reasons alone say: raise for a met hard reason, a decision for any other met reason,
// otherwise no reason.
func (r Reasons) Outcome() sitjudg.Outcome {
	outcome := sitjudg.OutcomeNoReason
	for _, reason := range r {
		if reason.Met && reason.Hard {
			return sitjudg.OutcomeRaise
		}
		if reason.Met {
			outcome = sitjudg.OutcomeNeedsDecision
		}
	}
	return outcome
}

// Judgment is a decision on a candidate, with the reasons it rests on and the judge that made it.
type Judgment struct {
	Decision    sitjudg.Decision
	Cited       []schematypes.SituationRaiseReason
	Explanation string
	Judge       string
}

// JudgeByRules raises whenever a reason is met, citing every met reason, and holds otherwise.
func (r Reasons) JudgeByRules() Judgment {
	judgment := Judgment{Decision: sitjudg.DecisionHold, Cited: []schematypes.SituationRaiseReason{}, Judge: JudgeRules}
	var details []string
	for _, reason := range r {
		if reason.Met {
			judgment.Cited = append(judgment.Cited, reason.Reason)
			details = append(details, reason.Detail)
		}
	}
	if len(judgment.Cited) == 0 {
		judgment.Explanation = "No reason to raise is met."
		return judgment
	}
	judgment.Decision = sitjudg.DecisionRaise
	judgment.Explanation = strings.Join(details, " ")
	return judgment
}

// Fingerprint identifies the decision the facts and reasons call for: a SHA-256 over an explicit projection
// of the inputs that can change it, including the model judge's context when the facts have it: entities with
// their names and properties, relationships. It leaves out the capture time, display
// refs, prose (titles, alert descriptions and definitions), attention set times, and the continuously growing
// active time and history age, whose threshold crossings the reasons already carry. A judgment with the same
// fingerprint stands.
func (r Reasons) Fingerprint(facts schematypes.SituationFacts) (string, error) {
	inputs := decisionInputs{
		Observations:    make([]decisionObservation, 0, len(facts.Observations)),
		LinkedIncidents: make([]decisionIncident, 0, len(facts.LinkedIncidents)),
		Reasons:         make([]decisionReason, 0, len(r)),
		Entities:        facts.Entities,
		Relationships:   facts.Relationships,
	}
	for _, observation := range facts.Observations {
		selected := decisionObservation{
			SituationObservationRef: observation.SituationObservationRef,
			SourceID:                observation.SourceID,
			Seeding:                 observation.Seeding,
			Attention:               observation.Attention,
			StartedAt:               observation.StartedAt,
			FinishedAt:              observation.FinishedAt,
			PriorOutcomes:           observation.PriorOutcomes,
			EntityIDs:               observation.EntityIDs,
		}
		if alert := observation.Alert; alert != nil {
			selected.Alert = &decisionAlert{
				Severity:              alert.Severity,
				Active:                alert.Active,
				ActiveGroupCount:      alert.ActiveGroupCount,
				InstanceCount:         alert.InstanceCount,
				ActiveInstanceCount:   alert.ActiveInstanceCount,
				TimedOutInstanceCount: alert.TimedOutInstanceCount,
				FlapCount:             alert.FlapCount,
				IdentityGroupLabels:   alert.IdentityGroupLabels,
				ActiveInstances:       alert.ActiveInstances,
				HasSufficientHistory:  alert.Baseline.HasSufficientHistory,
				Occurrences:           alert.Baseline.Occurrences,
				MedianDurationSeconds: alert.Baseline.MedianDurationSeconds,
			}
		}
		inputs.Observations = append(inputs.Observations, selected)
	}
	for _, incident := range facts.LinkedIncidents {
		inputs.LinkedIncidents = append(inputs.LinkedIncidents, decisionIncident{ID: incident.ID, ResolvedAt: incident.ResolvedAt})
	}
	slices.SortFunc(inputs.Observations, func(a, b decisionObservation) int {
		return cmp.Compare(a.String(), b.String())
	})
	for _, reason := range r {
		inputs.Reasons = append(inputs.Reasons, decisionReason{Reason: reason.Reason, Met: reason.Met, Hard: reason.Hard})
	}
	encoded, encodeErr := json.Marshal(inputs)
	if encodeErr != nil {
		return "", fmt.Errorf("encode situation decision inputs: %w", encodeErr)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

type decisionInputs struct {
	Observations    []decisionObservation                    `json:"observations"`
	LinkedIncidents []decisionIncident                       `json:"linked_incidents"`
	Reasons         []decisionReason                         `json:"reasons"`
	Entities        []schematypes.SituationEntityFacts       `json:"entities,omitempty"`
	Relationships   []schematypes.SituationRelationshipFacts `json:"relationships,omitempty"`
}

type decisionObservation struct {
	schematypes.SituationObservationRef
	SourceID      uuid.UUID                               `json:"source_id"`
	Seeding       bool                                    `json:"seeding"`
	Attention     string                                  `json:"attention"`
	StartedAt     time.Time                               `json:"started_at"`
	FinishedAt    *time.Time                              `json:"finished_at"`
	PriorOutcomes *schematypes.SituationPriorOutcomeFacts `json:"prior_outcomes"`
	EntityIDs     []uuid.UUID                             `json:"entity_ids,omitempty"`
	Alert         *decisionAlert                          `json:"alert"`
}

type decisionAlert struct {
	Severity              schematypes.SignalSeverity                `json:"severity"`
	Active                bool                                      `json:"active"`
	ActiveGroupCount      int                                       `json:"active_group_count"`
	InstanceCount         int                                       `json:"instance_count"`
	ActiveInstanceCount   int                                       `json:"active_instance_count"`
	TimedOutInstanceCount int                                       `json:"timed_out_instance_count"`
	FlapCount             int                                       `json:"flap_count"`
	IdentityGroupLabels   []string                                  `json:"identity_group_labels"`
	ActiveInstances       []schematypes.SituationAlertInstanceFacts `json:"active_instances"`
	HasSufficientHistory  bool                                      `json:"has_sufficient_history"`
	Occurrences           int                                       `json:"occurrences"`
	MedianDurationSeconds *int64                                    `json:"median_duration_seconds"`
}

type decisionIncident struct {
	ID         uuid.UUID  `json:"id"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

type decisionReason struct {
	Reason schematypes.SituationRaiseReason `json:"reason"`
	Met    bool                             `json:"met"`
	Hard   bool                             `json:"hard"`
}
