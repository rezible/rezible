package db

import (
	"cmp"
	"context"
	"fmt"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/pkg/situations"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type AlertServiceSuite struct {
	test.Suite
}

func TestAlertServiceSuite(t *testing.T) {
	suite.Run(t, &AlertServiceSuite{Suite: test.NewSuite()})
}

var alertTestStart = time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

type alertServiceFixture struct {
	clock       *test.Clock
	knowledge   *KnowledgeGraphIngestionService
	alerts      *AlertService
	settlements []jobs.SettleAlertEpisode
	// notified are the signals situations were told changed, in order.
	notified []uuid.UUID
}

func (s *AlertServiceSuite) newFixture(tdb rez.Database) *alertServiceFixture {
	knowledge, knowledgeErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)
	h := &alertServiceFixture{clock: test.NewClock(alertTestStart), knowledge: knowledge}
	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
			switch args := args.(type) {
			case jobs.SettleAlertEpisode:
				s.True(args.DueAt.Equal(opts.ScheduledAt), "a settlement is scheduled at its due time")
				h.settlements = append(h.settlements, args)
			case jobs.ProcessSituationSignal:
				h.notified = append(h.notified, args.SignalEntityID)
			}
			return &rivertype.JobInsertResult{Job: &rivertype.JobRow{}}, nil
		}).
		Maybe()
	alerts, alertsErr := NewAlertService(rez.DefaultConfig().Alerts, h.clock, tdb, jobService, knowledge, NewSituationSignalService(jobService))
	s.Require().NoError(alertsErr)
	h.alerts = alerts
	return h
}

// testAlert is one notification of the test definition, timed in minutes from alertTestStart.
type testAlert struct {
	definition string
	at         int
	start      int
	resolved   bool
	key        string
	labels     map[string]string
	severity   schematypes.SignalSeverity
	timeout    *int
	title      string
	ref        string
	groupBy    []string
}

// record delivers the notification as the alert projector does: definition evidence through knowledge
// ingestion, then recording.
func (s *AlertServiceSuite) record(ctx context.Context, tdb rez.Database, h *alertServiceFixture, n testAlert) *ent.AlertDefinition {
	minute := func(m int) time.Time { return alertTestStart.Add(time.Duration(m) * time.Minute) }
	definitionRef := rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "alert-tests", ResourceRef: cmp.Or(n.definition, "checkout-errors")}
	createEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider(definitionRef.Provider).
		SetProviderNamespace(definitionRef.ProviderNamespace).
		SetProviderResourceRef(definitionRef.ResourceRef).
		SetProviderEventSource("alerts").
		SetProviderEventRef(cmp.Or(n.ref, uuid.NewString())).
		SetKind(projections.KindAlertInstance).
		SetOccurredAt(minute(n.at)).
		SetReceivedAt(minute(n.at)).
		SetAttributes([]byte(`{}`))
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)
	definitionEntityRef := rez.KnowledgeEntityRef{Category: kne.CategorySignal, Kind: "alert", ProviderResourceRef: definitionRef}
	definitionEvidence := rez.KnowledgeEvidenceRef{
		Kind:        "observed",
		Assertion:   "alert_definition_observed",
		EffectiveAt: event.OccurredAt,
		Subject:     rez.KnowledgeSubjectRef{Entity: &definitionEntityRef},
	}
	s.Require().NoError(h.knowledge.IngestEvidence(ctx, event, definitionEvidence))
	queryAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderResourceRef(definitionRef.ResourceRef))
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)

	params := rez.RecordAlertInstanceParams{
		Event: event,
		Definition: rez.AlertDefinitionValues{
			KnowledgeEntityID:        *alias.EntityID,
			Title:                    cmp.Or(n.title, "Checkout errors"),
			ResolutionTimeoutSeconds: n.timeout,
			IdentityGroupLabels:      n.groupBy,
		},
		Instance: rez.AlertInstanceValues{
			InstanceID: n.key,
			Labels:     n.labels,
			Severity:   n.severity,
			Firing:     !n.resolved,
			StartedAt:  minute(n.start),
		},
	}
	recorded, recordErr := h.alerts.RecordAlertInstance(ctx, params)
	s.Require().NoError(recordErr)
	return recorded
}

// episodes loads the tenant's episodes by start, with windows by start and their event links.
func (s *AlertServiceSuite) episodes(ctx context.Context, tdb rez.Database) []*ent.AlertEpisode {
	queryEpisodes := tdb.Client(ctx).AlertEpisode.Query().
		WithInstances(func(query *ent.AlertInstanceQuery) {
			query.Order(ali.ByFiredAt(), ali.ByInstanceKey()).WithEvents()
		}).
		Order(ale.ByStartedAt())
	episodes, queryErr := queryEpisodes.All(ctx)
	s.Require().NoError(queryErr)
	return episodes
}

func (s *AlertServiceSuite) TestRemindersAndResolvesOfOneWindow() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.record(ctx, tdb, h, testAlert{key: "a", labels: map[string]string{"v": "1"}, severity: schematypes.SignalSeverityWarning})
	s.record(ctx, tdb, h, testAlert{at: 2, key: "a", labels: map[string]string{"v": "2"}, severity: schematypes.SignalSeverityCritical})
	s.record(ctx, tdb, h, testAlert{at: 1, key: "a", labels: map[string]string{"v": "late"}, severity: schematypes.SignalSeverityInfo})
	s.record(ctx, tdb, h, testAlert{at: 9, key: "a", labels: map[string]string{"v": "2"}, resolved: true})

	episodes := s.episodes(ctx, tdb)
	s.Require().Len(episodes, 1)
	s.Require().Len(episodes[0].Edges.Instances, 1)
	window := episodes[0].Edges.Instances[0]
	s.Len(window.Edges.Events, 4)
	s.True(alertTestStart.Add(2 * time.Minute).Equal(window.LastObservedAt))
	s.True(alertTestStart.Add(9 * time.Minute).Equal(*window.ResolvedAt))
	s.True(window.ResolvedAt.Equal(*window.EndedAt))
	s.Equal(ali.EndReasonResolved, *window.EndReason)
	s.Equal(schematypes.SignalSeverityCritical, window.Severity, "peak severity")
	s.Equal("2", window.Labels["v"], "labels of the newest notification")
}

func (s *AlertServiceSuite) TestInstanceKeysAndGroupingKeys() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	podA := map[string]string{"service": "checkout", "pod": "a"}
	podB := map[string]string{"service": "checkout", "pod": "b"}
	s.record(ctx, tdb, h, testAlert{labels: podA, groupBy: []string{"service"}, severity: schematypes.SignalSeverityCritical})
	s.record(ctx, tdb, h, testAlert{at: 1, start: 1, labels: podB, groupBy: []string{"pod"}, severity: schematypes.SignalSeverityWarning})
	s.record(ctx, tdb, h, testAlert{at: 2, labels: podA, resolved: true})

	episodes := s.episodes(ctx, tdb)
	s.Require().Len(episodes, 1)
	s.Equal([]string{"service"}, episodes[0].IdentityGroupLabels, "a source proposal fills only an empty value")
	windows := episodes[0].Edges.Instances
	s.Require().Len(windows, 2)
	s.NotEqual(windows[0].InstanceKey, windows[1].InstanceKey)
	s.Equal(schematypes.SignalSeverityWarning, windows[1].Severity, "each window keeps its own peak")
	s.Equal(schematypes.SignalSeverityCritical, episodes[0].HighestSeverity, "the episode stores the highest across its windows")
	s.Equal(windows[0].GroupingKey, windows[1].GroupingKey)
	s.NotNil(windows[0].EndedAt, "resolving one instance does not resolve the other")
	s.Nil(windows[1].EndedAt)
}

func (s *AlertServiceSuite) TestRefireWithinAndAfterTheGrace() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.record(ctx, tdb, h, testAlert{key: "a"})
	s.record(ctx, tdb, h, testAlert{at: 10, key: "a", resolved: true})
	s.record(ctx, tdb, h, testAlert{at: 14, start: 14, key: "a"})
	s.record(ctx, tdb, h, testAlert{at: 15, start: 14, key: "a", resolved: true})
	s.record(ctx, tdb, h, testAlert{at: 21, start: 21, key: "a"})

	episodes := s.episodes(ctx, tdb)
	s.Require().Len(episodes, 2)
	s.Len(episodes[0].Edges.Instances, 2, "a refire within the grace joins the episode")
	s.True(alertTestStart.Add(21*time.Minute).Equal(episodes[1].StartedAt), "a refire after the grace opens an episode")
}

func (s *AlertServiceSuite) TestTimeoutResumptionAndSupersession() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	timeout := new(600)

	h.clock.Set(alertTestStart.Add(11 * time.Minute))
	s.record(ctx, tdb, h, testAlert{key: "a", timeout: timeout})
	window := s.episodes(ctx, tdb)[0].Edges.Instances[0]
	s.Equal(ali.EndReasonTimeout, *window.EndReason, "a notification delivered after its timeout ends its window")
	s.True(alertTestStart.Add(10 * time.Minute).Equal(*window.EndedAt))

	s.record(ctx, tdb, h, testAlert{at: 5, key: "a"})
	s.Nil(s.episodes(ctx, tdb)[0].Edges.Instances[0].EndedAt, "a later firing resumes a timed-out window")

	s.record(ctx, tdb, h, testAlert{at: 11, start: 11, key: "a"})
	windows := s.episodes(ctx, tdb)[0].Edges.Instances
	s.Require().Len(windows, 2)
	s.Equal(ali.EndReasonSuperseded, *windows[0].EndReason)
	s.True(alertTestStart.Add(11 * time.Minute).Equal(*windows[0].EndedAt))
	s.Nil(windows[1].EndedAt)
}

func (s *AlertServiceSuite) TestSettlementAtDeadlines() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.record(ctx, tdb, h, testAlert{key: "a", timeout: new(600)})
	s.Require().Len(h.settlements, 1)
	s.True(alertTestStart.Add(10*time.Minute).Equal(h.settlements[0].DueAt), "the timeout is the next deadline")

	h.clock.Set(h.settlements[0].DueAt)
	s.Require().NoError(h.alerts.settleAlertEpisode(ctx, h.settlements[0]))
	episode := s.episodes(ctx, tdb)[0]
	s.Equal(ali.EndReasonTimeout, *episode.Edges.Instances[0].EndReason)
	s.Nil(episode.ClosedAt, "the episode waits out the grace")
	s.Require().Len(h.settlements, 2)

	h.clock.Set(h.settlements[1].DueAt)
	s.Require().NoError(h.alerts.settleAlertEpisode(ctx, h.settlements[1]))
	episode = s.episodes(ctx, tdb)[0]
	s.Require().NotNil(episode.ClosedAt)
	s.True(alertTestStart.Add(15 * time.Minute).Equal(*episode.ClosedAt))
	s.Len(h.settlements, 2, "a closed episode has no deadline")
}

func (s *AlertServiceSuite) TestSignalChangeNotifications() {
	_, tdb := s.SetupTestDatabase()
	ctx, _ := s.NewIdentity(tdb, "alert owner")
	h := s.newFixture(tdb)

	definition := s.record(ctx, tdb, h, testAlert{key: "a", timeout: new(600)})
	signal := *s.episodes(ctx, tdb)[0].KnowledgeEntityID
	s.Equal([]uuid.UUID{signal}, h.notified, "recording notifies")

	h.clock.Set(h.settlements[0].DueAt)
	s.Require().NoError(h.alerts.settleAlertEpisode(ctx, h.settlements[0]))
	s.Len(h.notified, 2, "a settlement that ends a window notifies")
	s.Require().NoError(h.alerts.settleAlertEpisode(ctx, h.settlements[0]))
	s.Len(h.notified, 2, "a settlement that changes nothing does not notify")

	_, watchErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelWatchOnly)
	s.Require().NoError(watchErr)
	s.Len(h.notified, 3, "an attention change notifies the unclosed episode")
	_, againErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelWatchOnly)
	s.Require().NoError(againErr)
	s.Len(h.notified, 3, "setting the current level does not notify")

	h.clock.Set(h.settlements[1].DueAt)
	s.Require().NoError(h.alerts.settleAlertEpisode(ctx, h.settlements[1]))
	s.Len(h.notified, 4, "closing the episode notifies")
	_, joinErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelJoinOnly)
	s.Require().NoError(joinErr)
	s.Len(h.notified, 4, "a closed episode is not notified of attention changes")
}

func (s *AlertServiceSuite) TestDeliveryOrderDoesNotChangeWindows() {
	deliveries := []testAlert{
		{key: "a", severity: schematypes.SignalSeverityWarning},
		{at: 2, key: "a", severity: schematypes.SignalSeverityCritical},
		{at: 1, start: 1, key: "b"},
		{at: 10, key: "a", resolved: true},
		{at: 12, start: 1, key: "b", resolved: true},
		{at: 13, start: 13, key: "a"},
		{at: 15, start: 13, key: "a", resolved: true},
		{at: 120, start: 120, key: "b"},
		{at: 123, start: 120, key: "b"},
	}
	expected := s.recordedWindows(deliveries)
	random := rand.New(rand.NewPCG(1, 2))
	for permutation := range 5 {
		shuffled := slices.Clone(deliveries)
		random.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		s.Equal(expected, s.recordedWindows(shuffled), "permutation %d", permutation)
	}
}

// recordedWindows records the deliveries in a new tenant, settles every episode at one time and
// describes the stored windows. Episode grouping is left out: it follows the windows known when each
// window was recorded.
func (s *AlertServiceSuite) recordedWindows(deliveries []testAlert) []string {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	for _, delivery := range deliveries {
		delivery.timeout = new(1800)
		s.record(ctx, tdb, h, delivery)
	}
	h.clock.Set(alertTestStart.Add(3 * time.Hour))
	var described []string
	for _, episode := range s.episodes(ctx, tdb) {
		s.Require().NoError(h.alerts.settleAlertEpisode(ctx, jobs.SettleAlertEpisode{EpisodeID: episode.ID}))
	}
	for _, episode := range s.episodes(ctx, tdb) {
		for _, w := range episode.Edges.Instances {
			resolvedAt, endedAt, endReason := "-", "-", "-"
			if w.ResolvedAt != nil {
				resolvedAt = w.ResolvedAt.String()
			}
			if w.EndedAt != nil {
				endedAt, endReason = w.EndedAt.String(), w.EndReason.String()
			}
			described = append(described, fmt.Sprint(w.InstanceKey, w.FiredAt, w.LastObservedAt, resolvedAt,
				endedAt, endReason, w.Severity, len(w.Edges.Events)))
		}
	}
	slices.Sort(described)
	return described
}

func (s *AlertServiceSuite) TestOlderNotificationsDoNotOverwriteDefinitionMetadata() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	suffix := uuid.NewString()
	s.record(ctx, tdb, h, testAlert{at: 5, title: "Newest", ref: "b-" + suffix, timeout: new(60)})
	s.record(ctx, tdb, h, testAlert{at: 5, title: "Tied lower ref", ref: "a-" + suffix, timeout: new(120)})
	recorded := s.record(ctx, tdb, h, testAlert{at: 1, title: "Older"})

	s.Equal("Newest", recorded.Title)
	s.Equal(60, recorded.ResolutionTimeoutSeconds)
}

func (s *AlertServiceSuite) TestDefinitionSettings() {
	_, tdb := s.SetupTestDatabase()
	ctx, _ := s.NewIdentity(tdb, "alert owner")
	h := s.newFixture(tdb)
	definition := s.record(ctx, tdb, h, testAlert{})

	labelled, labelsErr := h.alerts.SetAlertIdentityGroupLabels(ctx, definition.ID, []string{" service ", "service", ""})
	s.Require().NoError(labelsErr)
	s.Equal([]string{"service"}, labelled.IdentityGroupLabels)
	proposed := s.record(ctx, tdb, h, testAlert{at: 1, groupBy: []string{"pod"}})
	s.Equal([]string{"service"}, proposed.IdentityGroupLabels, "a source proposal does not replace a set value")

	unchanged, defaultErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelDefault)
	s.Require().NoError(defaultErr)
	s.Nil(unchanged.SituationSignalAttentionID, "setting the current level changes nothing")

	watched, watchErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelWatchOnly)
	s.Require().NoError(watchErr)
	s.Require().NotNil(watched.Edges.SituationSignalAttention)
	s.Equal(ssa.LevelWatchOnly, watched.Edges.SituationSignalAttention.Level)
	s.NotNil(watched.Edges.SituationSignalAttention.SetByUserID)

	h.clock.Advance(time.Hour)
	again, againErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, ssa.LevelWatchOnly)
	s.Require().NoError(againErr)
	s.True(watched.Edges.SituationSignalAttention.SetAt.Equal(again.Edges.SituationSignalAttention.SetAt))
}

func (s *AlertServiceSuite) TestIdentityGroupLabelsPreserveReturnedAttention() {
	for _, level := range []ssa.Level{ssa.LevelWatchOnly, ssa.LevelJoinOnly} {
		s.Run(level.String(), func() {
			_, tdb := s.SetupTestDatabase()
			ctx, _ := s.NewIdentity(tdb, "alert owner")
			h := s.newFixture(tdb)
			definition := s.record(ctx, tdb, h, testAlert{})
			setAt := h.clock.Now()

			_, attentionErr := h.alerts.SetAlertSituationSignalAttention(ctx, definition.ID, level)
			s.Require().NoError(attentionErr)
			h.clock.Advance(time.Hour)

			labelled, labelsErr := h.alerts.SetAlertIdentityGroupLabels(ctx, definition.ID, []string{"service"})

			s.Require().NoError(labelsErr)
			s.Equal([]string{"service"}, labelled.IdentityGroupLabels)
			s.Require().NotNil(labelled.Edges.SituationSignalAttention)
			s.Equal(level, labelled.Edges.SituationSignalAttention.Level)
			s.True(setAt.Equal(labelled.Edges.SituationSignalAttention.SetAt))

			h.clock.Advance(time.Hour)
			cleared, clearErr := h.alerts.SetAlertIdentityGroupLabels(ctx, definition.ID, nil)

			s.Require().NoError(clearErr)
			s.Empty(cleared.IdentityGroupLabels)
			s.Require().NotNil(cleared.Edges.SituationSignalAttention)
			s.Equal(level, cleared.Edges.SituationSignalAttention.Level)
			s.True(setAt.Equal(cleared.Edges.SituationSignalAttention.SetAt))
		})
	}
}

func (s *AlertServiceSuite) TestLoadSignals() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	h.clock.Set(alertTestStart.Add(20 * time.Minute))

	s.record(ctx, tdb, h, testAlert{definition: "latency", key: "a", labels: map[string]string{"pod": "a"}, severity: schematypes.SignalSeverityWarning})
	s.record(ctx, tdb, h, testAlert{definition: "latency", at: 1, start: 1, key: "b", severity: schematypes.SignalSeverityCritical})
	s.record(ctx, tdb, h, testAlert{definition: "latency", at: 5, start: 1, key: "b", resolved: true})
	s.record(ctx, tdb, h, testAlert{definition: "errors", key: "c", resolved: true, at: 2})

	episodes := s.episodes(ctx, tdb)
	s.Require().Len(episodes, 2)
	entityIDs := []uuid.UUID{*episodes[0].KnowledgeEntityID, *episodes[1].KnowledgeEntityID}
	signals, loadErr := h.alerts.LoadSignals(ctx, entityIDs, situations.LoadSignalsOptions{})
	s.Require().NoError(loadErr)
	s.Require().Len(signals, 2)

	latency := signals[slices.IndexFunc(signals, func(signal situations.Signal) bool { return signal.Active })]
	s.Equal(situations.SignalKindAlertEpisode, latency.Kind)
	s.Equal(schematypes.SignalSeverityCritical, latency.Severity)
	s.Equal(ssa.LevelDefault, latency.SignalAttention)
	s.Nil(latency.FinishedAt)
	s.Equal(3, latency.Revision)
	s.Len(latency.EventIDs, 3)
	s.True(alertTestStart.Add(5 * time.Minute).Equal(latency.LastActivityAt))
	s.Equal(2, latency.Alert.InstanceCount)
	s.Equal(1, latency.Alert.ActiveInstanceCount)
	s.Equal(int64(20*60), latency.Alert.ActiveSeconds)
	s.Require().Len(latency.Alert.ActiveInstances, 1)
	s.Equal("a", latency.Alert.ActiveInstances[0].Labels["pod"])

	finished := signals[slices.IndexFunc(signals, func(signal situations.Signal) bool { return !signal.Active })]
	s.Require().NotNil(finished.FinishedAt, "a resolved episode closes once the grace has passed")
	s.Equal(1, finished.Revision)
}

func (s *AlertServiceSuite) TestLoadSignalsBaseline() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	h.clock.Set(alertTestStart.Add(20 * time.Minute))
	day := 24 * 60

	// Beyond the novelty window: it only dates the tenant's history.
	s.record(ctx, tdb, h, testAlert{at: -40 * day, start: -40 * day, key: "a"})
	s.record(ctx, tdb, h, testAlert{at: -40*day + 5, start: -40 * day, key: "a", resolved: true})
	// Two resolved firings of 10 and 20 minutes, and one that timed out.
	s.record(ctx, tdb, h, testAlert{at: -10 * day, start: -10 * day, key: "a"})
	s.record(ctx, tdb, h, testAlert{at: -10*day + 10, start: -10 * day, key: "a", resolved: true})
	s.record(ctx, tdb, h, testAlert{at: -5 * day, start: -5 * day, key: "a"})
	s.record(ctx, tdb, h, testAlert{at: -5*day + 20, start: -5 * day, key: "a", resolved: true})
	s.record(ctx, tdb, h, testAlert{at: -3 * day, start: -3 * day, key: "a"})
	s.record(ctx, tdb, h, testAlert{key: "a"})

	episodes := s.episodes(ctx, tdb)
	s.Require().Len(episodes, 5)
	signals, loadErr := h.alerts.LoadSignals(ctx, []uuid.UUID{*episodes[4].KnowledgeEntityID}, situations.LoadSignalsOptions{History: true})
	s.Require().NoError(loadErr)
	s.Require().Len(signals, 1)

	baseline := signals[0].Alert.Baseline
	s.Equal(40, baseline.ObservedHistoryAgeDays)
	s.True(baseline.HasSufficientHistory)
	s.Equal(3, baseline.Occurrences, "earlier episodes within the novelty window")
	s.Require().NotNil(baseline.MedianDurationSeconds, "resolved firings give a median; a timed-out one does not")
	s.Equal(int64(15*60), *baseline.MedianDurationSeconds)
}
