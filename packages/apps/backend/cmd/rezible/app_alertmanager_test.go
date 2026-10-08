package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ald "github.com/rezible/rezible/ent/alertdefinition"
	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	rezhttp "github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/pkg/jobs"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

// alertmanagerJourney drives the application through the Alertmanager webhook, as a receiver would.
type alertmanagerJourney struct {
	*appHarness
	ctx     context.Context
	owner   test.Identity
	handler http.Handler
}

func (s *BackendSuite) newAlertmanagerJourney() *alertmanagerJourney {
	ah := s.newAppHarness(appTestOptions{ModelAction: incidentalInvestigationModel})
	ctx, owner := ah.NewIdentity("Alertmanager admin")
	server, serverErr := ah.app.invoke[*rezhttp.Server]()
	s.Require().NoError(serverErr, "resolve application HTTP server")
	return &alertmanagerJourney{appHarness: ah, ctx: ctx, owner: owner, handler: server.Handler()}
}

// install installs Alertmanager with client config, which installing ignores.
func (j *alertmanagerJourney) install() oapiv1.IntegrationInstallation {
	install := j.API(j.owner).Operation[oapiv1.InstallIntegrationRequest, oapiv1.InstallIntegrationResponse](oapiv1.InstallIntegration)
	request := oapiv1.InstallIntegrationRequest{Name: "alertmanager"}
	request.Body.Attributes.Config = map[string]any{"installationRef": "client-supplied"}
	return install.Call(j.suite.T().Context(), request).Body.Data
}

// issueWebhookPath issues a webhook URL through the full server and returns the URL's path.
func (j *alertmanagerJourney) issueWebhookPath(id uuid.UUID) string {
	cookies := httptest.NewRecorder()
	j.cookie.Set(cookies, j.owner.Session)
	basePath := j.app.mustInvoke[rez.Config]().HttpServer.BasePath
	request := httptest.NewRequest(http.MethodPost, basePath+"/v1/integrations/installations/"+id.String()+"/webhook-token", nil)
	for _, cookie := range cookies.Result().Cookies() {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	j.handler.ServeHTTP(recorder, request)
	j.suite.Require().Equal(http.StatusOK, recorder.Code, recorder.Body.String())
	j.suite.Equal("no-store", recorder.Header().Get("Cache-Control"))

	var response oapiv1.IssueIntegrationWebhookTokenResponse
	j.suite.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &response.Body))
	webhookUrl, parseErr := url.Parse(response.Body.Data.Url)
	j.suite.Require().NoError(parseErr)
	j.suite.Equal("https", webhookUrl.Scheme)
	j.suite.Equal(j.app.mustInvoke[rez.Config]().App.ApiDomain, webhookUrl.Host)
	return webhookUrl.Path
}

// deliver posts a webhook delivery and, once accepted, waits for its events to be processed and projected.
func (j *alertmanagerJourney) deliver(path string, status int, alerts ...map[string]any) {
	j.suite.T().Helper()
	payload := map[string]any{"version": "4", "status": "firing", "truncatedAlerts": 0, "alerts": alerts}
	body, encodeErr := json.Marshal(payload)
	j.suite.Require().NoError(encodeErr)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	j.handler.ServeHTTP(recorder, request)
	j.suite.Require().Equal(status, recorder.Code, recorder.Body.String())
	j.awaitProviderEvents()
}

// awaitProviderEvents waits for every process job, then every projection job they queued.
func (j *alertmanagerJourney) awaitProviderEvents() {
	for _, kind := range []string{(&jobs.ProcessProviderEventArgs{}).Kind(), jobs.ProjectNormalizedEvent{}.Kind()} {
		rows, listErr := j.jobsOfKind(j.ctx, kind)
		j.suite.Require().NoError(listErr)
		for _, row := range rows {
			j.AwaitJob(j.ctx, row.ID)
		}
	}
}

func (j *alertmanagerJourney) countEvents(installation uuid.UUID) int {
	query := j.Client(j.ctx).NormalizedEvent.Query().
		Where(ne.Provider("alertmanager"), ne.ProviderNamespace(installation.String()))
	count, countErr := query.Count(j.ctx)
	j.suite.Require().NoError(countErr)
	return count
}

func (j *alertmanagerJourney) countAliases(installation uuid.UUID, refPrefix string) int {
	query := j.Client(j.ctx).KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider("alertmanager"),
			ksa.ProviderNamespace(installation.String()),
			ksa.ProviderResourceRefHasPrefix(refPrefix),
		)
	count, countErr := query.Count(j.ctx)
	j.suite.Require().NoError(countErr)
	return count
}

func (j *alertmanagerJourney) windows(fingerprint string) []*ent.AlertInstance {
	query := j.Client(j.ctx).AlertInstance.Query().
		Where(ali.InstanceKey(fingerprint))
	windows, queryErr := query.All(j.ctx)
	j.suite.Require().NoError(queryErr)
	return windows
}

func alertmanagerAlert(fingerprint, status string, labels map[string]string, startsAt time.Time, endsAt *time.Time) map[string]any {
	alert := map[string]any{
		"status":       status,
		"labels":       labels,
		"annotations":  map[string]string{"summary": labels["alertname"] + " firing"},
		"startsAt":     startsAt.Format(time.RFC3339Nano),
		"endsAt":       "0001-01-01T00:00:00Z",
		"generatorURL": "http://prometheus.test/graph",
		"fingerprint":  fingerprint,
	}
	if endsAt != nil {
		alert["endsAt"] = endsAt.Format(time.RFC3339Nano)
	}
	return alert
}

func (s *BackendSuite) TestAlertmanagerDeliveries() {
	j := s.newAlertmanagerJourney()

	// Installing ignores client config; each installation has its own reference.
	first := j.install()
	second := j.install()
	s.NotEqual(first.Id, second.Id)
	s.NotEqual(first.Attributes.ProviderInstallationRef, second.Attributes.ProviderInstallationRef)
	s.NotEqual("client-supplied", first.Attributes.ProviderInstallationRef)

	// Issuing again replaces the earlier token; the installation says whether one was issued.
	get := j.API(j.owner).Operation[oapiv1.GetIntegrationInstallationRequest, oapiv1.GetIntegrationInstallationResponse](oapiv1.GetIntegrationInstallation)
	webhookUrlIssued := func() bool {
		return get.Call(s.T().Context(), oapiv1.GetIntegrationInstallationRequest{Id: first.Id}).Body.Data.Attributes.WebhookUrlIssued
	}
	s.False(first.Attributes.WebhookUrlIssued)
	s.False(webhookUrlIssued())
	replacedPath := j.issueWebhookPath(first.Id)
	s.True(webhookUrlIssued())
	firstPath := j.issueWebhookPath(first.Id)
	secondPath := j.issueWebhookPath(second.Id)

	startsAt := j.clock.Now().Add(-5 * time.Minute).UTC().Truncate(time.Second)
	checkoutLabels := map[string]string{"alertname": "HighErrorRate", "service": "Checkout_API", "severity": "critical"}
	checkout := alertmanagerAlert("checkout-1", "firing", checkoutLabels, startsAt, nil)
	aliasLabels := map[string]string{"alertname": "HighErrorRate", "service_name": "checkout-api"}
	aliased := alertmanagerAlert("checkout-2", "firing", aliasLabels, startsAt, nil)

	j.deliver(replacedPath, http.StatusNotFound, checkout)

	// Identical firing deliveries within a minute are one observation of each alert.
	j.deliver(firstPath, http.StatusNoContent, checkout, aliased)
	j.deliver(firstPath, http.StatusNoContent, checkout, aliased)
	s.Equal(2, j.countEvents(first.Id))

	// The next minute's delivery is another observation of the same windows.
	j.clock.Advance(time.Minute)
	j.deliver(firstPath, http.StatusNoContent, checkout, aliased)
	s.Equal(4, j.countEvents(first.Id))
	s.Len(j.windows("checkout-1"), 1)

	// Both service label spellings name one service and, for one alert name, one definition.
	s.Equal(1, j.countAliases(first.Id, "service:"))
	s.Equal(1, j.countAliases(first.Id, "alert:"))

	// A resolution redelivered across minutes is one normalized event.
	endsAt := startsAt.Add(4 * time.Minute)
	resolved := alertmanagerAlert("checkout-1", "resolved", checkoutLabels, startsAt, &endsAt)
	j.deliver(firstPath, http.StatusNoContent, resolved)
	j.clock.Advance(time.Minute)
	j.deliver(firstPath, http.StatusNoContent, resolved)
	s.Equal(5, j.countEvents(first.Id))
	resolvedWindows := j.windows("checkout-1")
	s.Require().Len(resolvedWindows, 1)
	s.Require().NotNil(resolvedWindows[0].ResolvedAt)
	s.True(endsAt.Equal(*resolvedWindows[0].ResolvedAt))

	// A resolution delivered before its firing leaves the window resolved at the source end time.
	earlyLabels := map[string]string{"alertname": "DiskFull", "service": "checkout-api"}
	j.deliver(firstPath, http.StatusNoContent, alertmanagerAlert("early-1", "resolved", earlyLabels, startsAt, &endsAt))
	j.deliver(firstPath, http.StatusNoContent, alertmanagerAlert("early-1", "firing", earlyLabels, startsAt, nil))
	earlyWindows := j.windows("early-1")
	s.Require().Len(earlyWindows, 1)
	s.Require().NotNil(earlyWindows[0].ResolvedAt)
	s.True(endsAt.Equal(*earlyWindows[0].ResolvedAt))

	// Another installation's identical alert has its own service, definition and window.
	j.deliver(secondPath, http.StatusNoContent, checkout)
	s.Equal(1, j.countAliases(second.Id, "service:"))
	s.Equal(1, j.countAliases(second.Id, "alert:"))
	s.Len(j.windows("checkout-1"), 2)
	definitionCount, definitionsErr := j.Client(j.ctx).AlertDefinition.Query().Where(ald.Title("HighErrorRate")).Count(j.ctx)
	s.Require().NoError(definitionsErr)
	s.Equal(2, definitionCount)

	// Editing service labels keeps the installation's identity, token and entities, and applies to new
	// observations.
	api := j.API(j.owner)
	update := api.Operation[oapiv1.UpdateIntegrationInstallationRequest, oapiv1.UpdateIntegrationInstallationResponse](oapiv1.UpdateIntegrationInstallation)
	invalidRequest := oapiv1.UpdateIntegrationInstallationRequest{Id: first.Id}
	invalidRequest.Body.Attributes.UserSettings = map[string]any{"service_labels": []any{"job", "job"}}
	update.ExpectStatus(s.T().Context(), invalidRequest, http.StatusBadRequest)

	updateRequest := oapiv1.UpdateIntegrationInstallationRequest{Id: first.Id}
	updateRequest.Body.Attributes.UserSettings = map[string]any{"service_labels": []any{"job"}}
	updated := update.Call(s.T().Context(), updateRequest).Body.Data
	s.Equal(first.Id, updated.Id)
	s.Equal(first.Attributes.ProviderInstallationRef, updated.Attributes.ProviderInstallationRef)

	paymentsLabels := map[string]string{"alertname": "HighErrorRate", "job": "payments", "service": "checkout"}
	j.deliver(firstPath, http.StatusNoContent, alertmanagerAlert("payments-1", "firing", paymentsLabels, startsAt, nil))
	s.Equal(2, j.countAliases(first.Id, "service:"), "the checkout service remains and payments is added")
	s.Equal(3, j.countAliases(first.Id, "alert:"), "HighErrorRate and DiskFull on checkout, HighErrorRate on payments")
}

func (s *BackendSuite) TestAlertmanagerAlertsForDifferentServicesStayApart() {
	j := s.newAlertmanagerJourney()
	installation := j.install()
	path := j.issueWebhookPath(installation.Id)

	startsAt := j.clock.Now().UTC().Truncate(time.Second)
	checkout := alertmanagerAlert("checkout-errors", "firing", map[string]string{"alertname": "HighErrorRate", "service": "checkout"}, startsAt, nil)
	search := alertmanagerAlert("search-errors", "firing", map[string]string{"alertname": "HighErrorRate", "service": "search"}, startsAt, nil)
	j.deliver(path, http.StatusNoContent, checkout, search)

	queryEpisodes := j.Client(j.ctx).AlertEpisode.Query().
		Where(ale.HasAlertDefinitionWith(ald.Title("HighErrorRate")))
	episodes, episodesErr := queryEpisodes.All(j.ctx)
	s.Require().NoError(episodesErr)
	s.Require().Len(episodes, 2)
	s.NotEqual(episodes[0].AlertDefinitionID, episodes[1].AlertDefinitionID)

	j.RunDueJobs(j.ctx, jobs.ProcessSituationSignal{}.Kind(), j.clock.Now())
	j.RunDueJobs(j.ctx, jobs.EvaluateSituation{}.Kind(), j.clock.Now())

	listSituations := j.API(j.owner).Operation[oapiv1.ListSituationsRequest, oapiv1.ListSituationsResponse](oapiv1.ListSituations)
	situations := listSituations.Call(s.T().Context(), oapiv1.ListSituationsRequest{}).Body.Data
	s.Require().Len(situations, 2, "matching alert names alone do not join situations")
	for _, situation := range situations {
		s.Equal(1, situation.Attributes.SignalCount)
	}
}
