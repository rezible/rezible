package db

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	aie "github.com/rezible/rezible/ent/alertinstanceevent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type SituationBriefServiceSuite struct {
	test.Suite
}

func TestSituationBriefServiceSuite(t *testing.T) {
	suite.Run(t, &SituationBriefServiceSuite{Suite: test.NewSuite()})
}

// briefOpenedAt is when the brief tests' situations open: changesAt(0), so deployments use the same times as
// the changes tests.
var briefOpenedAt = changesTestStart

type briefFixture struct {
	tdb        rez.Database
	clock      *test.Clock
	knowledge  *KnowledgeGraphIngestionService
	alerts     *AlertService
	situations *SituationService
	installed  *briefInstallations
	lookups    *countingChanges
	briefs     *SituationBriefService
	// deployments reports deployments and merges through the changes tests' helpers.
	deployments  *ChangesServiceSuite
	changesSetup *changesFixture
}

// briefInstallations is the organization's installations, each with only its capabilities.
type briefInstallations struct {
	installed []rez.InstalledIntegration
}

func (b *briefInstallations) GetInstalledIntegration(_ context.Context, id uuid.UUID) (rez.InstalledIntegration, error) {
	for _, ii := range b.installed {
		if ii.Integration().ID == id {
			return ii, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (b *briefInstallations) ListAllInstalled(context.Context, ...predicate.Integration) ([]rez.InstalledIntegration, error) {
	return b.installed, nil
}

func (b *briefInstallations) install(provider string, capabilities ...string) {
	intg := &ent.Integration{ID: uuid.New(), Provider: provider, Name: provider}
	b.installed = append(b.installed, &briefInstallation{intg: intg, capabilities: capabilities})
}

type briefInstallation struct {
	intg         *ent.Integration
	capabilities []string
}

func (b *briefInstallation) Integration() *ent.Integration             { return b.intg }
func (b *briefInstallation) Config() rez.IntegrationInstallationConfig { return nil }
func (b *briefInstallation) Capabilities() []string                    { return b.capabilities }

// briefLogReader is an installation that can read logs.
type briefLogReader struct {
	*briefInstallation
}

func newBriefLogReader() *briefLogReader {
	intg := &ent.Integration{ID: uuid.New(), Provider: "provider-c", Name: "provider-c"}
	return &briefLogReader{briefInstallation: &briefInstallation{intg: intg, capabilities: []string{"logs"}}}
}

func (b *briefLogReader) SearchLogs(context.Context, integrations.LogSearchParams) (*integrations.LogSearchResult, error) {
	return &integrations.LogSearchResult{}, nil
}

func (b *briefLogReader) CountLogs(context.Context, integrations.LogCountParams) (*integrations.LogCountResult, error) {
	return &integrations.LogCountResult{}, nil
}

func (b *briefLogReader) CheckLogsAvailable() error {
	return nil
}

// countingChanges records each changes lookup before running it.
type countingChanges struct {
	rez.ChangesService
	lookups []rez.ListServiceChangesParams
}

func (c *countingChanges) ListServiceChanges(ctx context.Context, params rez.ListServiceChangesParams) (*rez.ServiceChanges, error) {
	c.lookups = append(c.lookups, params)
	return c.ChangesService.ListServiceChanges(ctx, params)
}

func (s *SituationBriefServiceSuite) newFixture(tdb rez.Database) *briefFixture {
	h := &briefFixture{
		tdb:         tdb,
		clock:       test.NewClock(briefOpenedAt.Add(2 * time.Hour)),
		installed:   &briefInstallations{},
		deployments: &ChangesServiceSuite{},
	}
	h.deployments.SetT(s.T())
	h.changesSetup = h.deployments.newFixture(tdb)
	h.lookups = &countingChanges{ChangesService: h.changesSetup.changes}

	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{}}, nil).
		Maybe()
	knowledgeQuery, queryErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryErr)
	knowledge, knowledgeErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)
	analyses, analysesErr := NewSystemAnalysisService(tdb, knowledgeQuery)
	s.Require().NoError(analysesErr)
	alerts, alertsErr := NewAlertService(rez.DefaultConfig().Alerts, h.clock, tdb, jobService, knowledge, NewSituationSignalService(jobService))
	s.Require().NoError(alertsErr)
	agentSessions := &AiAgentSessionService{db: tdb, jobs: jobService}
	investigations := NewInvestigationService(tdb, agentSessions, jobService, analyses)
	situations, situationsErr := NewSituationService(tdb, h.clock, jobService, investigations, analyses, knowledgeQuery, nil, alerts)
	s.Require().NoError(situationsErr)
	proposals, proposalsErr := NewKnowledgeProposalService(tdb)
	s.Require().NoError(proposalsErr)

	briefs, briefsErr := NewSituationBriefService(tdb, h.clock, knowledgeQuery, h.lookups, proposals, h.installed)
	s.Require().NoError(briefsErr)

	h.knowledge = knowledge
	h.alerts = alerts
	h.situations = situations
	h.briefs = briefs
	return h
}

// serviceRef is a service as the alerting provider names it, linked to deployments by its normalized name.
func (s *SituationBriefServiceSuite) serviceRef(name string) rez.KnowledgeEntityRef {
	return s.containerRef("service", name)
}

func (s *SituationBriefServiceSuite) containerRef(kind string, name string) rez.KnowledgeEntityRef {
	ref := rez.KnowledgeEntityRef{
		Category: kne.CategoryContainer,
		Kind:     kind,
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "provider-a",
			ProviderNamespace: "alerts",
			ResourceRef:       kind + ":" + name,
		},
	}
	if kind == projections.EntityKindService {
		ref.LinkingAttributes = projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: projections.NormalizeServiceName(name),
		}
	}
	return ref
}

// openSituation records a firing alert at briefOpenedAt observing the entities, and opens a situation with it.
func (s *SituationBriefServiceSuite) openSituation(ctx context.Context, h *briefFixture, entities ...rez.KnowledgeEntityRef) *ent.Situation {
	definition := "alert-" + uuid.NewString()
	createEvent := h.tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("provider-a").
		SetProviderNamespace("alerts").
		SetProviderResourceRef(definition).
		SetProviderEventSource("alerts").
		SetProviderEventRef(uuid.NewString()).
		SetKind("alert").
		SetOccurredAt(briefOpenedAt).
		SetReceivedAt(briefOpenedAt).
		SetAttributes([]byte(`{}`))
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	definitionRef := rez.KnowledgeEntityRef{
		Category: kne.CategorySignal,
		Kind:     "alert",
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "provider-a",
			ProviderNamespace: "alerts",
			ResourceRef:       "alert:" + definition,
		},
	}
	observed := func(entity *rez.KnowledgeEntityRef) rez.KnowledgeEvidenceRef {
		_, name, _ := strings.Cut(entity.ProviderResourceRef.ResourceRef, ":")
		return rez.KnowledgeEvidenceRef{
			Kind:         ke.KindObserved,
			Assertion:    "observed",
			EffectiveAt:  briefOpenedAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: name},
			Subject:      rez.KnowledgeSubjectRef{Entity: entity},
		}
	}
	evidence := []rez.KnowledgeEvidenceRef{observed(&definitionRef)}
	for _, entity := range entities {
		evidence = append(evidence, observed(&entity))
	}
	s.Require().NoError(h.knowledge.IngestEvidence(ctx, event, evidence...))
	definitionAlias, aliasErr := h.knowledge.ResolveInternalSubject(ctx, rez.KnowledgeSubjectRef{Entity: &definitionRef})
	s.Require().NoError(aliasErr)

	record := rez.RecordAlertInstanceParams{
		Event:      event,
		Definition: rez.AlertDefinitionValues{KnowledgeEntityID: *definitionAlias.EntityID, Title: definition},
		Instance: rez.AlertInstanceValues{
			InstanceID: definition,
			Severity:   schematypes.SignalSeverityCritical,
			Firing:     true,
			StartedAt:  briefOpenedAt,
		},
	}
	_, recordErr := h.alerts.RecordAlertInstance(ctx, record)
	s.Require().NoError(recordErr)
	queryEpisode := h.tdb.Client(ctx).AlertEpisode.Query().
		Where(ale.HasInstancesWith(ali.HasEventsWith(aie.EventID(event.ID))))
	episode, episodeErr := queryEpisode.Only(ctx)
	s.Require().NoError(episodeErr)

	signal := *episode.KnowledgeEntityID
	params := rez.CreateSituationParams{
		Title:             "Checkout errors",
		SeedEntityID:      signal,
		ObservationGroups: []rez.SituationObservationGroupParams{{Title: "Checkout", SignalEntityIDs: []uuid.UUID{signal}}},
	}
	created, createErr := h.situations.CreateSituation(ctx, params)
	s.Require().NoError(createErr)
	s.Require().True(briefOpenedAt.Equal(created.OpenedAt))
	return created
}

func (s *SituationBriefServiceSuite) closeAt(ctx context.Context, h *briefFixture, situation *ent.Situation, at time.Time) {
	h.clock.Set(at)
	_, closeErr := h.situations.CloseSituation(ctx, situation.ID, rez.CloseSituationParams{})
	s.Require().NoError(closeErr)
}

func (s *SituationBriefServiceSuite) deploy(ctx context.Context, h *briefFixture, d testDeployment) *ent.KnowledgeEntity {
	return h.deployments.deploy(ctx, h.tdb, h.changesSetup, d)
}

func (s *SituationBriefServiceSuite) brief(ctx context.Context, h *briefFixture, situation *ent.Situation) *rez.SituationBrief {
	brief, briefErr := h.briefs.GetSituationBrief(ctx, situation.ID)
	s.Require().NoError(briefErr)
	return brief
}

// briefDeploymentIDs lists the deployments of the brief's only service, in order.
func (s *SituationBriefServiceSuite) briefDeploymentIDs(brief *rez.SituationBrief) []uuid.UUID {
	s.Require().Len(brief.Changes.Services, 1)
	ids := make([]uuid.UUID, 0)
	for _, deployment := range brief.Changes.Services[0].Deployments {
		ids = append(ids, deployment.Deployment.ID)
	}
	return ids
}

func (s *SituationBriefServiceSuite) TestSituationBriefChanges() {
	s.Run("deployments before and after opening", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		h.deployments.merge(ctx, tdb, h.changesSetup, testMerge{number: 7, mergedAt: changesAt(-40)})
		s.deploy(ctx, h, checkoutDeployment("too-early", changesAt(-25*60)))
		before := s.deploy(ctx, h, checkoutDeployment("before", changesAt(-30)))
		after := s.deploy(ctx, h, checkoutDeployment("after", changesAt(10)))

		brief := s.brief(ctx, h, situation)

		s.True(brief.Changes.Available)
		s.True(changesAt(-24 * 60).Equal(brief.Changes.Start))
		s.True(h.clock.Now().Equal(brief.Changes.End), "an open situation's window ends now")
		s.Equal([]uuid.UUID{after.ID, before.ID}, s.briefDeploymentIDs(brief), "one 25 hours before is not listed")
		listed := brief.Changes.Services[0].Deployments
		s.True(listed[0].AfterOpened)
		s.False(listed[1].AfterOpened)
		s.Require().Len(listed[1].MergedChanges, 1, "merged changes are linked")
		s.Equal(7, listed[1].MergedChanges[0].Number)
		s.Empty(listed[1].MergedChangesUnlinked)
		s.NotNil(listed[1].LatestEvidenceID, "a deployment carries its evidence")
	})

	s.Run("many deployments after opening do not push out the one before", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		before := s.deploy(ctx, h, checkoutDeployment("before", changesAt(-30)))
		for index := range 25 {
			s.deploy(ctx, h, checkoutDeployment("after-"+strconv.Itoa(index), changesAt(index+1)))
		}

		brief := s.brief(ctx, h, situation)

		ids := s.briefDeploymentIDs(brief)
		s.Contains(ids, before.ID)
		s.NotEmpty(brief.Changes.Services[0].Limits, "the cut is said")
	})

	s.Run("a closed situation's window ends when it closed", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))
		s.closeAt(ctx, h, situation, changesAt(60))
		h.clock.Set(changesAt(180))

		during := s.deploy(ctx, h, checkoutDeployment("during", changesAt(30)))
		s.deploy(ctx, h, checkoutDeployment("after-closing", changesAt(90)))

		brief := s.brief(ctx, h, situation)

		s.True(changesAt(60).Equal(brief.Changes.End))
		s.Equal([]uuid.UUID{during.ID}, s.briefDeploymentIDs(brief))
	})

	s.Run("deployments whose merged changes cannot be linked say why", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		withoutRepositoryDeployment := checkoutDeployment("without-repository", changesAt(-20))
		withoutRepositoryDeployment.repository = ""
		s.deploy(ctx, h, withoutRepositoryDeployment)
		unseenDeployment := checkoutDeployment("unseen", changesAt(-10))
		unseenDeployment.repository = "acme/unseen"
		s.deploy(ctx, h, unseenDeployment)

		brief := s.brief(ctx, h, situation)

		listed := brief.Changes.Services[0].Deployments
		s.Require().Len(listed, 2)
		s.Equal(rez.SituationBriefUnlinkedRepositoryNotForge, listed[0].MergedChangesUnlinked)
		s.Equal(rez.SituationBriefUnlinkedNoRepository, listed[1].MergedChangesUnlinked)
	})

	s.Run("a container that is not a service is left out", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"), s.containerRef("host", "node-1"))

		brief := s.brief(ctx, h, situation)

		s.Require().Len(brief.Services, 1)
		s.Equal("checkout-api", brief.Services[0].Entity.State.DisplayName)
		s.Len(h.lookups.lookups, 2, "only the service is looked up")
	})

	s.Run("without a deployments integration nothing is looked up", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		brief := s.brief(ctx, h, situation)

		s.False(brief.Changes.Available)
		s.NotEmpty(brief.Changes.UnavailableReason)
		s.Empty(brief.Changes.Services)
		s.Empty(h.lookups.lookups)
	})

	s.Run("a deployment at the opening is listed once, before it", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		atOpening := s.deploy(ctx, h, checkoutDeployment("at-opening", briefOpenedAt))

		brief := s.brief(ctx, h, situation)

		s.Equal([]uuid.UUID{atOpening.ID}, s.briefDeploymentIDs(brief))
		s.False(brief.Changes.Services[0].Deployments[0].AfterOpened)
	})

	s.Run("a situation closed when it opened runs only the first lookup", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))
		s.closeAt(ctx, h, situation, briefOpenedAt)

		s.brief(ctx, h, situation)

		s.Require().Len(h.lookups.lookups, 1)
		s.True(briefOpenedAt.Equal(h.lookups.lookups[0].End))
	})
}

func (s *SituationBriefServiceSuite) TestSituationBriefWithoutServices() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	h.installed.install("provider-b", "deployments")
	situation := s.openSituation(ctx, h)

	brief := s.brief(ctx, h, situation)

	s.Empty(brief.Services)
	s.True(brief.Changes.Available)
	s.Empty(brief.Changes.Services)
	s.Empty(h.lookups.lookups)
}

func (s *SituationBriefServiceSuite) TestLatestDeploymentBefore() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"), s.serviceRef("payments-api"))

	none, noneErr := h.briefs.LatestDeploymentBefore(ctx, situation.ID)
	s.Require().NoError(noneErr)
	s.Nil(none, "none when there is no deployment")

	s.deploy(ctx, h, checkoutDeployment("checkout", changesAt(-120)))
	failedPayments := checkoutDeployment("payments", changesAt(-30))
	failedPayments.service = "payments-api"
	failedPayments.status = projections.DeploymentStatusFailed
	failedPayments.sha = "abc123"
	s.deploy(ctx, h, failedPayments)
	s.deploy(ctx, h, checkoutDeployment("after-opening", changesAt(5)))

	latest, latestErr := h.briefs.LatestDeploymentBefore(ctx, situation.ID)

	s.Require().NoError(latestErr)
	s.Require().NotNil(latest)
	s.Equal("payments-api", latest.Service.State.DisplayName)
	s.True(changesAt(-30).Equal(latest.DeployedAt))
	s.Equal(projections.DeploymentStatusFailed, latest.Status, "a deployment of any status counts")
	s.Equal("production", latest.Environment)
	s.Equal("abc123", latest.Sha)
}

func (s *SituationBriefServiceSuite) TestSituationBriefUnseen() {
	kinds := func(brief *rez.SituationBrief) []rez.SituationBriefUnseenKind {
		listed := make([]rez.SituationBriefUnseenKind, 0, len(brief.Unseen))
		for _, unseen := range brief.Unseen {
			listed = append(listed, unseen.Kind)
		}
		return listed
	}

	s.Run("nothing installed", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		brief := s.brief(ctx, h, situation)

		s.Equal([]rez.SituationBriefUnseenKind{
			rez.SituationBriefUnseenSourceUnavailable,
			rez.SituationBriefUnseenSourceUnavailable,
			rez.SituationBriefUnseenSourceUnavailable,
			rez.SituationBriefUnseenNotRead,
		}, kinds(brief))
		s.Contains(brief.Unseen[0].Text, "no logs integration installed", "logs say why")
	})

	s.Run("no code changes integration", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments")
		h.installed.installed = append(h.installed.installed, newBriefLogReader())
		situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

		brief := s.brief(ctx, h, situation)

		s.Equal([]rez.SituationBriefUnseenKind{
			rez.SituationBriefUnseenSourceUnavailable,
			rez.SituationBriefUnseenNotRead,
		}, kinds(brief), "only code changes cannot be read")
	})

	s.Run("eleven services", func() {
		ctx, tdb := s.SetupTestDatabase()
		h := s.newFixture(tdb)
		h.installed.install("provider-b", "deployments", "code_changes")
		h.installed.installed = append(h.installed.installed, newBriefLogReader())
		var services []rez.KnowledgeEntityRef
		for index := range 11 {
			services = append(services, s.serviceRef("service-"+strconv.Itoa(index)))
		}
		situation := s.openSituation(ctx, h, services...)

		brief := s.brief(ctx, h, situation)

		s.Len(brief.Services, rez.MaxBriefServices)
		s.Equal([]rez.SituationBriefUnseenKind{
			rez.SituationBriefUnseenServicesCut,
			rez.SituationBriefUnseenNotRead,
		}, kinds(brief))
	})
}

func (s *SituationBriefServiceSuite) TestSituationBriefUnconfirmedMatches() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	situation := s.openSituation(ctx, h, s.serviceRef("checkout-api"))

	// Proposals sort their names, so checkout-api is name_a for the other names but checkout.
	otherNames := map[string]kp.Status{
		"checkout-svc":    kp.StatusConfirmed,
		"checkout-server": kp.StatusRejected,
	}
	for name, status := range otherNames {
		otherRef := s.serviceRef(name)
		otherRef.ProviderResourceRef.Provider = "provider-b"
		s.openSituation(ctx, h, otherRef)
		// Other names are proposed against each other too; only the proposal naming checkout-api matters.
		updateStatus := tdb.Client(ctx).KnowledgeProposal.Update().
			Where(kp.NameA("checkout-api"), kp.NameB(name)).
			SetStatus(status)
		updated, updateErr := updateStatus.Save(ctx)
		s.Require().NoError(updateErr)
		s.Require().Equal(1, updated, "ingesting %s proposes it", name)
	}
	checkoutRef := s.serviceRef("checkout")
	checkoutRef.ProviderResourceRef.Provider = "provider-b"
	s.openSituation(ctx, h, checkoutRef)

	brief := s.brief(ctx, h, situation)

	s.Require().Len(brief.Services, 1)
	matches := brief.Services[0].UnconfirmedMatches
	s.Require().Len(matches, 1, "confirmed and rejected proposals are not listed")
	s.Equal("checkout", matches[0].Name)
	s.Equal("checkout", matches[0].Entity.State.DisplayName)
	s.Contains(matches[0].Rationale, "`api`")
}
