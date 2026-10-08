package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/jobs"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

// installWebhook installs the webhook integration with a preset through the API.
func (j *alertmanagerJourney) installWebhook(config map[string]any, status int) uuid.UUID {
	install := j.API(j.owner).Operation[oapiv1.InstallIntegrationRequest, oapiv1.InstallIntegrationResponse](oapiv1.InstallIntegration)
	request := oapiv1.InstallIntegrationRequest{Name: "webhook"}
	request.Body.Attributes.Config = config
	if status != http.StatusOK {
		install.ExpectStatus(j.suite.T().Context(), request, status)
		return uuid.Nil
	}
	return install.Call(j.suite.T().Context(), request).Body.Data.Id
}

// reportDeployment posts a deployment report and waits for it to be processed and projected.
func (j *alertmanagerJourney) reportDeployment(path string, report map[string]any) {
	j.suite.T().Helper()
	body, encodeErr := json.Marshal(report)
	j.suite.Require().NoError(encodeErr)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	j.handler.ServeHTTP(recorder, request)
	j.suite.Require().Equal(http.StatusNoContent, recorder.Code, recorder.Body.String())
	j.awaitProviderEvents()
}

func (j *alertmanagerJourney) listSituations() []oapiv1.Situation {
	j.RunDueJobs(j.ctx, jobs.ProcessSituationSignal{}.Kind(), j.clock.Now())
	j.RunDueJobs(j.ctx, jobs.EvaluateSituation{}.Kind(), j.clock.Now())
	listSituations := j.API(j.owner).Operation[oapiv1.ListSituationsRequest, oapiv1.ListSituationsResponse](oapiv1.ListSituations)
	return listSituations.Call(j.suite.T().Context(), oapiv1.ListSituationsRequest{}).Body.Data
}

func (s *BackendSuite) TestDeploymentReportsLandOnTheAlertedServiceWithoutJoiningItsSituation() {
	j := s.newAlertmanagerJourney()
	alertmanager := j.install()
	s.Empty(alertmanager.Attributes.Metadata)
	alertmanagerPath := j.issueWebhookPath(alertmanager.Id)

	// Installing needs a known preset, which updating cannot change.
	j.installWebhook(map[string]any{}, http.StatusBadRequest)
	j.installWebhook(map[string]any{"preset": "incident"}, http.StatusBadRequest)
	webhookID := j.installWebhook(map[string]any{"preset": "deployment"}, http.StatusOK)
	update := j.API(j.owner).Operation[oapiv1.UpdateIntegrationInstallationRequest, oapiv1.UpdateIntegrationInstallationResponse](oapiv1.UpdateIntegrationInstallation)
	updateRequest := oapiv1.UpdateIntegrationInstallationRequest{Id: webhookID}
	updateRequest.Body.Attributes.UserSettings = map[string]any{"preset": "incident"}
	update.ExpectStatus(s.T().Context(), updateRequest, http.StatusBadRequest)
	installation, installationErr := j.Client(j.ctx).Integration.Get(j.ctx, webhookID)
	s.Require().NoError(installationErr)
	s.JSONEq(`{"preset":"deployment"}`, string(installation.InstallationConfig))
	get := j.API(j.owner).Operation[oapiv1.GetIntegrationInstallationRequest, oapiv1.GetIntegrationInstallationResponse](oapiv1.GetIntegrationInstallation)
	installed := get.Call(s.T().Context(), oapiv1.GetIntegrationInstallationRequest{Id: webhookID}).Body.Data
	s.Equal(map[string]string{"preset": "deployment"}, installed.Attributes.Metadata)
	webhookPath := j.issueWebhookPath(webhookID)

	startsAt := j.clock.Now().Add(-time.Minute).UTC().Truncate(time.Second)
	labels := map[string]string{"alertname": "HighErrorRate", "service": "Checkout_API"}
	j.deliver(alertmanagerPath, http.StatusNoContent, alertmanagerAlert("checkout-1", "firing", labels, startsAt, nil))
	watching := j.listSituations()
	s.Require().Len(watching, 1)
	s.Equal(1, watching[0].Attributes.SignalCount)

	report := map[string]any{
		"id":          "run-1",
		"service":     "checkout-api",
		"environment": "production",
		"status":      "succeeded",
		"repository":  "acme/checkout",
	}
	j.reportDeployment(webhookPath, report)

	// The deployment lands on the alerted service.
	queryObserves := j.Client(j.ctx).KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateObserves))
	observes, observesErr := queryObserves.Only(j.ctx)
	s.Require().NoError(observesErr)
	queryImpacts := j.Client(j.ctx).KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateImpacts))
	impacts, impactsErr := queryImpacts.Only(j.ctx)
	s.Require().NoError(impactsErr)
	s.Equal(observes.TargetEntityID, impacts.TargetEntityID)

	// It adds no signal to the watching situation and opens none.
	after := j.listSituations()
	s.Require().Len(after, 1)
	s.Equal(watching[0].Id, after[0].Id)
	s.Equal(1, after[0].Attributes.SignalCount)
}
