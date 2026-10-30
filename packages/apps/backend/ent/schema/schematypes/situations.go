package schematypes

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

// SituationObservationKind is what a situation observation is of.
type SituationObservationKind string

const SituationObservationKindAlertEpisode SituationObservationKind = "alert_episode"

// SituationObservationRef names what one observation is of, with one field per observation kind mirroring
// the observation's reference columns. A valid reference sets exactly one field. It is comparable, so it keys
// maps.
type SituationObservationRef struct {
	AlertEpisodeID uuid.UUID `json:"alert_episode_id,omitzero"`
}

// Kind is the kind of the one field set, or empty for a reference that is not valid.
func (r SituationObservationRef) Kind() SituationObservationKind {
	if r.AlertEpisodeID != uuid.Nil {
		return SituationObservationKindAlertEpisode
	}
	return ""
}

// String is "<kind>:<id>", a stable key and order for the reference, or empty for one that is not valid.
func (r SituationObservationRef) String() string {
	switch kind := r.Kind(); kind {
	case SituationObservationKindAlertEpisode:
		return string(kind) + ":" + r.AlertEpisodeID.String()
	}
	return ""
}

// SignalSeverity is the normalized severity a signal source reports, ordered from lowest to highest. It is
// the Go type of the alert window and episode severity enums.
type SignalSeverity string

const (
	SignalSeverityUnknown  SignalSeverity = "unknown"
	SignalSeverityInfo     SignalSeverity = "info"
	SignalSeverityWarning  SignalSeverity = "warning"
	SignalSeverityCritical SignalSeverity = "critical"
)

var signalSeverityOrder = []SignalSeverity{
	SignalSeverityUnknown,
	SignalSeverityInfo,
	SignalSeverityWarning,
	SignalSeverityCritical,
}

// Values lists the severities for the Ent enum fields that use this type.
func (SignalSeverity) Values() []string {
	values := make([]string, 0, len(signalSeverityOrder))
	for _, severity := range signalSeverityOrder {
		values = append(values, string(severity))
	}
	return values
}

// Compare orders severities from lowest to highest, like cmp.Compare, for use with slices.MaxFunc and
// similar helpers.
func (s SignalSeverity) Compare(other SignalSeverity) int {
	return slices.Index(signalSeverityOrder, s) - slices.Index(signalSeverityOrder, other)
}

// SituationAlertFacts describe an observed alert episode from its stored state.
type SituationAlertFacts struct {
	Description string `json:"description,omitempty"`
	Definition  string `json:"definition,omitempty"`
	// Severity is the highest severity of the episode's windows.
	Severity SignalSeverity `json:"severity"`
	// Active is true while a window is firing, not while the episode only waits out its flap grace.
	Active bool `json:"active"`
	// ActiveGroupCount is the number of distinct grouping keys among the active windows.
	ActiveGroupCount int `json:"active_group_count"`
	// ActiveSeconds is the union of the windows' firing intervals, excluding gaps and grace.
	ActiveSeconds       int64 `json:"active_seconds"`
	InstanceCount       int   `json:"instance_count"`
	ActiveInstanceCount int   `json:"active_instance_count"`
	// TimedOutInstanceCount counts windows that ended by the definition's resolution timeout.
	TimedOutInstanceCount int `json:"timed_out_instance_count"`
	// FlapCount counts windows of a key that fired again after an earlier window ended other than by supersession.
	FlapCount           int      `json:"flap_count"`
	IdentityGroupLabels []string `json:"identity_group_labels,omitempty"`
	// ActiveInstances are some of the active windows, by instance key.
	ActiveInstances []SituationAlertInstanceFacts `json:"active_instances"`
	Baseline        SituationBaselineFacts        `json:"baseline"`
}

// SituationBaselineFacts describe how often an alert definition fired before an episode, and for how long.
type SituationBaselineFacts struct {
	// ObservedHistoryAgeDays is how long before the episode started the tenant's earliest episode started: an
	// age heuristic, not proof of uninterrupted collection.
	ObservedHistoryAgeDays int `json:"observed_history_age_days"`
	// HasSufficientHistory is true when the observed history age is long enough for baselines.
	HasSufficientHistory bool `json:"has_sufficient_history"`
	// Occurrences counts the definition's other episodes that started within the novelty window before this one.
	Occurrences int `json:"occurrences"`
	// MedianDurationSeconds is the median firing duration of those episodes that closed with every window
	// resolved by the source; absent without such a sample.
	MedianDurationSeconds *int64 `json:"median_duration_seconds,omitempty"`
}

// SituationAlertInstanceFacts describe one active alert window.
type SituationAlertInstanceFacts struct {
	InstanceKey string            `json:"instance_key"`
	GroupingKey string            `json:"grouping_key"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// SituationRaiseReason names one of the grounds for raising a candidate situation.
type SituationRaiseReason string

const (
	SituationRaiseReasonLinkedIncident SituationRaiseReason = "linked_incident"
	SituationRaiseReasonBreadth        SituationRaiseReason = "breadth"
	SituationRaiseReasonNovelty        SituationRaiseReason = "novelty"
	SituationRaiseReasonPersistence    SituationRaiseReason = "persistence"
	SituationRaiseReasonPastIncident   SituationRaiseReason = "past_incident"
)

// SituationReasonResult is one reason checked against a situation's facts.
type SituationReasonResult struct {
	Reason SituationRaiseReason `json:"reason"`
	Met    bool                 `json:"met"`
	Hard   bool                 `json:"hard"`
	// Detail is one plain sentence with the numbers, shown to people.
	Detail string `json:"detail"`
}

// SituationPriorOutcomeFacts count how other situations observing the same signal ended up.
type SituationPriorOutcomeFacts struct {
	Raised             int `json:"raised"`
	MutedNotNoteworthy int `json:"muted_not_noteworthy"`
	IncidentLinked     int `json:"incident_linked"`
}

// SituationObservationFacts describe one observation as of the facts' time.
type SituationObservationFacts struct {
	// Ref is "o1", "o2", … by attachment, then observation reference.
	Ref string `json:"ref"`
	SituationObservationRef
	Kind  SituationObservationKind `json:"kind"`
	Title string                   `json:"title"`
	// SourceID is the signal that produced it, such as an episode's alert definition.
	SourceID uuid.UUID `json:"source_id"`
	// Seeding is true when the observation may start a candidate and so counts toward raise reasons.
	Seeding bool `json:"seeding"`
	// Attention is the signal's attention level.
	Attention      string     `json:"attention"`
	AttentionSetAt *time.Time `json:"attention_set_at,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	// FinishedAt is absent while the observation is unfinished.
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	// PriorOutcomes describe other situations observing the same signal.
	PriorOutcomes *SituationPriorOutcomeFacts `json:"prior_outcomes,omitempty"`
	// EntityIDs are the runtime entities the observation is about, by ID. Only the model judge's facts have them.
	EntityIDs []uuid.UUID `json:"entity_ids,omitempty"`
	// Alert describes an alert episode.
	Alert *SituationAlertFacts `json:"alert,omitempty"`
}

// SituationIncidentFacts describe one linked incident.
type SituationIncidentFacts struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	// ResolvedAt is when the incident was resolved, if it has been.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// SituationEntityFacts describe one of a situation's runtime entities from its knowledge entity state.
type SituationEntityFacts struct {
	ID          uuid.UUID      `json:"id"`
	Category    string         `json:"category"`
	Kind        string         `json:"kind"`
	DisplayName string         `json:"display_name,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	// Matching is true when an observation that is not broad touches the entity.
	Matching bool `json:"matching"`
}

// SituationRelationshipFacts is a depends-on or adjacent relationship between two of a situation's entities.
type SituationRelationshipFacts struct {
	SourceID  uuid.UUID `json:"source_id"`
	Predicate string    `json:"predicate"`
	TargetID  uuid.UUID `json:"target_id"`
}

// SituationFacts are a snapshot of what Rezible knew about a situation at one processing time. Entities,
// Relationships and each observation's EntityIDs are context only the model judge reads; facts judged by
// the rules leave them out.
type SituationFacts struct {
	AsOf            time.Time                   `json:"as_of"`
	OpenedAt        time.Time                   `json:"opened_at"`
	Observations    []SituationObservationFacts `json:"observations"`
	LinkedIncidents []SituationIncidentFacts    `json:"linked_incidents"`
	// Entities are ordered matching first, then by ID.
	Entities []SituationEntityFacts `json:"entities,omitempty"`
	// Relationships are the depends-on and adjacent relationships among the entities.
	Relationships []SituationRelationshipFacts `json:"relationships,omitempty"`
}
