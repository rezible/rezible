package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	usr "github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/ent/userauthsession"
	internalhttp "github.com/rezible/rezible/internal/http"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

func (s *BackendSuite) TestIngestAndQuery() {
	ah := s.newAppHarness(appTestOptions{})

	baseCtx := s.T().Context()
	ctx, owner := ah.NewIdentity("Provider owner")
	api := ah.API(owner)

	pipeline, pipelineErr := ah.app.invoke[rez.ProviderEventPipelineService]()
	s.Require().NoError(pipelineErr, "resolve provider pipeline")

	event := rez.ProviderEvent{
		Provider:            "demo",
		ProviderNamespace:   "app-test",
		ProviderEventSource: "users",
		ProviderEventRef:    "alice-delivery",
		ReceivedAt:          time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		// TODO: use an attributes struct and encode it
		Attributes: []byte(`{
			"external_id": "alice",
			"name": "Alice",
			"email": "alice@example.com",
			"chat_id": "UALICE",
			"timezone": "Australia/Perth",
			"updated_at": "2026-06-01T10:00:00Z"
		}`),
	}

	// Queue the delivery and wait for the real process and projection workers.
	s.Require().NoError(pipeline.Ingest(ctx, event), "ingest demo user delivery")
	ah.AwaitProviderEvents(ctx)

	queryUser := ah.Client(ctx).User.Query().
		Where(usr.Email("alice@example.com"))
	projectedUser, userErr := queryUser.Only(ctx)
	s.Require().NoError(userErr, "read projected Alice user")

	// Query the projected user through the registered, authenticated v1 operation.
	getUser := api.Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	userRequest := oapiv1.GetUserRequest{Id: projectedUser.ID}
	userResponse := getUser.Call(baseCtx, userRequest)
	s.Equal("Alice", userResponse.Body.Data.Attributes.Name)
	s.Equal("alice@example.com", userResponse.Body.Data.Attributes.Email)

	// The same endpoint requires a session when called through the full server.
	anonymousGetUser := ah.api.Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	anonymousGetUser.ExpectStatus(baseCtx, userRequest, http.StatusUnauthorized)
}

func (s *BackendSuite) TestRegeneratedWebhookUrlReplacesTheEarlierOne() {
	ah := s.newAppHarness(appTestOptions{})
	baseCtx := s.T().Context()

	ctx, admin := ah.NewIdentity("Webhook admin")
	setAdmin := ah.Client(ctx).UserAuthSession.UpdateOne(admin.Session).SetRole(userauthsession.RoleAdmin)
	s.Require().NoError(setAdmin.Exec(ctx), "make the session an admin's")
	api := ah.API(admin)

	// A member of the same organization.
	memberUser, memberUserErr := ah.Client(ctx).User.Create().
		SetName("Webhook member").
		SetEmail(uuid.NewString() + "@example.com").
		Save(ctx)
	s.Require().NoError(memberUserErr, "create member user")
	memberSession, memberSessionErr := ah.Client(ctx).UserAuthSession.Create().
		SetOrganizationID(admin.Session.OrganizationID).
		SetUserID(memberUser.ID).
		SetRole(userauthsession.RoleMember).
		SetExpiresAt(time.Now().UTC().Add(time.Hour)).
		Save(ctx)
	s.Require().NoError(memberSessionErr, "create member session")
	memberAPI := ah.API(test.Identity{Session: memberSession})

	server, serverErr := ah.app.invoke[*internalhttp.Server]()
	s.Require().NoError(serverErr, "resolve application HTTP server")
	deliver := func(webhookUrl string, body string) int {
		parsed, parseErr := url.Parse(webhookUrl)
		s.Require().NoError(parseErr)
		request := httptest.NewRequest(http.MethodPost, parsed.Path, bytes.NewReader([]byte(body)))
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, request)
		return recorder.Code
	}

	install := api.Operation[oapiv1.InstallIntegrationRequest, oapiv1.InstallIntegrationResponse](oapiv1.InstallIntegration)
	memberIssue := memberAPI.Operation[oapiv1.IssueIntegrationWebhookTokenRequest, oapiv1.IssueIntegrationWebhookTokenResponse](oapiv1.IssueIntegrationWebhookToken)
	issue := api.Operation[oapiv1.IssueIntegrationWebhookTokenRequest, oapiv1.IssueIntegrationWebhookTokenResponse](oapiv1.IssueIntegrationWebhookToken)

	cases := []struct {
		integration string
		config      map[string]any
		delivery    string
	}{
		{
			integration: "alertmanager",
			config:      map[string]any{},
			delivery:    `{"version":"4","status":"firing","truncatedAlerts":0,"alerts":[]}`,
		},
		{
			integration: "webhook",
			config:      map[string]any{"preset": "deployment"},
			delivery: `{"id":"run-1","service":"checkout-api","environment":"production","status":"succeeded",` +
				`"started_at":"2026-10-07T14:01:10Z","finished_at":"2026-10-07T14:03:52Z"}`,
		},
	}
	for _, tc := range cases {
		var installRequest oapiv1.InstallIntegrationRequest
		installRequest.Name = tc.integration
		installRequest.Body.Attributes.Config = tc.config
		installation := install.Call(baseCtx, installRequest).Body.Data

		issueRequest := oapiv1.IssueIntegrationWebhookTokenRequest{Id: installation.Id}
		firstUrl := issue.Call(baseCtx, issueRequest).Body.Data.Url
		s.Equal(http.StatusNoContent, deliver(firstUrl, tc.delivery), "%s: the first URL works", tc.integration)

		regeneratedUrl := issue.Call(baseCtx, issueRequest).Body.Data.Url
		s.NotEqual(firstUrl, regeneratedUrl, tc.integration)
		s.Equal(http.StatusNotFound, deliver(firstUrl, tc.delivery), "%s: the earlier URL stops working", tc.integration)
		s.Equal(http.StatusNoContent, deliver(regeneratedUrl, tc.delivery), "%s: the regenerated URL works", tc.integration)

		memberIssue.ExpectStatus(baseCtx, issueRequest, http.StatusForbidden)
		s.Equal(http.StatusNoContent, deliver(regeneratedUrl, tc.delivery), "%s: a member cannot replace the URL", tc.integration)
	}
	ah.AwaitProviderEvents(ctx)
}
