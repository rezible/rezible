package situations

import (
	"bytes"
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
	for _, signal := range facts.Signals {
		if signal.Seeding {
			rules.seeding = append(rules.seeding, signal)
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
	// seeding are the signals that count toward reasons.
	seeding []schematypes.SituationSignalFacts
}

// distinctSources counts the seeding signals' sources; a signal without a source is its own source.
func (r raiseRules) distinctSources() int {
	sources := mapset.NewSet[uuid.UUID]()
	for _, signal := range r.seeding {
		if signal.SourceEntityID != nil {
			sources.Add(*signal.SourceEntityID)
		} else {
			sources.Add(signal.EntityID)
		}
	}
	return sources.Cardinality()
}

// watchOnlyAlone is the watch-only gate: every seeding source is watch-only and there are fewer than two.
func (r raiseRules) watchOnlyAlone() bool {
	if len(r.seeding) == 0 {
		return false
	}
	for _, signal := range r.seeding {
		if signal.Attention != string(ssa.LevelWatchOnly) {
			return false
		}
	}
	return r.distinctSources() < 2
}

// defaultAlert is the alert facts of a default-attention seeding signal, or nil.
func (r raiseRules) defaultAlert(signal schematypes.SituationSignalFacts) *schematypes.SituationAlertFacts {
	if signal.Attention != string(ssa.LevelDefault) {
		return nil
	}
	return signal.Alert
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
	for _, signal := range r.seeding {
		if alert := r.defaultAlert(signal); alert != nil && alert.ActiveGroupCount > widestCount {
			widestRef, widestCount = signal.Ref, alert.ActiveGroupCount
		}
	}
	result := schematypes.SituationReasonResult{Reason: schematypes.SituationRaiseReasonBreadth}
	switch {
	case sources >= BreadthSources:
		result.Met = true
		result.Hard = sources >= HardBreadthSources
		result.Detail = fmt.Sprintf("Signals from %d different sources.", sources)
	case widestCount >= BreadthActiveCount:
		result.Met = true
		result.Detail = fmt.Sprintf("%s covers %d distinct active groups.", widestRef, widestCount)
	default:
		result.Detail = fmt.Sprintf("Signals from %s; %d needed, or one alert covering %d distinct active groups.",
			plural(sources, "source"), BreadthSources, BreadthActiveCount)
	}
	return result
}

func (r raiseRules) novelty() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonNovelty,
		Detail: "No active default-attention alert is rare in observed history.",
	}
	for _, signal := range r.seeding {
		alert := r.defaultAlert(signal)
		if alert == nil || !alert.Active || !alert.Baseline.HasSufficientHistory || alert.Baseline.Occurrences > NoveltyMaxOccurrences {
			continue
		}
		if active := time.Duration(alert.ActiveSeconds) * time.Second; active < NoveltyMinActive {
			result.Detail = fmt.Sprintf("%s is rare in observed history but has been active %s of %s.",
				signal.Ref, active, NoveltyMinActive)
			continue
		}
		result.Met = true
		result.Detail = fmt.Sprintf("%s is rare in observed history: its source fired %s in the last %d days.",
			signal.Ref, plural(alert.Baseline.Occurrences, "time"), NoveltyWindow/(24*time.Hour))
		return result
	}
	return result
}

func (r raiseRules) persistence() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonPersistence,
		Detail: "No default-attention seeding alert is active.",
	}
	for _, signal := range r.seeding {
		alert := r.defaultAlert(signal)
		if alert == nil || !alert.Active {
			continue
		}
		active := time.Duration(alert.ActiveSeconds) * time.Second
		threshold := persistenceThreshold(alert)
		if active >= threshold {
			result.Met = true
			result.Detail = fmt.Sprintf("%s has been active %s, at least %s.", signal.Ref, active, threshold)
			return result
		}
		result.Detail = fmt.Sprintf("%s has been active %s; it persists at %s.", signal.Ref, active, threshold)
	}
	return result
}

func (r raiseRules) pastIncident() schematypes.SituationReasonResult {
	result := schematypes.SituationReasonResult{
		Reason: schematypes.SituationRaiseReasonPastIncident,
		Detail: fmt.Sprintf("No other situation with a seeding signal's source led to an incident in the last %d days.",
			PastIncidentLookback/(24*time.Hour)),
	}
	for _, signal := range r.seeding {
		if signal.PriorOutcomes == nil || signal.PriorOutcomes.IncidentLinked == 0 {
			continue
		}
		result.Met = true
		result.Detail = fmt.Sprintf("%s with %s's source led to an incident.",
			plural(signal.PriorOutcomes.IncidentLinked, "other situation"), signal.Ref)
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
// their names and properties, relationships and earlier situations. It leaves out the capture time, display
// refs, prose (titles, alert descriptions and definitions), attention set times, and the continuously growing
// active time and history age, whose threshold crossings the reasons already carry. A judgment with the same
// fingerprint stands.
func (r Reasons) Fingerprint(facts schematypes.SituationFacts) (string, error) {
	inputs := decisionInputs{
		Signals:         make([]decisionSignal, 0, len(facts.Signals)),
		LinkedIncidents: make([]decisionIncident, 0, len(facts.LinkedIncidents)),
		Reasons:         make([]decisionReason, 0, len(r)),
		Entities:        facts.Entities,
		Relationships:   facts.Relationships,
		Earlier:         facts.Earlier,
	}
	for _, signal := range facts.Signals {
		selected := decisionSignal{
			EntityID:       signal.EntityID,
			Kind:           signal.Kind,
			SourceEntityID: signal.SourceEntityID,
			Seeding:        signal.Seeding,
			Attention:      signal.Attention,
			MatchKind:      signal.MatchKind,
			StartedAt:      signal.StartedAt,
			FinishedAt:     signal.FinishedAt,
			PriorOutcomes:  signal.PriorOutcomes,
			EntityIDs:      signal.EntityIDs,
		}
		if alert := signal.Alert; alert != nil {
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
		inputs.Signals = append(inputs.Signals, selected)
	}
	for _, incident := range facts.LinkedIncidents {
		inputs.LinkedIncidents = append(inputs.LinkedIncidents, decisionIncident{ID: incident.ID, ResponseState: incident.ResponseState})
	}
	slices.SortFunc(inputs.Signals, func(a, b decisionSignal) int {
		return bytes.Compare(a.EntityID[:], b.EntityID[:])
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
	Signals         []decisionSignal                         `json:"signals"`
	LinkedIncidents []decisionIncident                       `json:"linked_incidents"`
	Reasons         []decisionReason                         `json:"reasons"`
	Entities        []schematypes.SituationEntityFacts       `json:"entities,omitempty"`
	Relationships   []schematypes.SituationRelationshipFacts `json:"relationships,omitempty"`
	Earlier         []schematypes.SituationEarlierFacts      `json:"earlier,omitempty"`
}

type decisionSignal struct {
	EntityID       uuid.UUID                               `json:"entity_id"`
	Kind           string                                  `json:"kind"`
	SourceEntityID *uuid.UUID                              `json:"source_entity_id"`
	Seeding        bool                                    `json:"seeding"`
	Attention      string                                  `json:"attention"`
	MatchKind      string                                  `json:"match_kind"`
	StartedAt      time.Time                               `json:"started_at"`
	FinishedAt     *time.Time                              `json:"finished_at"`
	PriorOutcomes  *schematypes.SituationPriorOutcomeFacts `json:"prior_outcomes"`
	EntityIDs      []uuid.UUID                             `json:"entity_ids,omitempty"`
	Alert          *decisionAlert                          `json:"alert"`
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
	ID            uuid.UUID `json:"id"`
	ResponseState string    `json:"response_state"`
}

type decisionReason struct {
	Reason schematypes.SituationRaiseReason `json:"reason"`
	Met    bool                             `json:"met"`
	Hard   bool                             `json:"hard"`
}
