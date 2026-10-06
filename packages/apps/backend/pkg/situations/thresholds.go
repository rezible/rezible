package situations

import "time"

const (
	// EntityResolutionDepth is the most structure steps climbed when resolving a signal's entities to the
	// runtime level.
	EntityResolutionDepth = 3
	// BroadSignalEntityLimit is the most runtime entities a signal may touch and still make its situation
	// attract signals about them.
	BroadSignalEntityLimit = 4
	// HoldDefault is a hold's length when none is given.
	HoldDefault = 24 * time.Hour
	// SharedEntityWindow is how long after its last activity a situation with no active signal still
	// attracts signals about one of its matching entities.
	SharedEntityWindow = 6 * time.Hour
	// OneHopWindow is how long a quiet situation still attracts signals one relationship away. A signal
	// that finished longer ago than this is not placed.
	OneHopWindow = 30 * time.Minute
	// RecurrenceLookback is how long ago a situation of the same source may have closed to be linked as
	// a recurrence of a new candidate.
	RecurrenceLookback = 7 * 24 * time.Hour

	// SituationQuietPeriod is how long a situation waits after its last evidence finished or attached before
	// it closes.
	SituationQuietPeriod = 30 * time.Minute
	// CandidateMaxAge is how long after its creation a candidate with unfinished evidence expires.
	CandidateMaxAge = 24 * time.Hour
	// CollectionWindow is how long after its creation a candidate raises only for a hard reason.
	CollectionWindow = 90 * time.Second
	// BreadthSources is the number of distinct seeding sources that makes breadth true.
	BreadthSources = 2
	// BreadthActiveCount is the number of distinct active groups of one default-attention alert that makes
	// breadth true.
	BreadthActiveCount = 3
	// HardBreadthSources is the number of distinct seeding sources that raises without a judge.
	HardBreadthSources = 4
	// NoveltyWindow is the history window for novelty and duration baselines.
	NoveltyWindow = 30 * 24 * time.Hour
	// NoveltyMaxOccurrences is the most earlier episodes of a definition for its alert to be novel.
	NoveltyMaxOccurrences = 1
	// SufficientHistoryAge is how long the tenant's alert history must be for baselines to count.
	SufficientHistoryAge = 14 * 24 * time.Hour
	// NoveltyMinActive is how long a novel alert must have been active.
	NoveltyMinActive = 5 * time.Minute
	// PersistenceFactor multiplies a definition's median firing duration into its persistence threshold.
	PersistenceFactor = 3
	// PersistenceMin is the least active time for persistence.
	PersistenceMin = 10 * time.Minute
	// PersistenceNoBaseline is the active time for persistence without a baseline.
	PersistenceNoBaseline = 30 * time.Minute
	// PastIncidentLookback is how far back an incident-linked situation of a source counts.
	PastIncidentLookback = 90 * 24 * time.Hour
	// AutoInvestigationLimit is the most investigations automatic raises start per tenant within
	// AutoInvestigationWindow.
	AutoInvestigationLimit  = 3
	AutoInvestigationWindow = time.Hour

	// JudgeMaxSignals is the most signals sent to the model judge.
	JudgeMaxSignals = 20
	// JudgeMaxEntities is the most entities sent to the model judge.
	JudgeMaxEntities = 30
	// JudgeCallTimeout bounds one model judge call in real time.
	JudgeCallTimeout = 30 * time.Second
	// JudgeRetryAfter is how long after the judge was unavailable the same facts are judged again.
	JudgeRetryAfter = 5 * time.Minute
)

// IsBroad reports whether a signal touching this many runtime entities is broad.
func IsBroad(entityCount int) bool {
	return entityCount > BroadSignalEntityLimit
}
