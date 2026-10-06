package db

import (
	"bytes"
	"cmp"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/rezible/rezible/ent"
	ali "github.com/rezible/rezible/ent/alertinstance"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/situations"
)

// alertEpisodeState is one episode's windows, ordered by provider start and then ID, from which its
// effective state and signal facts are computed.
type alertEpisodeState []*ent.AlertInstance

func newAlertEpisodeState(windows []*ent.AlertInstance) alertEpisodeState {
	sorted := slices.Clone(windows)
	slices.SortFunc(sorted, func(a, b *ent.AlertInstance) int {
		return cmp.Or(a.FiredAt.Compare(b.FiredAt), bytes.Compare(a.ID[:], b.ID[:]))
	})
	return sorted
}

// alertWindowEnd is a window's effective end and why it ended.
type alertWindowEnd struct {
	at     time.Time
	reason ali.EndReason
}

func (e *alertWindowEnd) isStoredOn(window *ent.AlertInstance) bool {
	if e == nil || window.EndedAt == nil {
		return e == nil && window.EndedAt == nil && window.EndReason == nil
	}
	return e.at.Equal(*window.EndedAt) && window.EndReason != nil && e.reason == *window.EndReason
}

// alertEpisodeSettlement is an episode's effective state as of a time.
type alertEpisodeSettlement struct {
	// ends holds each window's effective end in window order; nil while the window is active.
	ends     []*alertWindowEnd
	closedAt *time.Time
	// nextDeadline is the earliest later time at which settling again changes the result.
	nextDeadline *time.Time
}

// settle computes the episode's effective state as of asOf from the definition's resolution timeout (0 never
// times out) and the flap grace. A window ends when resolved; otherwise when a later window of its
// instance key starts; otherwise at its timeout. The episode closes at its latest window end plus the
// grace, once every window has ended.
func (ws alertEpisodeState) settle(timeout, grace time.Duration, asOf time.Time) alertEpisodeSettlement {
	settlement := alertEpisodeSettlement{ends: make([]*alertWindowEnd, len(ws))}
	var deadlines []time.Time
	successors := make(map[string]time.Time)
	for i := len(ws) - 1; i >= 0; i-- {
		window := ws[i]
		successor, hasSuccessor := successors[window.InstanceKey]
		successors[window.InstanceKey] = window.FiredAt
		if window.ResolvedAt != nil {
			settlement.ends[i] = &alertWindowEnd{at: *window.ResolvedAt, reason: ali.EndReasonResolved}
			continue
		}
		if hasSuccessor && !asOf.Before(successor) {
			settlement.ends[i] = &alertWindowEnd{at: successor, reason: ali.EndReasonSuperseded}
			continue
		}
		if hasSuccessor {
			deadlines = append(deadlines, successor)
		}
		if timeout <= 0 {
			continue
		}
		timeoutAt := window.LastObservedAt.Add(timeout)
		if asOf.Before(timeoutAt) {
			deadlines = append(deadlines, timeoutAt)
		} else {
			settlement.ends[i] = &alertWindowEnd{at: timeoutAt, reason: ali.EndReasonTimeout}
		}
	}

	if closesAt := settlement.closesAt(grace); closesAt != nil {
		if asOf.Before(*closesAt) {
			deadlines = append(deadlines, *closesAt)
		} else {
			settlement.closedAt = closesAt
		}
	}
	if len(deadlines) > 0 {
		settlement.nextDeadline = new(slices.MinFunc(deadlines, time.Time.Compare))
	}
	return settlement
}

// closesAt is the latest window end plus the grace once every window has ended, or nil.
func (s alertEpisodeSettlement) closesAt(grace time.Duration) *time.Time {
	if len(s.ends) == 0 || slices.Contains(s.ends, nil) {
		return nil
	}
	lastEnd := slices.MaxFunc(s.ends, func(a, b *alertWindowEnd) int {
		return a.at.Compare(b.at)
	})
	return new(lastEnd.at.Add(grace))
}

// facts describes the windows' stored state as of asOf for an alert signal.
func (ws alertEpisodeState) facts(asOf time.Time) schematypes.SituationAlertFacts {
	facts := schematypes.SituationAlertFacts{
		InstanceCount:   len(ws),
		ActiveInstances: make([]schematypes.SituationAlertInstanceFacts, 0),
	}
	activeGroups := mapset.NewSet[string]()
	previousByKey := make(map[string]*ent.AlertInstance)
	for _, window := range ws {
		previous, seen := previousByKey[window.InstanceKey]
		if seen && previous.EndReason != nil && *previous.EndReason != ali.EndReasonSuperseded {
			facts.FlapCount++
		}
		previousByKey[window.InstanceKey] = window
		if window.EndReason != nil && *window.EndReason == ali.EndReasonTimeout {
			facts.TimedOutInstanceCount++
		}
		if window.EndedAt == nil {
			activeGroups.Add(window.GroupingKey)
			facts.ActiveInstances = append(facts.ActiveInstances, schematypes.SituationAlertInstanceFacts{
				InstanceKey: window.InstanceKey,
				GroupingKey: window.GroupingKey,
				Labels:      window.Labels,
			})
		}
	}
	slices.SortStableFunc(facts.ActiveInstances, func(a, b schematypes.SituationAlertInstanceFacts) int {
		return cmp.Compare(a.InstanceKey, b.InstanceKey)
	})

	facts.ActiveInstanceCount = len(facts.ActiveInstances)
	facts.Active = facts.ActiveInstanceCount > 0
	facts.ActiveGroupCount = activeGroups.Cardinality()
	facts.ActiveSeconds = int64(ws.firingDuration(asOf) / time.Second)
	facts.ActiveInstances = facts.ActiveInstances[:min(facts.ActiveInstanceCount, situations.AlertFactsMaxInstances)]
	return facts
}

// highestSeverity is the highest severity of the windows.
func (ws alertEpisodeState) highestSeverity() schematypes.SignalSeverity {
	highest := schematypes.SignalSeverityUnknown
	for _, window := range ws {
		if window.Severity.Compare(highest) > 0 {
			highest = window.Severity
		}
	}
	return highest
}

// firingDuration is the length of the union of the windows' firing intervals, capped at their stored
// ends and at asOf. Overlapping windows count once; gaps and the grace do not count.
func (ws alertEpisodeState) firingDuration(asOf time.Time) time.Duration {
	var total time.Duration
	var spanStart, spanEnd time.Time
	for _, window := range ws {
		end := asOf
		if window.EndedAt != nil && window.EndedAt.Before(end) {
			end = *window.EndedAt
		}
		if !end.After(window.FiredAt) {
			continue
		}
		if !spanEnd.IsZero() && !window.FiredAt.After(spanEnd) {
			if end.After(spanEnd) {
				spanEnd = end
			}
			continue
		}
		total += spanEnd.Sub(spanStart)
		spanStart, spanEnd = window.FiredAt, end
	}
	return total + spanEnd.Sub(spanStart)
}
