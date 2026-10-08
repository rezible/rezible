package db

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/internal/db/eventprojection"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type ChangesServiceSuite struct {
	test.Suite
}

func TestServiceChanges(t *testing.T) {
	suite.Run(t, &ChangesServiceSuite{Suite: test.NewSuite()})
}

var changesTestStart = time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

// changesAt is `minutes` after changesTestStart.
func changesAt(minutes int) time.Time {
	return changesTestStart.Add(time.Duration(minutes) * time.Minute)
}

type changesFixture struct {
	projection *eventprojection.ProjectionService
	changes    *ChangesService
}

func (s *ChangesServiceSuite) newFixture(tdb rez.Database) *changesFixture {
	knowledge, knowledgeErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeErr)

	knowledgeQuery, knowledgeQueryErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(knowledgeQueryErr)

	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{}}, nil).
		Maybe()
	clock := test.NewClock(changesTestStart)
	alerts, alertsErr := NewAlertService(rez.DefaultConfig().Alerts, clock, tdb, jobService, knowledge, NewSituationSignalService(jobService))
	s.Require().NoError(alertsErr)

	projection, projectionErr := eventprojection.NewProjectionService(tdb, knowledge, knowledgeQuery, nil, nil, alerts)
	s.Require().NoError(projectionErr)

	changes, changesErr := NewChangesService(tdb, knowledgeQuery)
	s.Require().NoError(changesErr)

	return &changesFixture{projection: projection, changes: changes}
}

// project records a normalized event and projects it, as the provider event pipeline does.
func (s *ChangesServiceSuite) project(ctx context.Context, tdb rez.Database, h *changesFixture, event *ent.NormalizedEvent) {
	createEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider(event.Provider).
		SetProviderNamespace(event.ProviderNamespace).
		SetProviderResourceRef(event.ProviderResourceRef).
		SetProviderEventSource(event.ProviderEventSource).
		SetProviderEventRef(uuid.NewString()).
		SetKind(event.Kind).
		SetOccurredAt(event.OccurredAt).
		SetReceivedAt(event.OccurredAt).
		SetAttributes(event.Attributes)
	saved, saveErr := createEvent.Save(ctx)
	s.Require().NoError(saveErr)

	projector, ok := h.projection.GetEventProjectorFunc(saved)
	s.Require().True(ok)
	_, projectErr := projector(ctx, saved)
	s.Require().NoError(projectErr)
}

// testDeployment is a deployment as the webhook integration's deployment preset reports it.
type testDeployment struct {
	id          string
	service     string
	environment string
	status      string
	repository  string
	sha         string
	at          time.Time
}

// deploy reports the deployment and returns its entity.
func (s *ChangesServiceSuite) deploy(ctx context.Context, tdb rez.Database, h *changesFixture, d testDeployment) *ent.KnowledgeEntity {
	serviceName := projections.NormalizeServiceName(d.service)
	environmentName := projections.NormalizeServiceName(d.environment)
	attrs := projections.DeploymentEventAttributes{
		ExternalID: d.id,
		Status:     d.status,
		Service: projections.EntityObservation{
			Ref: rez.ProviderResourceRef{
				Provider:          "webhook",
				ProviderNamespace: "webhook-installation",
				ResourceRef:       "service:" + serviceName,
			},
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: d.service,
			LinkingAttributes: projections.LinkingAttributes{
				projections.LinkingAttributeServiceName: serviceName,
			},
		},
		Environment: projections.DeploymentEnvironment{
			Name:        environmentName,
			DisplayName: d.environment,
		},
		Sha: d.sha,
	}
	if d.repository != "" {
		attrs.Repository = &projections.EntityObservation{
			Ref: rez.ProviderResourceRef{
				Provider:          "webhook",
				ProviderNamespace: "webhook-installation",
				ResourceRef:       "repository:" + projections.NormalizeRepositoryFullName(d.repository),
			},
			Category:          kne.CategoryCode,
			Kind:              "repository",
			DisplayName:       d.repository,
			LinkingAttributes: projections.RepositoryLinkingAttributes(d.repository),
		}
	}
	if attrs.Finished() {
		attrs.FinishedAt = &d.at
	} else {
		attrs.StartedAt = &d.at
	}
	event := &ent.NormalizedEvent{
		Provider:            "webhook",
		ProviderNamespace:   "webhook-installation",
		ProviderResourceRef: fmt.Sprintf("deployment:%s:%s:%s", serviceName, environmentName, d.id),
		ProviderEventSource: "deployments",
		Kind:                projections.KindDeployment,
		OccurredAt:          d.at,
	}
	event.Attributes = s.encode(attrs)

	s.project(ctx, tdb, h, event)

	queryAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider("webhook"),
			ksa.ProviderResourceRef(event.ProviderResourceRef),
		)
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)
	deployment, getErr := tdb.Client(ctx).KnowledgeEntity.Get(ctx, *alias.EntityID)
	s.Require().NoError(getErr)
	return deployment
}

// checkoutDeployment is a succeeded production deployment of checkout-api from acme/checkout.
func checkoutDeployment(id string, at time.Time) testDeployment {
	return testDeployment{
		id:          id,
		service:     "checkout-api",
		environment: "production",
		status:      projections.DeploymentStatusSucceeded,
		repository:  "acme/checkout",
		at:          at,
	}
}

// testMerge is a change GitHub reports merged into a repository, Acme/Checkout unless named, whose default
// branch is main.
type testMerge struct {
	repository string
	number     int
	baseRef    string
	sha        string
	mergedAt   time.Time
}

func (s *ChangesServiceSuite) merge(ctx context.Context, tdb rez.Database, h *changesFixture, m testMerge) {
	baseRef := m.baseRef
	if baseRef == "" {
		baseRef = "main"
	}
	sha := m.sha
	if sha == "" {
		sha = fmt.Sprintf("%040d", m.number)
	}
	repository := m.repository
	if repository == "" {
		repository = "Acme/Checkout"
	}
	repositoryRef := "repository:" + projections.NormalizeRepositoryFullName(repository)
	attrs := projections.CodeChangeEventAttributes{
		Repository: projections.EntityObservation{
			Ref: rez.ProviderResourceRef{
				Provider:          "github",
				ProviderNamespace: "github-installation",
				ResourceRef:       repositoryRef,
			},
			Category:          kne.CategoryCode,
			Kind:              "repository",
			DisplayName:       repository,
			LinkingAttributes: projections.RepositoryLinkingAttributes(repository),
		},
		DisplayName: "Change " + strconv.Itoa(m.number),
		Merge: &projections.CodeChangeMerge{
			MergedAt:          m.mergedAt,
			MergeCommitSha:    sha,
			BaseRef:           baseRef,
			IntoDefaultBranch: baseRef == "main",
			Number:            m.number,
		},
	}
	event := &ent.NormalizedEvent{
		Provider:            "github",
		ProviderNamespace:   "github-installation",
		ProviderResourceRef: fmt.Sprintf("change:%s:pr:%d", repositoryRef, m.number),
		ProviderEventSource: "pull_request",
		Kind:                projections.KindCodeChange,
		OccurredAt:          m.mergedAt,
	}
	event.Attributes = s.encode(attrs)

	s.project(ctx, tdb, h, event)
}

func (s *ChangesServiceSuite) encode(attrs any) []byte {
	encoded, encodeErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(encodeErr)
	return encoded
}

// serviceEntityID is the entity the webhook installation's observation of a service resolves to.
func (s *ChangesServiceSuite) serviceEntityID(ctx context.Context, tdb rez.Database, service string) uuid.UUID {
	queryAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider("webhook"),
			ksa.ProviderResourceRef("service:"+projections.NormalizeServiceName(service)),
		)
	alias, queryErr := queryAlias.Only(ctx)
	s.Require().NoError(queryErr)
	return *alias.EntityID
}

func (s *ChangesServiceSuite) list(ctx context.Context, h *changesFixture, params rez.ListServiceChangesParams) *rez.ServiceChanges {
	changes, listErr := h.changes.ListServiceChanges(ctx, params)
	s.Require().NoError(listErr)
	return changes
}

func deploymentIDs(changes *rez.ServiceChanges) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(changes.Deployments))
	for _, listed := range changes.Deployments {
		ids = append(ids, listed.Deployment.ID)
	}
	return ids
}

// linkedNumbers lists a deployment's merged change numbers with how each was linked, in order.
func linkedNumbers(listed rez.ServiceDeployment) []string {
	linked := make([]string, 0, len(listed.MergedChanges))
	for _, change := range listed.MergedChanges {
		linked = append(linked, fmt.Sprintf("#%d %s", change.Number, change.Link))
	}
	return linked
}

func (s *ChangesServiceSuite) TestWindowServiceAndEnvironment() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	atStart := s.deploy(ctx, tdb, h, checkoutDeployment("at-start", changesAt(0)))
	failedDeployment := checkoutDeployment("production", changesAt(10))
	failedDeployment.status = projections.DeploymentStatusFailed
	production := s.deploy(ctx, tdb, h, failedDeployment)
	stagingDeployment := checkoutDeployment("staging", changesAt(20))
	stagingDeployment.environment = "staging"
	stagingDeployment.status = projections.DeploymentStatusStarted
	staging := s.deploy(ctx, tdb, h, stagingDeployment)
	atEnd := s.deploy(ctx, tdb, h, checkoutDeployment("at-end", changesAt(60)))
	s.deploy(ctx, tdb, h, checkoutDeployment("before", changesAt(-1)))
	s.deploy(ctx, tdb, h, checkoutDeployment("after", changesAt(61)))
	paymentsDeployment := checkoutDeployment("payments", changesAt(30))
	paymentsDeployment.service = "payments-api"
	s.deploy(ctx, tdb, h, paymentsDeployment)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	everyEnvironment := s.list(ctx, h, params)
	s.Equal([]uuid.UUID{atEnd.ID, staging.ID, production.ID, atStart.ID}, deploymentIDs(everyEnvironment), "deployments of any status are listed")
	s.Empty(everyEnvironment.Limits)

	params.Environment = "Production"
	productionOnly := s.list(ctx, h, params)
	s.Equal([]uuid.UUID{atEnd.ID, production.ID, atStart.ID}, deploymentIDs(productionOnly))
}

func (s *ChangesServiceSuite) TestTiesOrderStably() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	first := s.deploy(ctx, tdb, h, checkoutDeployment("first", changesAt(10)))
	second := s.deploy(ctx, tdb, h, checkoutDeployment("second", changesAt(10)))
	expected := []uuid.UUID{first.ID, second.ID}
	if second.ID.String() < first.ID.String() {
		expected = []uuid.UUID{second.ID, first.ID}
	}

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	for range 3 {
		s.Equal(expected, deploymentIDs(s.list(ctx, h, params)))
	}
}

func (s *ChangesServiceSuite) TestMergeCommitLink() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	mergeCommit := "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392"

	s.merge(ctx, tdb, h, testMerge{number: 7, sha: mergeCommit, mergedAt: changesAt(-10)})
	s.merge(ctx, tdb, h, testMerge{number: 8, mergedAt: changesAt(-5)})
	otherRepositoryMatch := testMerge{
		repository: "Acme/Payments",
		number:     20,
		sha:        mergeCommit,
		mergedAt:   changesAt(-30 * 24 * 60),
	}
	s.merge(ctx, tdb, h, otherRepositoryMatch)
	otherRepositoryInRange := testMerge{
		repository: "Acme/Payments",
		number:     21,
		mergedAt:   changesAt(-3),
	}
	s.merge(ctx, tdb, h, otherRepositoryInRange)
	deployment := checkoutDeployment("run-1", changesAt(10))
	deployment.sha = mergeCommit
	s.deploy(ctx, tdb, h, deployment)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.Equal("acme/checkout", listed.RepositoryName, "the latest observation names the repository")
	s.True(listed.RepositorySeenByCodeForge)
	expected := []string{
		"#7 merge_commit",
		"#8 merged_in_lookback",
	}
	s.Equal(expected, linkedNumbers(listed), "the merge commit match is listed once, first, and only the deployment's repository is matched")

	match := listed.MergedChanges[0]
	s.Equal("Change 7", match.Change.State.DisplayName)
	s.True(changesAt(-10).Equal(match.MergedAt))
	s.Equal(mergeCommit, match.MergeCommitSha)
	s.Equal("main", match.BaseRef)
	s.True(match.IntoDefaultBranch)
}

func (s *ChangesServiceSuite) TestMergedSincePreviousDeployment() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.deploy(ctx, tdb, h, checkoutDeployment("older", changesAt(-30)))
	s.deploy(ctx, tdb, h, checkoutDeployment("previous", changesAt(0)))
	paymentsDeployment := checkoutDeployment("payments", changesAt(50))
	paymentsDeployment.service = "payments-api"
	s.deploy(ctx, tdb, h, paymentsDeployment)
	stagingDeployment := checkoutDeployment("staging", changesAt(15))
	stagingDeployment.environment = "staging"
	s.deploy(ctx, tdb, h, stagingDeployment)
	otherRepositoryDeployment := checkoutDeployment("other-repository", changesAt(16))
	otherRepositoryDeployment.repository = "acme/checkout-worker"
	s.deploy(ctx, tdb, h, otherRepositoryDeployment)
	failedDeployment := checkoutDeployment("failed", changesAt(20))
	failedDeployment.status = projections.DeploymentStatusFailed
	s.deploy(ctx, tdb, h, failedDeployment)
	s.deploy(ctx, tdb, h, checkoutDeployment("this", changesAt(60)))

	s.merge(ctx, tdb, h, testMerge{number: 7, mergedAt: changesAt(-10)})
	s.merge(ctx, tdb, h, testMerge{number: 1, mergedAt: changesAt(0)})
	s.merge(ctx, tdb, h, testMerge{number: 2, mergedAt: changesAt(10)})
	s.merge(ctx, tdb, h, testMerge{number: 3, mergedAt: changesAt(25)})
	s.merge(ctx, tdb, h, testMerge{number: 4, mergedAt: changesAt(60)})
	s.merge(ctx, tdb, h, testMerge{number: 5, mergedAt: changesAt(61)})
	s.merge(ctx, tdb, h, testMerge{number: 6, baseRef: "release-1", mergedAt: changesAt(30)})

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(30),
		End:             changesAt(90),
		Environment:     "production",
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.Require().NotNil(listed.PreviousDeployedAt)
	s.True(changesAt(0).Equal(*listed.PreviousDeployedAt), "the latest earlier success is previous; a failed deployment, or one of another service, environment or repository, is not")
	expected := []string{
		"#4 merged_since_previous_deployment",
		"#3 merged_since_previous_deployment",
		"#2 merged_since_previous_deployment",
	}
	s.Equal(expected, linkedNumbers(listed), "merges in (previous, this], into the default branch only")
	s.False(listed.MergedChangesCut)
}

func (s *ChangesServiceSuite) TestMergedInLookback() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	lookback := 7 * 24 * time.Hour

	s.deploy(ctx, tdb, h, checkoutDeployment("first", changesAt(0)))
	s.merge(ctx, tdb, h, testMerge{number: 1, mergedAt: changesAt(0).Add(-lookback)})
	s.merge(ctx, tdb, h, testMerge{number: 2, mergedAt: changesAt(1).Add(-lookback)})
	s.merge(ctx, tdb, h, testMerge{number: 3, mergedAt: changesAt(0)})

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(-60),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.Nil(listed.PreviousDeployedAt)
	expected := []string{
		"#3 merged_in_lookback",
		"#2 merged_in_lookback",
	}
	s.Equal(expected, linkedNumbers(listed))
}

func (s *ChangesServiceSuite) TestDeploymentsWithoutMergedChanges() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.merge(ctx, tdb, h, testMerge{number: 1, mergedAt: changesAt(-5)})
	withoutRepositoryDeployment := checkoutDeployment("without-repository", changesAt(10))
	withoutRepositoryDeployment.repository = ""
	withoutRepository := s.deploy(ctx, tdb, h, withoutRepositoryDeployment)
	unseenDeployment := checkoutDeployment("unseen", changesAt(20))
	unseenDeployment.repository = "acme/unseen"
	unseen := s.deploy(ctx, tdb, h, unseenDeployment)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Equal([]uuid.UUID{unseen.ID, withoutRepository.ID}, deploymentIDs(changes))

	unseenListed := changes.Deployments[0]
	s.Equal("acme/unseen", unseenListed.RepositoryName)
	s.False(unseenListed.RepositorySeenByCodeForge)
	s.Empty(unseenListed.MergedChanges)

	withoutRepositoryListed := changes.Deployments[1]
	s.Nil(withoutRepositoryListed.RepositoryEntityID)
	s.Empty(withoutRepositoryListed.RepositoryName)
	s.Empty(withoutRepositoryListed.MergedChanges)
}

func (s *ChangesServiceSuite) TestPreviousDeploymentWithoutRepository() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	earlier := checkoutDeployment("earlier", changesAt(0))
	earlier.repository = ""
	s.deploy(ctx, tdb, h, earlier)
	s.deploy(ctx, tdb, h, checkoutDeployment("with-repository", changesAt(10)))
	later := checkoutDeployment("later", changesAt(20))
	later.repository = ""
	s.deploy(ctx, tdb, h, later)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(15),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.Nil(listed.RepositoryEntityID)
	s.Require().NotNil(listed.PreviousDeployedAt)
	s.True(changesAt(0).Equal(*listed.PreviousDeployedAt), "the previous deployment also names no repository")
}

func (s *ChangesServiceSuite) TestDeploymentWhoseRepositoryChangedBetweenReports() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	s.merge(ctx, tdb, h, testMerge{number: 1, mergedAt: changesAt(-5)})
	started := checkoutDeployment("run-1", changesAt(0))
	started.status = projections.DeploymentStatusStarted
	started.repository = "acme/checkout-old"
	s.deploy(ctx, tdb, h, started)
	deployment := s.deploy(ctx, tdb, h, checkoutDeployment("run-1", changesAt(1)))

	queryTouches := tdb.Client(ctx).KnowledgeRelationship.Query().
		Where(
			knr.SourceEntityID(deployment.ID),
			knr.PredicateEQ(knr.PredicateTouches),
		)
	touches, touchesErr := queryTouches.Count(ctx)
	s.Require().NoError(touchesErr)
	s.Require().Equal(2, touches, "the deployment touches both reported repositories")

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.Equal("acme/checkout", listed.RepositoryName, "the repository the deployment's state names")
	s.Equal([]string{"#1 merged_in_lookback"}, linkedNumbers(listed))
}

func (s *ChangesServiceSuite) TestInvalidRequests() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	s.deploy(ctx, tdb, h, checkoutDeployment("run-1", changesAt(0)))
	serviceID := s.serviceEntityID(ctx, tdb, "checkout-api")
	queryRepository := tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.CategoryEQ(kne.CategoryCode), kne.Kind("repository"))
	repository, repositoryErr := queryRepository.Only(ctx)
	s.Require().NoError(repositoryErr)

	sevenDays := rez.ListServiceChangesParams{
		ServiceEntityID: serviceID,
		Start:           changesAt(0),
		End:             changesAt(0).Add(7 * 24 * time.Hour),
	}
	_, sevenDaysErr := h.changes.ListServiceChanges(ctx, sevenDays)
	s.NoError(sevenDaysErr, "a window of exactly 7 days is allowed")

	overSevenDays := sevenDays
	overSevenDays.End = sevenDays.End.Add(time.Second)
	_, overErr := h.changes.ListServiceChanges(ctx, overSevenDays)
	s.ErrorIs(overErr, rez.ErrInvalidInput)
	s.ErrorContains(overErr, "7 day limit")

	invalid := map[string]rez.ListServiceChangesParams{
		"end equal to start": {
			ServiceEntityID: serviceID,
			Start:           changesAt(0),
			End:             changesAt(0),
		},
		"end before start": {
			ServiceEntityID: serviceID,
			Start:           changesAt(10),
			End:             changesAt(0),
		},
		"an environment normalizing to nothing": {
			ServiceEntityID: serviceID,
			Start:           changesAt(0),
			End:             changesAt(60),
			Environment:     " -- ",
		},
		"an unknown entity": {
			ServiceEntityID: uuid.New(),
			Start:           changesAt(0),
			End:             changesAt(60),
		},
		"an entity that is not a service": {
			ServiceEntityID: repository.ID,
			Start:           changesAt(0),
			End:             changesAt(60),
		},
	}
	for name, params := range invalid {
		_, listErr := h.changes.ListServiceChanges(ctx, params)
		s.ErrorIs(listErr, rez.ErrInvalidInput, name)
	}
}

func (s *ChangesServiceSuite) TestDeploymentLimit() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	for i := range 20 {
		s.deploy(ctx, tdb, h, checkoutDeployment("run-"+strconv.Itoa(i), changesAt(i)))
	}
	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	atTheLimit := s.list(ctx, h, params)

	s.Len(atTheLimit.Deployments, 20)
	s.Empty(atTheLimit.Limits, "exactly 20 deployments are not cut")

	s.deploy(ctx, tdb, h, checkoutDeployment("run-20", changesAt(20)))
	changes := s.list(ctx, h, params)

	s.Len(changes.Deployments, 20)
	s.True(changesAt(20).Equal(changes.Deployments[0].DeployedAt), "the most recent are kept")
	s.True(changesAt(1).Equal(changes.Deployments[19].DeployedAt))
	s.Len(changes.Limits, 1)
}

func (s *ChangesServiceSuite) TestMergedChangeLimitKeepsMergeCommitMatches() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	mergeCommit := "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392"

	oldMatch := testMerge{
		number:   100,
		baseRef:  "release-1",
		sha:      mergeCommit,
		mergedAt: changesAt(-30 * 24 * 60),
	}
	s.merge(ctx, tdb, h, oldMatch)
	for number := 1; number <= 21; number++ {
		s.merge(ctx, tdb, h, testMerge{number: number, mergedAt: changesAt(-number)})
	}
	deployment := checkoutDeployment("run-1", changesAt(0))
	deployment.sha = mergeCommit
	s.deploy(ctx, tdb, h, deployment)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.True(listed.MergedChangesCut)
	s.Empty(changes.Limits, "a merged change cut is recorded on the deployment")
	linked := linkedNumbers(listed)
	s.Require().Len(linked, 20)
	s.Equal("#100 merge_commit", linked[0], "an old merge commit match into another branch is kept")
	s.Equal("#1 merged_in_lookback", linked[1])
	s.Equal("#19 merged_in_lookback", linked[19])
}

func (s *ChangesServiceSuite) TestMergedChangeLimitCountsAMatchInTheRangeOnce() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	mergeCommit := "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392"

	s.merge(ctx, tdb, h, testMerge{number: 100, sha: mergeCommit, mergedAt: changesAt(-1)})
	for number := 1; number <= 19; number++ {
		s.merge(ctx, tdb, h, testMerge{number: number, mergedAt: changesAt(-1 - number)})
	}
	deployment := checkoutDeployment("run-1", changesAt(0))
	deployment.sha = mergeCommit
	s.deploy(ctx, tdb, h, deployment)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	atTheLimit := s.list(ctx, h, params)

	s.Require().Len(atTheLimit.Deployments, 1)
	atTheLimitLinked := linkedNumbers(atTheLimit.Deployments[0])
	s.Require().Len(atTheLimitLinked, 20)
	s.Equal("#100 merge_commit", atTheLimitLinked[0])
	s.Equal("#19 merged_in_lookback", atTheLimitLinked[19])
	s.False(atTheLimit.Deployments[0].MergedChangesCut, "the match and 19 other merges are 20 changes")

	s.merge(ctx, tdb, h, testMerge{number: 20, mergedAt: changesAt(-21)})
	overTheLimit := s.list(ctx, h, params)

	s.Require().Len(overTheLimit.Deployments, 1)
	overTheLimitLinked := linkedNumbers(overTheLimit.Deployments[0])
	s.Require().Len(overTheLimitLinked, 20)
	s.Equal("#100 merge_commit", overTheLimitLinked[0])
	s.Equal("#19 merged_in_lookback", overTheLimitLinked[19])
	s.True(overTheLimit.Deployments[0].MergedChangesCut, "the match and 20 other merges are 21 changes")
}

func (s *ChangesServiceSuite) TestMergedChangeTiesOrderByEntityID() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	for number := 1; number <= 22; number++ {
		s.merge(ctx, tdb, h, testMerge{number: number, mergedAt: changesAt(-5)})
	}
	s.deploy(ctx, tdb, h, checkoutDeployment("run-1", changesAt(0)))

	queryChanges := tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.Kind("code_change")).
		Order(kne.ByID())
	changeIDs, idsErr := queryChanges.IDs(ctx)
	s.Require().NoError(idsErr)
	s.Require().Len(changeIDs, 22)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: s.serviceEntityID(ctx, tdb, "checkout-api"),
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Require().Len(changes.Deployments, 1)
	listed := changes.Deployments[0]
	s.True(listed.MergedChangesCut)
	linkedIDs := make([]uuid.UUID, 0, len(listed.MergedChanges))
	for _, change := range listed.MergedChanges {
		linkedIDs = append(linkedIDs, change.Change.ID)
	}
	s.Equal(changeIDs[:20], linkedIDs, "merges at the same time are ordered by entity ID, and the last are cut")
}

func (s *ChangesServiceSuite) TestReportedDeploymentSharesAlertedService() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	alertedService := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "alertmanager",
			ProviderNamespace: "alertmanager-installation",
			ResourceRef:       "service:checkout-api",
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: "Checkout_API",
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: "checkout-api",
		},
	}
	alert := projections.AlertInstanceEventAttributes{
		Title:            "HighErrorRate",
		State:            projections.AlertStateFiring,
		Severity:         "critical",
		StartedAt:        changesAt(30),
		ObservedEntities: []projections.EntityObservation{alertedService},
	}
	alertEvent := &ent.NormalizedEvent{
		Provider:            "alertmanager",
		ProviderNamespace:   "alertmanager-installation",
		ProviderResourceRef: "HighErrorRate",
		ProviderEventSource: "alerts",
		Kind:                projections.KindAlertInstance,
		OccurredAt:          changesAt(30),
	}
	alertEvent.Attributes = s.encode(alert)
	s.project(ctx, tdb, h, alertEvent)

	reported := s.deploy(ctx, tdb, h, checkoutDeployment("run-1", changesAt(10)))

	queryAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider("alertmanager"),
			ksa.ProviderResourceRef("service:checkout-api"),
		)
	alertedAlias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)

	params := rez.ListServiceChangesParams{
		ServiceEntityID: *alertedAlias.EntityID,
		Start:           changesAt(0),
		End:             changesAt(60),
	}
	changes := s.list(ctx, h, params)

	s.Equal([]uuid.UUID{reported.ID}, deploymentIDs(changes))
}
