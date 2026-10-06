package situations

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	inc "github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// EvaluationInput is an unclosed situation's state as evaluation reads it, at one processing time.
type EvaluationInput struct {
	Now       time.Time
	OpenedAt  time.Time
	CreatedAt time.Time
	Raised    bool
	Muted     bool
	HoldUntil *time.Time
	Members   []Member
	Incidents []LinkedIncident
	// PriorOutcomes are by source entity; only judgment reads them.
	PriorOutcomes map[uuid.UUID]schematypes.SituationPriorOutcomeFacts
	// Latest is an unmuted candidate's latest judgment, if any.
	Latest *LatestJudgment
	// Context is set only when the model judge is about to be called.
	Context *JudgeContext
}

// LatestJudgment is what deciding again depends on from a situation's latest judgment.
type LatestJudgment struct {
	JudgedAt    time.Time
	Judge       string
	Fingerprint string
}

// JudgeContext is what only the model judge reads: the situation's entities, each signal's entities, the
// relationships among them, and the situations it recurs.
type JudgeContext struct {
	Entities []schematypes.SituationEntityFacts
	// SignalEntityIDs are each signal's runtime entities, by signal entity.
	SignalEntityIDs map[uuid.UUID][]uuid.UUID
	Relationships   []schematypes.SituationRelationshipFacts
	Earlier         []schematypes.SituationEarlierFacts
}

// Member is one of the situation's signals with its membership.
type Member struct {
	Signal     Signal
	AttachedAt time.Time
	MatchKind  sitsig.MatchKind
}

type LinkedIncident struct {
	ID            uuid.UUID
	Title         string
	ResponseState inc.ResponseState
	ResolvedAt    *time.Time
}

// CloseableAt is the earliest time the situation may close given its data, even if already past. It is nil
// while something without a deadline keeps it open: an unresolved incident, or unfinished evidence in a
// raised situation. A candidate's age counts from when watching began, not from an old provider start.
func (in *EvaluationInput) CloseableAt() *time.Time {
	var latest time.Time
	for _, incident := range in.Incidents {
		if incident.ResponseState != inc.ResponseStateResolved {
			return nil
		}
		if incident.ResolvedAt != nil {
			latest = later(latest, *incident.ResolvedAt)
		}
	}
	var closeableAt time.Time
	if in.finished() {
		for _, member := range in.Members {
			latest = later(later(latest, *member.Signal.FinishedAt), member.AttachedAt)
		}
		closeableAt = latest.Add(SituationQuietPeriod)
	} else if in.Raised {
		return nil
	} else {
		closeableAt = in.CreatedAt.Add(CandidateMaxAge)
	}
	if in.HoldUntil != nil {
		closeableAt = later(closeableAt, *in.HoldUntil)
	}
	return &closeableAt
}

func (in *EvaluationInput) finished() bool {
	return !slices.ContainsFunc(in.Members, func(member Member) bool {
		return member.Signal.FinishedAt == nil
	})
}

// CloseReason is why the situation closes automatically. A muted situation is dismissed, even a candidate
// reaching its maximum age.
func (in *EvaluationInput) CloseReason() sit.CloseReason {
	switch {
	case in.Muted:
		return sit.CloseReasonDismissed
	case !in.Raised:
		return sit.CloseReasonExpired
	default:
		return sit.CloseReasonStabilized
	}
}

// JudgmentStands reports whether the latest judgment decides facts with this fingerprint: it judged the same
// facts, and is not an unavailable judgment due to be retried.
func (in *EvaluationInput) JudgmentStands(fingerprint string) bool {
	if in.Latest == nil || in.Latest.Fingerprint != fingerprint {
		return false
	}
	retryAt := in.retryJudgeAt()
	return retryAt == nil || in.Now.Before(*retryAt)
}

// Record updates the input with a judgment of facts with this fingerprint, recorded at Now.
func (in *EvaluationInput) Record(judgment Judgment, fingerprint string) {
	in.Raised = in.Raised || judgment.Decision == sitjudg.DecisionRaise
	in.Latest = &LatestJudgment{JudgedAt: in.Now, Judge: judgment.Judge, Fingerprint: fingerprint}
}

// retryJudgeAt is when an unavailable latest judgment is due to be retried, or nil.
func (in *EvaluationInput) retryJudgeAt() *time.Time {
	if in.Latest == nil || in.Latest.Judge != JudgeUnavailable {
		return nil
	}
	return new(in.Latest.JudgedAt.Add(JudgeRetryAfter))
}

// Collecting reports whether the candidate is within its collection window, when only a hard reason raises.
func (in *EvaluationInput) Collecting() bool {
	return in.Now.Before(in.CreatedAt.Add(CollectionWindow))
}

// NextDeadline is the earliest future time at which the evaluation could change on its own: closure, the end
// of collection, an unmuted candidate's alert crossing its persistence or novelty active time, or the retry of
// its unavailable judge. Nil when there is none.
func (in *EvaluationInput) NextDeadline() *time.Time {
	var deadlines []time.Time
	if closeableAt := in.CloseableAt(); closeableAt != nil {
		deadlines = append(deadlines, *closeableAt)
	}
	if !in.Raised {
		deadlines = append(deadlines, in.CreatedAt.Add(CollectionWindow))
	}
	if !in.Raised && !in.Muted {
		if retryAt := in.retryJudgeAt(); retryAt != nil {
			deadlines = append(deadlines, *retryAt)
		}
		for _, member := range in.Members {
			signal := member.Signal
			alert := signal.Alert
			if !signal.Seeds() || signal.SignalAttention != ssa.LevelDefault || alert == nil || !alert.Active {
				continue
			}
			active := time.Duration(alert.ActiveSeconds) * time.Second
			for _, threshold := range []time.Duration{persistenceThreshold(alert), NoveltyMinActive} {
				// Active seconds are whole seconds; rounding up keeps repeated evaluations on one deadline.
				crossing := in.Now.Add(threshold - active)
				if rounded := crossing.Truncate(time.Second); !rounded.Equal(crossing) {
					crossing = rounded.Add(time.Second)
				}
				deadlines = append(deadlines, crossing)
			}
		}
	}
	var next *time.Time
	for _, deadline := range deadlines {
		if deadline.After(in.Now) && (next == nil || deadline.Before(*next)) {
			next = &deadline
		}
	}
	return next
}

// Facts describe the situation as of Now: its signals ordered by attachment then entity ID, and its linked
// incidents by ID. With a judge context, they also hold its entities, matching first then by ID, each
// signal's entities, the relationships among them and the situations it recurs.
func (in *EvaluationInput) Facts() schematypes.SituationFacts {
	members := slices.Clone(in.Members)
	slices.SortFunc(members, func(a, b Member) int {
		return cmp.Or(a.AttachedAt.Compare(b.AttachedAt), compareIDs(a.Signal.EntityID, b.Signal.EntityID))
	})
	facts := schematypes.SituationFacts{
		AsOf:            in.Now,
		OpenedAt:        in.OpenedAt,
		Signals:         make([]schematypes.SituationSignalFacts, 0, len(members)),
		LinkedIncidents: make([]schematypes.SituationIncidentFacts, 0, len(in.Incidents)),
	}
	for i, member := range members {
		signal := member.Signal
		signalFacts := schematypes.SituationSignalFacts{
			Ref:            fmt.Sprintf("s%d", i+1),
			EntityID:       signal.EntityID,
			Kind:           string(signal.Kind),
			Title:          signal.Title,
			SourceEntityID: signal.SourceEntityID,
			Seeding:        signal.Seeds(),
			Attention:      string(signal.SignalAttention),
			AttentionSetAt: signal.SignalAttentionSetAt,
			MatchKind:      string(member.MatchKind),
			StartedAt:      signal.StartedAt,
			FinishedAt:     signal.FinishedAt,
			Alert:          signal.Alert,
		}
		if signal.SourceEntityID != nil {
			signalFacts.PriorOutcomes = new(in.PriorOutcomes[*signal.SourceEntityID])
		}
		if in.Context != nil {
			signalFacts.EntityIDs = slices.SortedFunc(slices.Values(in.Context.SignalEntityIDs[signal.EntityID]), compareIDs)
		}
		facts.Signals = append(facts.Signals, signalFacts)
	}
	for _, incident := range in.Incidents {
		incidentFacts := schematypes.SituationIncidentFacts{
			ID:            incident.ID,
			Title:         incident.Title,
			ResponseState: string(incident.ResponseState),
		}
		facts.LinkedIncidents = append(facts.LinkedIncidents, incidentFacts)
	}
	slices.SortFunc(facts.LinkedIncidents, func(a, b schematypes.SituationIncidentFacts) int {
		return compareIDs(a.ID, b.ID)
	})
	if in.Context != nil {
		in.Context.addTo(&facts)
	}
	return facts
}

// addTo adds the context to the facts in a stable order, without duplicate relationships.
func (c JudgeContext) addTo(facts *schematypes.SituationFacts) {
	facts.Entities = slices.SortedFunc(slices.Values(c.Entities), func(a, b schematypes.SituationEntityFacts) int {
		if a.Matching != b.Matching {
			if a.Matching {
				return -1
			}
			return 1
		}
		return compareIDs(a.ID, b.ID)
	})
	facts.Relationships = slices.Compact(slices.SortedFunc(slices.Values(c.Relationships), func(a, b schematypes.SituationRelationshipFacts) int {
		return cmp.Or(compareIDs(a.SourceID, b.SourceID), cmp.Compare(a.Predicate, b.Predicate), compareIDs(a.TargetID, b.TargetID))
	}))
	facts.Earlier = slices.SortedFunc(slices.Values(c.Earlier), func(a, b schematypes.SituationEarlierFacts) int {
		return compareIDs(a.SituationID, b.SituationID)
	})
}

func compareIDs(a, b uuid.UUID) int {
	return bytes.Compare(a[:], b[:])
}

// persistenceThreshold is the active time at which a seeding alert persists: a multiple of its definition's
// median firing duration with a baseline, otherwise a fixed time.
func persistenceThreshold(alert *schematypes.SituationAlertFacts) time.Duration {
	median := alert.Baseline.MedianDurationSeconds
	if !alert.Baseline.HasSufficientHistory || median == nil {
		return PersistenceNoBaseline
	}
	return max(PersistenceFactor*time.Duration(*median)*time.Second, PersistenceMin)
}

func later(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}
