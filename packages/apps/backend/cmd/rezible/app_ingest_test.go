package main

import (
	"net/http"
	"time"

	rez "github.com/rezible/rezible"
	usr "github.com/rezible/rezible/ent/user"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
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
