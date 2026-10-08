package eventprojection

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test/mocks"
)

const (
	testWebhookInstallation = "webhook-installation"
	testDeploymentRef       = "deployment:checkout-api-production-run-1"
)

var (
	deploymentStartedAt  = time.Date(2026, 10, 7, 14, 1, 10, 0, time.UTC)
	deploymentFinishedAt = time.Date(2026, 10, 7, 14, 3, 52, 0, time.UTC)
)

// deploymentReportFixture is one report as the webhook integration's deployment preset normalizes it.
type deploymentReportFixture struct {
	status     string
	receivedAt time.Time
	attrs      projections.DeploymentEventAttributes
}

func newDeploymentReport(status string, receivedAt time.Time, service string) deploymentReportFixture {
	serviceName := projections.NormalizeServiceName(service)
	attrs := projections.DeploymentEventAttributes{
		ExternalID: "run-1",
		Status:     status,
		Service: projections.EntityObservation{
			Ref: rez.ProviderResourceRef{
				Provider:          "webhook",
				ProviderNamespace: testWebhookInstallation,
				ResourceRef:       "service:" + serviceName,
			},
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: service,
			LinkingAttributes: projections.LinkingAttributes{
				projections.LinkingAttributeServiceName: serviceName,
			},
		},
		Environment: projections.DeploymentEnvironment{
			Name:        "production",
			DisplayName: "production",
		},
	}
	return deploymentReportFixture{status: status, receivedAt: receivedAt, attrs: attrs}
}

func (r deploymentReportFixture) withRepository(fullName string) deploymentReportFixture {
	r.attrs.Repository = &projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "webhook",
			ProviderNamespace: testWebhookInstallation,
			ResourceRef:       "repository:" + projections.NormalizeRepositoryFullName(fullName),
		},
		Category:          kne.CategoryCode,
		Kind:              knowledgeEntityKindRepository,
		DisplayName:       fullName,
		LinkingAttributes: projections.RepositoryLinkingAttributes(fullName),
	}
	return r
}

// startedReport is checkout-api's deployment starting.
func startedReport() deploymentReportFixture {
	report := newDeploymentReport(projections.DeploymentStatusStarted, deploymentStartedAt, "checkout-api")
	report.attrs.StartedAt = &deploymentStartedAt
	return report
}

// finishedReport is checkout-api's deployment finishing with a status at a time.
func finishedReport(status string, finishedAt time.Time) deploymentReportFixture {
	report := newDeploymentReport(status, finishedAt, "checkout-api")
	report.attrs.FinishedAt = &finishedAt
	return report
}

// deploymentFixtureProcessor stands in for the webhook integration's processor: the provider event's
// attributes are the normalized attributes, and the event's time is the deployment's time.
type deploymentFixtureProcessor struct{}

func (deploymentFixtureProcessor) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var attrs projections.DeploymentEventAttributes
	if decodeErr := json.Unmarshal(prov.Attributes, &attrs); decodeErr != nil {
		return nil, decodeErr
	}
	occurredAt := prov.ReceivedAt
	if attrs.FinishedAt != nil {
		occurredAt = *attrs.FinishedAt
	} else if attrs.StartedAt != nil {
		occurredAt = *attrs.StartedAt
	}
	normalized := &ent.NormalizedEvent{
		Provider:            prov.Provider,
		ProviderNamespace:   prov.ProviderNamespace,
		ProviderResourceRef: testDeploymentRef,
		ProviderEventSource: prov.ProviderEventSource,
		ProviderEventRef:    prov.ProviderEventRef,
		Kind:                projections.KindDeployment,
		OccurredAt:          occurredAt,
		ReceivedAt:          prov.ReceivedAt,
		Attributes:          prov.Attributes,
	}
	return ent.NormalizedEvents{normalized}, nil
}

// deploymentPipeline persists reports through the provider event pipeline, which drops redelivered reports by
// the normalized event's unique index, and projects them when the test chooses.
type deploymentPipeline struct {
	suite    *ProjectionServiceSuite
	pipeline *db.ProviderEventPipelineService
	// queued are the projection jobs the last persisted report queued.
	queued []uuid.UUID
}

func (s *ProjectionServiceSuite) newDeploymentPipeline(tdb rez.Database, service *ProjectionService) *deploymentPipeline {
	p := &deploymentPipeline{suite: s}
	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
			results := make([]*rivertype.JobInsertResult, 0, len(params))
			for _, param := range params {
				projectArgs := param.Args.(jobs.ProjectNormalizedEvent)
				p.queued = append(p.queued, projectArgs.EventId)
				results = append(results, &rivertype.JobInsertResult{Job: &rivertype.JobRow{}})
			}
			return results, nil
		}).
		Maybe()
	processors := map[string]rez.ProviderEventProcessor{
		"webhook": deploymentFixtureProcessor{},
	}
	pipeline, pipelineErr := db.NewProviderEventPipelineService(tdb, jobService, processors, service)
	s.Require().NoError(pipelineErr)
	p.pipeline = pipeline
	return p
}

// persist processes a report's provider event and returns the normalized event it stored, or nil when the
// report was a duplicate.
func (p *deploymentPipeline) persist(ctx context.Context, report deploymentReportFixture) *uuid.UUID {
	encoded, encodeErr := projections.EncodeAttributes(report.attrs)
	p.suite.Require().NoError(encodeErr)
	event := rez.ProviderEvent{
		Provider:            "webhook",
		ProviderNamespace:   testWebhookInstallation,
		ProviderEventSource: "deployment",
		ProviderEventRef:    testDeploymentRef + ":" + report.status,
		Attributes:          encoded,
		ReceivedAt:          report.receivedAt,
	}
	p.queued = nil
	processArgs := &jobs.ProcessProviderEventArgs{Event: event}
	p.suite.Require().NoError(p.pipeline.HandleProcessEventJob(ctx, processArgs))

	if len(p.queued) == 0 {
		return nil
	}
	p.suite.Require().Len(p.queued, 1)
	return &p.queued[0]
}

func (p *deploymentPipeline) project(ctx context.Context, eventID uuid.UUID) {
	projectArgs := jobs.ProjectNormalizedEvent{EventId: eventID}
	p.suite.Require().NoError(p.pipeline.HandleEventProjectionJob(ctx, projectArgs))
}

// receiveDeploymentReports persists and projects reports one at a time, in the order given.
func (s *ProjectionServiceSuite) receiveDeploymentReports(ctx context.Context, tdb rez.Database, service *ProjectionService, reports ...deploymentReportFixture) {
	p := s.newDeploymentPipeline(tdb, service)
	for _, report := range reports {
		eventID := p.persist(ctx, report)
		if eventID != nil {
			p.project(ctx, *eventID)
		}
	}
}

func (s *ProjectionServiceSuite) onlyDeployment(ctx context.Context, tdb rez.Database) *ent.KnowledgeEntity {
	deployments := s.entitiesOfKind(ctx, tdb, kne.CategoryEvent, knowledgeEntityKindDeployment)
	s.Require().Len(deployments, 1)
	return deployments[0]
}

func (s *ProjectionServiceSuite) TestDeploymentStateIsTheFinishedReportInEitherProcessingOrder() {
	started := startedReport().
		withRepository("acme/checkout")
	started.attrs.Sha = "0000000000000000000000000000000000000000"
	started.attrs.Version = "v1.41.0"
	started.attrs.URL = "https://ci.test/runs/1"

	succeeded := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt).
		withRepository("Acme/Checkout")
	// The started report's step ran late, so it is received after the succeeded report.
	started.receivedAt = deploymentFinishedAt.Add(time.Minute)
	succeeded.attrs.StartedAt = &deploymentStartedAt
	succeeded.attrs.Sha = "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392"
	succeeded.attrs.Version = "v1.42.0"

	orders := map[string][]deploymentReportFixture{
		"started first":   {started, succeeded},
		"succeeded first": {succeeded, started},
	}
	for name, order := range orders {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			service := s.projectionService(tdb)

			s.receiveDeploymentReports(ctx, tdb, service, order...)

			deployment := s.onlyDeployment(ctx, tdb)
			s.Equal("checkout-api to production", deployment.State.DisplayName)
			// The url only the started report sent is not kept: a report is complete.
			expectedProperties := map[string]any{
				"environment": "production",
				"status":      "succeeded",
				"external_id": "run-1",
				"repository":  "acme/checkout",
				"sha":         "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392",
				"version":     "v1.42.0",
				"started_at":  "2026-10-07T14:01:10Z",
				"finished_at": "2026-10-07T14:03:52Z",
			}
			s.Equal(expectedProperties, deployment.State.Properties)
			s.Require().NotNil(deployment.StateEffectiveAt)
			s.True(deploymentFinishedAt.Equal(*deployment.StateEffectiveAt), "a finished deployment's time is finished_at")

			// Graph evidence is kept for every report.
			queryReportEvidence := tdb.Client(ctx).KnowledgeEvidence.Query().
				Where(ke.Assertion(knowledgeAssertionDeploymentObserved))
			reportEvidence, countErr := queryReportEvidence.Count(ctx)
			s.Require().NoError(countErr)
			s.Equal(2, reportEvidence)
		})
	}
}

func (s *ProjectionServiceSuite) TestLaterFinishedReportWinsInEitherProcessingOrder() {
	later := deploymentFinishedAt.Add(time.Minute)
	// The report later by deployment time is received first: deployment time decides, not receipt.
	receivedInReverse := func(status string, finishedAt time.Time) deploymentReportFixture {
		report := finishedReport(status, finishedAt)
		report.receivedAt = later.Add(time.Hour).Add(-finishedAt.Sub(deploymentFinishedAt))
		return report
	}

	cases := map[string]struct {
		reports  []deploymentReportFixture
		expected string
	}{
		"failed later, succeeded processed first": {
			reports: []deploymentReportFixture{
				receivedInReverse(projections.DeploymentStatusSucceeded, deploymentFinishedAt),
				receivedInReverse(projections.DeploymentStatusFailed, later),
			},
			expected: projections.DeploymentStatusFailed,
		},
		"failed later, failed processed first": {
			reports: []deploymentReportFixture{
				receivedInReverse(projections.DeploymentStatusFailed, later),
				receivedInReverse(projections.DeploymentStatusSucceeded, deploymentFinishedAt),
			},
			expected: projections.DeploymentStatusFailed,
		},
		"succeeded later, failed processed first": {
			reports: []deploymentReportFixture{
				receivedInReverse(projections.DeploymentStatusFailed, deploymentFinishedAt),
				receivedInReverse(projections.DeploymentStatusSucceeded, later),
			},
			expected: projections.DeploymentStatusSucceeded,
		},
		"succeeded later, succeeded processed first": {
			reports: []deploymentReportFixture{
				receivedInReverse(projections.DeploymentStatusSucceeded, later),
				receivedInReverse(projections.DeploymentStatusFailed, deploymentFinishedAt),
			},
			expected: projections.DeploymentStatusSucceeded,
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			service := s.projectionService(tdb)

			s.receiveDeploymentReports(ctx, tdb, service, tc.reports...)

			deployment := s.onlyDeployment(ctx, tdb)
			s.Equal(tc.expected, deployment.State.Properties["status"])
			s.Require().NotNil(deployment.StateEffectiveAt)
			s.True(later.Equal(*deployment.StateEffectiveAt))
		})
	}
}

func (s *ProjectionServiceSuite) TestReportsAtTheSameTimeGoToTheLastProcessed() {
	succeeded := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt)
	succeeded.attrs.Version = "v1.42.0"
	failed := finishedReport(projections.DeploymentStatusFailed, deploymentFinishedAt)
	failed.attrs.Version = "v1.43.0"

	cases := map[string]struct {
		order           []deploymentReportFixture
		expectedStatus  string
		expectedVersion string
	}{
		"failed processed last":    {order: []deploymentReportFixture{succeeded, failed}, expectedStatus: "failed", expectedVersion: "v1.43.0"},
		"succeeded processed last": {order: []deploymentReportFixture{failed, succeeded}, expectedStatus: "succeeded", expectedVersion: "v1.42.0"},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			service := s.projectionService(tdb)

			s.receiveDeploymentReports(ctx, tdb, service, tc.order...)

			properties := s.onlyDeployment(ctx, tdb).State.Properties
			s.Equal(tc.expectedStatus, properties["status"])
			s.Equal(tc.expectedVersion, properties["version"])
		})
	}
}

func (s *ProjectionServiceSuite) TestRedeliveredReportAddsNothing() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	succeeded := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt).
		withRepository("acme/checkout")
	succeeded.attrs.Version = "v1.42.0"
	s.receiveDeploymentReports(ctx, tdb, service, succeeded)
	before := s.onlyDeployment(ctx, tdb)
	evidenceBefore, beforeErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(beforeErr)

	// A later delivery of a report with the same status is a duplicate, even with other fields changed.
	redelivered := succeeded
	redelivered.receivedAt = deploymentFinishedAt.Add(time.Minute)
	redelivered.attrs.Version = "v1.43.0"
	p := s.newDeploymentPipeline(tdb, service)
	redeliveredID := p.persist(ctx, redelivered)

	s.Nil(redeliveredID, "the pipeline drops the redelivered report")
	after := s.onlyDeployment(ctx, tdb)
	s.Equal(before.State, after.State)
	evidenceAfter, afterErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(afterErr)
	s.Equal(evidenceBefore, evidenceAfter)
}

func (s *ProjectionServiceSuite) TestDeploymentImpactsItsServiceAndTouchesItsRepository() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)

	succeeded := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt).
		withRepository("acme/checkout")
	s.receiveDeploymentReports(ctx, tdb, service, succeeded)

	deployment := s.onlyDeployment(ctx, tdb)
	services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
	repositories := s.entitiesOfKind(ctx, tdb, kne.CategoryCode, knowledgeEntityKindRepository)
	s.Require().Len(services, 1)
	s.Require().Len(repositories, 1)

	impacts := s.relationshipsOf(ctx, tdb, knr.PredicateImpacts)
	s.Require().Len(impacts, 1)
	s.Equal(deployment.ID, impacts[0].SourceEntityID)
	s.Equal(services[0].ID, impacts[0].TargetEntityID)

	touches := s.relationshipsOf(ctx, tdb, knr.PredicateTouches)
	s.Require().Len(touches, 1)
	s.Equal(deployment.ID, touches[0].SourceEntityID)
	s.Equal(repositories[0].ID, touches[0].TargetEntityID)
}

func (s *ProjectionServiceSuite) TestDeploymentAndAlertObserveOneService() {
	deployment := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt)

	s.Run("alert first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.projectServiceAlert(ctx, tdb, service, "alertmanager-installation", "Checkout_API")
		s.receiveDeploymentReports(ctx, tdb, service, deployment)

		s.assertDeploymentAndAlertShareService(ctx, tdb)
	})

	s.Run("deployment first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.receiveDeploymentReports(ctx, tdb, service, deployment)
		s.projectServiceAlert(ctx, tdb, service, "alertmanager-installation", "Checkout_API")

		s.assertDeploymentAndAlertShareService(ctx, tdb)
	})
}

func (s *ProjectionServiceSuite) assertDeploymentAndAlertShareService(ctx context.Context, tdb rez.Database) {
	services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
	s.Require().Len(services, 1)
	observes := s.relationshipsOf(ctx, tdb, knr.PredicateObserves)
	s.Require().Len(observes, 1)
	s.Equal(services[0].ID, observes[0].TargetEntityID)
	impacts := s.relationshipsOf(ctx, tdb, knr.PredicateImpacts)
	s.Require().Len(impacts, 1)
	s.Equal(services[0].ID, impacts[0].TargetEntityID)
}

func (s *ProjectionServiceSuite) TestDeploymentAndGithubResolveToOneRepository() {
	deployment := finishedReport(projections.DeploymentStatusSucceeded, deploymentFinishedAt).
		withRepository("acme/checkout")
	githubCheckout := githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout")

	s.Run("deployment first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.receiveDeploymentReports(ctx, tdb, service, deployment)
		s.Require().NoError(s.projectGithubChange(ctx, tdb, service, "change:42:abc123", githubCheckout))

		s.assertDeploymentAndGithubShareRepository(ctx, tdb)
	})

	s.Run("GitHub first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.Require().NoError(s.projectGithubChange(ctx, tdb, service, "change:42:abc123", githubCheckout))
		s.receiveDeploymentReports(ctx, tdb, service, deployment)

		s.assertDeploymentAndGithubShareRepository(ctx, tdb)
	})
}

func (s *ProjectionServiceSuite) assertDeploymentAndGithubShareRepository(ctx context.Context, tdb rez.Database) {
	repositories := s.entitiesOfKind(ctx, tdb, kne.CategoryCode, knowledgeEntityKindRepository)
	s.Require().Len(repositories, 1)
	touches := s.relationshipsOf(ctx, tdb, knr.PredicateTouches)
	s.Require().Len(touches, 2, "the GitHub change and the deployment touch the repository")
	for _, relationship := range touches {
		s.Equal(repositories[0].ID, relationship.TargetEntityID)
	}
}
