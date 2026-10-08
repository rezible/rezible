package db

import (
	"context"
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
	aie "github.com/rezible/rezible/ent/alertinstanceevent"
	"github.com/rezible/rezible/ent/incident"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	sitha "github.com/rezible/rezible/ent/situationhazardassessment"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/situations"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type SituationServiceSuite struct {
	test.Suite
}

func TestSituationServiceSuite(t *testing.T) {
	suite.Run(t, &SituationServiceSuite{
		Suite: test.NewSuite(),
	})
}

var situationTestStart = time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

type situationServiceFixture struct {
	clock      *test.Clock
	knowledge  *KnowledgeGraphIngestionService
	alerts     *AlertService
	situations *SituationService
	// notified are the signals captured from inserted process-situation-signal jobs, not yet processed.
	notified []uuid.UUID
	// evaluations are the captured evaluate-situation jobs, in order.
	evaluations []jobs.EvaluateSituation
}

func (s *SituationServiceSuite) newFixture(tdb rez.Database) *situationServiceFixture {
	h := &situationServiceFixture{clock: test.NewClock(situationTestStart.Add(30 * time.Minute))}
	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
			switch args := args.(type) {
			case jobs.ProcessSituationSignal:
				h.notified = append(h.notified, args.SignalEntityID)
			case jobs.EvaluateSituation:
				if args.DueAt != nil {
					s.True(args.DueAt.Equal(opts.ScheduledAt), "an evaluation is scheduled at its due time")
				}
				h.evaluations = append(h.evaluations, args)
			}
			return &rivertype.JobInsertResult{Job: &rivertype.JobRow{}}, nil
		}).
		Maybe()
	agentService := &AiAgentSessionService{
		db:   tdb,
		jobs: jobService,
	}
	knowledgeQuery, queryServiceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryServiceErr)
	knowledge, knowledgeErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)
	analysisService, analysisServiceErr := NewSystemAnalysisService(tdb, knowledgeQuery)
	s.Require().NoError(analysisServiceErr)

	h.knowledge = knowledge
	alerts, alertsErr := NewAlertService(rez.DefaultConfig().Alerts, h.clock, tdb, jobService, knowledge, NewSituationSignalService(jobService))
	s.Require().NoError(alertsErr)
	investigations := NewInvestigationService(tdb, agentService, jobService)
	situationService, situationServiceErr := NewSituationService(tdb, h.clock, jobService, investigations, analysisService, knowledgeQuery, nil, alerts)
	s.Require().NoError(situationServiceErr)
	h.alerts = alerts
	h.situations = situationService
	return h
}

func (s *SituationServiceSuite) entityRef(category kne.Category, name string) rez.KnowledgeEntityRef {
	resourceRef := rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "situations", ResourceRef: name}
	return rez.KnowledgeEntityRef{Category: category, Kind: string(category), ProviderResourceRef: resourceRef}
}

func (s *SituationServiceSuite) testEvent(ctx context.Context, tdb rez.Database, resourceRef string, at time.Time) *ent.NormalizedEvent {
	createEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("situations").
		SetProviderResourceRef(resourceRef).
		SetProviderEventSource("situation-test").
		SetProviderEventRef(uuid.NewString()).
		SetKind("alert").
		SetOccurredAt(at).
		SetReceivedAt(at).
		SetAttributes([]byte(`{}`))
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)
	return event
}

func (s *SituationServiceSuite) observed(at time.Time, subject rez.KnowledgeSubjectRef) rez.KnowledgeEvidenceRef {
	return rez.KnowledgeEvidenceRef{Kind: knev.KindObserved, Assertion: "observed", EffectiveAt: at, Subject: subject}
}

// ingestRelationship records the relationship from source to target.
func (s *SituationServiceSuite) ingestRelationship(ctx context.Context, tdb rez.Database, h *situationServiceFixture, source rez.KnowledgeEntityRef, predicate knr.Predicate, target rez.KnowledgeEntityRef) {
	event := s.testEvent(ctx, tdb, "topology", situationTestStart)
	relationshipRef := rez.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "situations",
		ResourceRef:       source.ProviderResourceRef.ResourceRef + "/" + target.ProviderResourceRef.ResourceRef,
	}
	relationship := rez.KnowledgeRelationshipRef{
		Predicate:           predicate,
		ProviderResourceRef: relationshipRef,
		Source:              source,
		Target:              target,
	}
	evidence := []rez.KnowledgeEvidenceRef{
		s.observed(event.OccurredAt, rez.KnowledgeSubjectRef{Entity: &source}),
		s.observed(event.OccurredAt, rez.KnowledgeSubjectRef{Entity: &target}),
		s.observed(event.OccurredAt, rez.KnowledgeSubjectRef{Relationship: &relationship}),
	}
	s.Require().NoError(h.knowledge.IngestEvidence(ctx, event, evidence...))
}

// signal records a firing critical alert of its own definition, starting `start` minutes after
// situationTestStart, whose notification observed the entities. It returns the episode's knowledge entity.
func (s *SituationServiceSuite) signal(ctx context.Context, tdb rez.Database, h *situationServiceFixture, definition string, start int, entities ...rez.KnowledgeEntityRef) uuid.UUID {
	return s.signalWithSeverity(ctx, tdb, h, definition, schematypes.SignalSeverityCritical, start, entities...)
}

func (s *SituationServiceSuite) signalWithSeverity(ctx context.Context, tdb rez.Database, h *situationServiceFixture, definition string, severity schematypes.SignalSeverity, start int, entities ...rez.KnowledgeEntityRef) uuid.UUID {
	startedAt := situationTestStart.Add(time.Duration(start) * time.Minute)
	event := s.testEvent(ctx, tdb, definition, startedAt)
	definitionRef := s.entityRef(kne.CategorySignal, definition)
	evidence := []rez.KnowledgeEvidenceRef{s.observed(startedAt, rez.KnowledgeSubjectRef{Entity: &definitionRef})}
	for _, entity := range entities {
		evidence = append(evidence, s.observed(startedAt, rez.KnowledgeSubjectRef{Entity: &entity}))
	}
	s.Require().NoError(h.knowledge.IngestEvidence(ctx, event, evidence...))
	definitionAlias, aliasErr := h.knowledge.ResolveInternalSubject(ctx, rez.KnowledgeSubjectRef{Entity: &definitionRef})
	s.Require().NoError(aliasErr)

	params := rez.RecordAlertInstanceParams{
		Event:      event,
		Definition: rez.AlertDefinitionValues{KnowledgeEntityID: *definitionAlias.EntityID, Title: definition},
		Instance: rez.AlertInstanceValues{
			InstanceID: definition,
			Severity:   severity,
			Firing:     true,
			StartedAt:  startedAt,
		},
	}
	_, recordErr := h.alerts.RecordAlertInstance(ctx, params)
	s.Require().NoError(recordErr)
	queryEpisode := tdb.Client(ctx).AlertEpisode.Query().
		Where(ale.HasInstancesWith(ali.HasEventsWith(aie.EventID(event.ID))))
	episode, episodeErr := queryEpisode.Only(ctx)
	s.Require().NoError(episodeErr)
	return *episode.KnowledgeEntityID
}

// create creates a candidate seeded by the first signal, with one group per title.
func (s *SituationServiceSuite) create(ctx context.Context, h *situationServiceFixture, groups map[string][]uuid.UUID, seed uuid.UUID) *ent.Situation {
	params := rez.CreateSituationParams{Title: "Checkout degradation", SeedEntityID: seed}
	for title, signalIDs := range groups {
		group := rez.SituationObservationGroupParams{Title: title, SignalEntityIDs: signalIDs}
		params.ObservationGroups = append(params.ObservationGroups, group)
	}
	created, createErr := h.situations.CreateSituation(ctx, params)
	s.Require().NoError(createErr)
	return created
}

func (s *SituationServiceSuite) entities(situation *ent.Situation) map[uuid.UUID]bool {
	matching := make(map[uuid.UUID]bool)
	for _, entity := range situation.Edges.Entities {
		matching[entity.KnowledgeEntityID] = entity.Matching
	}
	return matching
}

func (s *SituationServiceSuite) members(situation *ent.Situation) map[uuid.UUID]*ent.SituationSignal {
	members := make(map[uuid.UUID]*ent.SituationSignal)
	for _, group := range situation.Edges.ObservationGroups {
		for _, member := range group.Edges.Signals {
			members[member.KnowledgeEntityID] = member
		}
	}
	return members
}

func (s *SituationServiceSuite) entityID(ctx context.Context, h *situationServiceFixture, ref rez.KnowledgeEntityRef) uuid.UUID {
	alias, aliasErr := h.knowledge.ResolveInternalSubject(ctx, rez.KnowledgeSubjectRef{Entity: &ref})
	s.Require().NoError(aliasErr)
	return *alias.EntityID
}

func (s *SituationServiceSuite) TestCreateAndAttach() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	handler := s.entityRef(kne.CategoryComponent, "checkout-handler")
	s.ingestRelationship(ctx, tdb, h, checkout, knr.PredicateContains, handler)
	var fleet []rez.KnowledgeEntityRef
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		fleet = append(fleet, s.entityRef(kne.CategoryContainer, "fleet-"+name))
	}

	seed := s.signal(ctx, tdb, h, "checkout-errors", 5, handler)
	broad := s.signal(ctx, tdb, h, "fleet-errors", 6, fleet...)
	created := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {seed}, "Fleet": {broad}}, seed)

	s.Nil(created.RaisedAt, "a new situation is a candidate")
	s.True(situationTestStart.Add(5 * time.Minute).Equal(created.OpenedAt))
	s.True(h.clock.Now().Equal(created.CreatedAt))
	members := s.members(created)
	s.Equal(sitsig.MatchKindSeed, members[seed].MatchKind)
	s.Equal(sitsig.MatchKindManual, members[broad].MatchKind)
	entities := s.entities(created)
	s.Len(entities, 6)
	s.True(entities[s.entityID(ctx, h, checkout)], "a component resolves to its container")
	s.False(entities[s.entityID(ctx, h, fleet[0])], "a broad signal's entities do not match")

	_, conflictErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:             "Duplicate",
		SeedEntityID:      seed,
		ObservationGroups: []rez.SituationObservationGroupParams{{Title: "Checkout", SignalEntityIDs: []uuid.UUID{seed}}},
	})
	s.ErrorIs(conflictErr, errs.ErrConflict)
	containerID := s.entityID(ctx, h, checkout)
	_, kindErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:             "Not a signal",
		SeedEntityID:      containerID,
		ObservationGroups: []rez.SituationObservationGroupParams{{Title: "Checkout", SignalEntityIDs: []uuid.UUID{containerID}}},
	})
	s.ErrorIs(kindErr, errs.ErrInvalidInput)

	earlier := s.signal(ctx, tdb, h, "checkout-latency", 1, checkout)
	attach := AttachSituationSignalsParams{
		SituationID:     created.ID,
		GroupTitle:      "Checkout",
		SignalEntityIDs: []uuid.UUID{earlier, seed},
		MatchKind:       sitsig.MatchKindSharedEntity,
	}
	s.Require().NoError(h.situations.AttachSituationSignals(ctx, attach))
	attached, getErr := h.situations.GetSituation(ctx, created.ID)
	s.Require().NoError(getErr)
	s.Len(attached.Edges.ObservationGroups, 2, "the group with the same title is reused")
	s.Len(s.members(attached), 3)
	s.Equal(sitsig.MatchKindSeed, s.members(attached)[seed].MatchKind, "a signal already held is skipped")
	s.True(situationTestStart.Add(time.Minute).Equal(attached.OpenedAt), "an earlier signal lowers the opening")
	s.Len(s.entities(attached), 6)
}

func (s *SituationServiceSuite) TestMerge() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	checkout := s.entityRef(kne.CategoryContainer, "checkout")
	payments := s.entityRef(kne.CategoryContainer, "payments")
	first := s.signal(ctx, tdb, h, "checkout-errors", 5, checkout)
	second := s.signal(ctx, tdb, h, "checkout-latency", 2, checkout)
	third := s.signal(ctx, tdb, h, "payment-errors", 3, payments)
	target := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first}}, first)
	source := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {second}, "Payments": {third}}, second)

	h.clock.Advance(time.Minute)
	params := rez.MergeSituationsParams{SourceID: source.ID, TargetID: target.ID, Explanation: "same outage"}
	merged, mergeErr := h.situations.MergeSituations(ctx, params)
	s.Require().NoError(mergeErr)

	s.Len(merged.Edges.ObservationGroups, 2, "same-title groups join; others move")
	members := s.members(merged)
	s.Len(members, 3)
	s.Equal(sitsig.MatchKindManual, members[third].MatchKind)
	s.Equal("same outage", members[third].MatchExplanation)
	s.True(h.clock.Now().Equal(members[second].AttachedAt), "a merge resets attachment time")
	s.True(situationTestStart.Add(2 * time.Minute).Equal(merged.OpenedAt))
	s.True(s.entities(merged)[s.entityID(ctx, h, payments)])

	closed, getErr := h.situations.GetSituation(ctx, source.ID)
	s.Require().NoError(getErr)
	s.Equal(sit.CloseReasonMerged, *closed.CloseReason)
	s.Equal(source.SeedEntityID, closed.SeedEntityID)
	s.Empty(closed.Edges.Entities)
	s.Require().Len(closed.Edges.Links, 1)
	s.Equal(sitlink.KindMergedInto, closed.Edges.Links[0].Kind)
	s.Equal(target.ID, closed.Edges.Links[0].LinkedSituationID)
	s.Equal(sitact.ActionMerged, closed.Edges.Actions[len(closed.Edges.Actions)-1].Action)

	_, closedErr := h.situations.MergeSituations(ctx, params)
	s.ErrorIs(closedErr, errs.ErrConflict)
}

func (s *SituationServiceSuite) TestLifecycleVerbs() {
	_, tdb := s.SetupTestDatabase()
	ctx, identity := s.NewIdentity(tdb, "responder")
	h := s.newFixture(tdb)
	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	created := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	actionCount := func(situation *ent.Situation) int { return len(situation.Edges.Actions) }

	muted, muteErr := h.situations.SetSituationMute(ctx, created.ID, &rez.SituationMute{Reason: sit.MuteReasonNotNoteworthy})
	s.Require().NoError(muteErr)
	again, _ := h.situations.SetSituationMute(ctx, created.ID, &rez.SituationMute{Reason: sit.MuteReasonNotNoteworthy})
	s.Equal(actionCount(muted), actionCount(again), "the same reason is a no-op")
	h.clock.Advance(time.Minute)
	remuted, _ := h.situations.SetSituationMute(ctx, created.ID, &rez.SituationMute{Reason: sit.MuteReasonExpected})
	s.Equal(actionCount(muted)+1, actionCount(remuted))
	s.True(muted.MutedAt.Equal(*remuted.MutedAt), "a new reason keeps the mute time")

	held, holdErr := h.situations.SetSituationHold(ctx, created.ID, &rez.SituationHold{})
	s.Require().NoError(holdErr)
	s.True(h.clock.Now().Add(situations.HoldDefault).Equal(*held.HoldUntil))
	_, pastErr := h.situations.SetSituationHold(ctx, created.ID, &rez.SituationHold{Until: new(situationTestStart)})
	s.ErrorIs(pastErr, errs.ErrInvalidInput)

	raised, raiseErr := h.situations.RaiseSituation(ctx, created.ID, rez.RaiseSituationParams{Reason: "looks real"})
	s.Require().NoError(raiseErr)
	s.Nil(raised.MutedAt, "raising clears a mute")
	s.Equal([]sitact.Action{sitact.ActionMuted, sitact.ActionMuted, sitact.ActionHeld, sitact.ActionUnmuted, sitact.ActionRaised}, s.actions(raised))
	s.Equal(identity.Session.UserID, *raised.Edges.Actions[0].UserID)
	reraised, _ := h.situations.RaiseSituation(ctx, created.ID, rez.RaiseSituationParams{})
	s.True(raised.RaisedAt.Equal(*reraised.RaisedAt))
	s.Equal(actionCount(raised), actionCount(reraised))

	closed, closeErr := h.situations.CloseSituation(ctx, created.ID, rez.CloseSituationParams{Note: "recovered"})
	s.Require().NoError(closeErr)
	s.Equal(sit.CloseReasonStabilized, *closed.CloseReason)
	_, repeatErr := h.situations.CloseSituation(ctx, created.ID, rez.CloseSituationParams{})
	s.NoError(repeatErr)
	_, closedMuteErr := h.situations.SetSituationMute(ctx, created.ID, nil)
	s.ErrorIs(closedMuteErr, errs.ErrConflict)

	other := s.signal(ctx, tdb, h, "payment-errors", 5)
	dismissed := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {other}}, other)
	_, _ = h.situations.SetSituationMute(ctx, dismissed.ID, &rez.SituationMute{Reason: sit.MuteReasonExpected})
	closedMuted, _ := h.situations.CloseSituation(ctx, dismissed.ID, rez.CloseSituationParams{})
	s.Equal(sit.CloseReasonDismissed, *closedMuted.CloseReason)
}

func (s *SituationServiceSuite) actions(situation *ent.Situation) []sitact.Action {
	actions := make([]sitact.Action, len(situation.Edges.Actions))
	for i, action := range situation.Edges.Actions {
		actions[i] = action.Action
	}
	return actions
}

// newIncidents returns an incident service over the fixture's situations, and a mutation that fills a new
// incident's required fields.
func (s *SituationServiceSuite) newIncidents(ctx context.Context, tdb rez.Database, h *situationServiceFixture) (*IncidentService, func(*ent.IncidentMutation)) {
	msgs := mocks.NewMockMessageQueue(s.T())
	msgs.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Maybe()
	retrospectives, retrospectivesErr := NewRetrospectiveService(tdb)
	s.Require().NoError(retrospectivesErr)
	incidents, incidentsErr := NewIncidentService(tdb, msgs, h.situations, retrospectives)
	s.Require().NoError(incidentsErr)
	createSeverity := tdb.Client(ctx).IncidentSeverity.Create().
		SetName("SEV-1").
		SetRank(1)
	severity, severityErr := createSeverity.Save(ctx)
	s.Require().NoError(severityErr)
	createType := tdb.Client(ctx).IncidentType.Create().
		SetName("Outage")
	incidentType, typeErr := createType.Save(ctx)
	s.Require().NoError(typeErr)
	newIncident := func(m *ent.IncidentMutation) {
		m.SetTitle("Checkout outage")
		m.SetSeverityID(severity.ID)
		m.SetTypeID(incidentType.ID)
	}
	return incidents, newIncident
}

func (s *SituationServiceSuite) TestIncidentLinks() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	incidents, newIncident := s.newIncidents(ctx, tdb, h)

	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	candidate := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	linked, linkErr := incidents.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		newIncident(m)
		m.AddSituationIDs(candidate.ID)
	})
	s.Require().NoError(linkErr)
	raised, getErr := h.situations.GetSituation(ctx, candidate.ID)
	s.Require().NoError(getErr)
	s.NotNil(raised.RaisedAt)
	s.NotNil(raised.Edges.Investigation, "linking an incident starts the investigation")
	s.Equal(linked.ID, raised.Edges.Incidents[0].ID)

	other := s.signal(ctx, tdb, h, "payment-errors", 5)
	closed := s.create(ctx, h, map[string][]uuid.UUID{"Payments": {other}}, other)
	_, closeErr := h.situations.CloseSituation(ctx, closed.ID, rez.CloseSituationParams{})
	s.Require().NoError(closeErr)
	_, closedErr := incidents.Set(ctx, linked.ID, func(m *ent.IncidentMutation) {
		m.SetTitle("Renamed")
		m.AddSituationIDs(closed.ID)
	})
	s.ErrorIs(closedErr, errs.ErrConflict)
	unchanged, incidentErr := incidents.Get(ctx, incident.ID(linked.ID))
	s.Require().NoError(incidentErr)
	s.Equal("Checkout outage", unchanged.Title, "the incident write rolls back")

	_, clearErr := incidents.Set(ctx, linked.ID, func(m *ent.IncidentMutation) {
		m.ClearSituations()
	})
	s.Require().NoError(clearErr)
	cleared, clearedErr := h.situations.GetSituation(ctx, candidate.ID)
	s.Require().NoError(clearedErr)
	s.Empty(cleared.Edges.Incidents, "clearing an incident's situations unlinks them")
	s.NotNil(cleared.RaisedAt, "unlinking changes no stage")
}

func (s *SituationServiceSuite) TestListAlertEpisodesOfASituation() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	first := s.signal(ctx, tdb, h, "checkout-errors", 5)
	second := s.signal(ctx, tdb, h, "checkout-latency", 6)
	other := s.signal(ctx, tdb, h, "payment-errors", 5)
	situation := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {first, second}}, first)
	s.create(ctx, h, map[string][]uuid.UUID{"Payments": {other}}, other)

	episodes, listErr := h.alerts.ListAlertEpisodes(ctx, rez.ListAlertEpisodesParams{SituationID: situation.ID})
	s.Require().NoError(listErr)
	var entityIDs []uuid.UUID
	for _, episode := range episodes.Data {
		entityIDs = append(entityIDs, *episode.KnowledgeEntityID)
	}
	s.ElementsMatch([]uuid.UUID{first, second}, entityIDs)
}

func (s *SituationServiceSuite) TestListSituationsByStageAndMute() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	ids := make(map[string]uuid.UUID)
	for _, name := range []string{"candidate", "raised", "muted", "closed"} {
		signal := s.signal(ctx, tdb, h, name, 5)
		ids[name] = s.create(ctx, h, map[string][]uuid.UUID{name: {signal}}, signal).ID
	}
	_, raiseErr := h.situations.RaiseSituation(ctx, ids["raised"], rez.RaiseSituationParams{})
	s.Require().NoError(raiseErr)
	_, muteErr := h.situations.SetSituationMute(ctx, ids["muted"], &rez.SituationMute{Reason: sit.MuteReasonExpected})
	s.Require().NoError(muteErr)
	_, closeErr := h.situations.CloseSituation(ctx, ids["closed"], rez.CloseSituationParams{})
	s.Require().NoError(closeErr)

	list := func(params rez.ListSituationsParams) []uuid.UUID {
		listed, listErr := h.situations.ListSituations(ctx, params)
		s.Require().NoError(listErr)
		var listedIDs []uuid.UUID
		for _, situation := range listed.Data {
			listedIDs = append(listedIDs, situation.ID)
		}
		return listedIDs
	}
	candidates := rez.ListSituationsParams{Stages: []rez.SituationStage{rez.SituationStageCandidate}, Muted: new(false)}
	s.ElementsMatch([]uuid.UUID{ids["candidate"]}, list(candidates))
	open := rez.ListSituationsParams{Stages: []rez.SituationStage{rez.SituationStageRaised, rez.SituationStageClosed}}
	s.ElementsMatch([]uuid.UUID{ids["raised"], ids["closed"]}, list(open))
	s.ElementsMatch([]uuid.UUID{ids["muted"]}, list(rez.ListSituationsParams{Muted: new(true)}))
}

func (s *SituationServiceSuite) TestInvestigationAnalysisDeduplicatesSubjects() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	signal := s.signal(ctx, tdb, h, "checkout-errors", 5, s.entityRef(kne.CategoryContainer, "checkout"))
	params := rez.CreateSituationParams{
		Title:             "Checkout degradation",
		SeedEntityID:      signal,
		ObservationGroups: []rez.SituationObservationGroupParams{{Title: "Checkout", SignalEntityIDs: []uuid.UUID{signal}}},
		Raise:             &rez.RaiseSituationParams{StartInvestigation: true},
	}
	created, createErr := h.situations.CreateSituation(ctx, params)
	s.Require().NoError(createErr)
	s.Require().NotNil(created.Edges.Investigation)
	queryInvestigation := tdb.Client(ctx).Investigation.Query()
	investigation, investigationErr := queryInvestigation.Only(ctx)
	s.Require().NoError(investigationErr)
	eventIDs, eventsErr := h.situations.ListSituationEventIDs(ctx, created.ID)
	s.Require().NoError(eventsErr)
	s.Require().Len(eventIDs, 1)

	countSubjects := func() int {
		queryEntries := tdb.Client(ctx).SystemAnalysisEntry.Query().
			Where(sae.AnalysisID(investigation.SystemAnalysisID))
		entries, entriesErr := queryEntries.All(ctx)
		s.Require().NoError(entriesErr)
		s.Require().Len(entries, 1)
		s.Equal("normalized_event:"+eventIDs[0].String(), *entries[0].Reference)
		querySubjects := tdb.Client(ctx).SystemAnalysisEntrySubject.Query().
			Where(saes.EntryID(entries[0].ID))
		subjects, subjectsErr := querySubjects.Count(ctx)
		s.Require().NoError(subjectsErr)
		return subjects
	}
	subjects := countSubjects()
	s.Positive(subjects)
	s.Require().NoError(h.situations.prepareOrRefreshSituationAnalysis(ctx, investigation.SystemAnalysisID, eventIDs))
	s.Equal(subjects, countSubjects(), "refreshing adds no duplicate subjects")
}

func (s *SituationServiceSuite) TestSystemHazardRetirementAndRiskAssessmentRevisions() {
	ctx, tdb := s.SetupTestDatabase()

	knowledge, knowledgeServiceErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeServiceErr)

	hazards, hazardServiceErr := NewSystemHazardService(tdb, knowledge)
	s.Require().NoError(hazardServiceErr)

	hazardParams := rez.CreateSystemHazardParams{
		Title:                 "Database unavailable",
		Description:           "A required datastore cannot serve requests.",
		PotentialConsequences: "Requests fail or become unavailable.",
	}

	hazard, createErr := hazards.CreateSystemHazard(ctx, hazardParams)
	s.Require().NoError(createErr)

	firstParams := rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "possible",
		Consequence:    "major",
		RiskLevel:      "high",
		AssessedAt:     time.Now().UTC(),
	}

	first, firstErr := hazards.AddSystemHazardRiskAssessment(ctx, firstParams)
	s.Require().NoError(firstErr)

	secondParams := rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "likely",
		Consequence:    "major",
		RiskLevel:      "critical",
		Rationale:      "Recent capacity data increased confidence.",
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	}

	second, secondErr := hazards.AddSystemHazardRiskAssessment(ctx, secondParams)
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	latest, latestErr := hazards.GetLatestSystemHazardRiskAssessment(ctx, hazard.ID)
	s.Require().NoError(latestErr)
	s.Equal(second.ID, latest.ID)

	retired, retireErr := hazards.RetireSystemHazard(ctx, hazard.ID)
	s.Require().NoError(retireErr)
	s.Equal("retired", string(retired.Status))
}

func (s *SituationServiceSuite) TestSituationHazardAssessmentRevisionsAndAssessorConstraint() {
	ctx, tdb := s.SetupTestDatabase()

	knowledge, knowledgeServiceErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeServiceErr)

	hazards, hazardServiceErr := NewSystemHazardService(tdb, knowledge)
	s.Require().NoError(hazardServiceErr)

	h := s.newFixture(tdb)

	hazardParams := rez.CreateSystemHazardParams{
		Title: "Checkout unavailable",
	}

	hazard, hazardErr := hazards.CreateSystemHazard(ctx, hazardParams)
	s.Require().NoError(hazardErr)

	signal := s.signal(ctx, tdb, h, "checkout-errors", 5)
	situation := s.create(ctx, h, map[string][]uuid.UUID{"Checkout": {signal}}, signal)
	userId, userIdErr := tdb.Client(ctx).User.Query().FirstID(ctx)
	s.Require().NoError(userIdErr)

	firstParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    situation.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusSuspected,
		Summary:        "The observed symptoms may represent this hazard.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC(),
	}

	first, firstErr := h.situations.AddSituationHazardAssessment(ctx, firstParams)
	s.Require().NoError(firstErr)

	secondParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    situation.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusDisproven,
		Summary:        "The symptoms do not match the hazard after review.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	}

	second, secondErr := h.situations.AddSituationHazardAssessment(ctx, secondParams)
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	noAssessorParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    situation.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Missing provenance.",
	}

	_, noAssessorErr := h.situations.AddSituationHazardAssessment(ctx, noAssessorParams)
	s.ErrorIs(noAssessorErr, errs.ErrInvalidInput)

	twoAssessorsParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    situation.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Two provenance sources are not supported.",
		UserID:         &userId,
		AgentTurnID:    new(uuid.New()),
	}

	_, twoAssessorsErr := h.situations.AddSituationHazardAssessment(ctx, twoAssessorsParams)
	s.ErrorIs(twoAssessorsErr, errs.ErrInvalidInput)
}
