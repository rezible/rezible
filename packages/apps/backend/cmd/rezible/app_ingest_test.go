package main

import (
	"net/http"

	rez "github.com/rezible/rezible"
	usr "github.com/rezible/rezible/ent/user"
	demoprovider "github.com/rezible/rezible/internal/integrations/demo"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *BackendSuite) TestIngestAndQuery() {
	ah := s.newAppHarness(appTestOptions{})
	ctx, owner := ah.NewIdentity("Provider owner")

	eventUser := projections.UserEventAttributes{
		Name:     "Alice",
		Email:    "alice@example.com",
		ChatId:   "UALICE",
		Timezone: "Australia/Perth",
	}
	event, eventErr := demoprovider.MakeUserProviderEvent(eventUser)
	s.Require().NoError(eventErr, "make demo user event")

	pipeline, pipelineErr := ah.app.invoke[rez.ProviderEventPipelineService]()
	s.Require().NoError(pipelineErr, "resolve provider pipeline")

	// Queue the delivery and wait for the real process and projection workers.
	s.Require().NoError(pipeline.Ingest(ctx, event), "ingest demo user delivery")
	ah.AwaitProviderEvents(ctx)

	queryUser := ah.Client(ctx).User.Query().Where(usr.Email(eventUser.Email))
	projectedUser, userErr := queryUser.Only(ctx)
	s.Require().NoError(userErr, "read projected event user")

	// Query the projected user through the registered, authenticated v1 operation.
	getUser := ah.API(owner).Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	userRequest := oapiv1.GetUserRequest{Id: projectedUser.ID}
	userResponse := getUser.Call(s.T().Context(), userRequest)
	s.Equal("Alice", userResponse.Body.Data.Attributes.Name)
	s.Equal("alice@example.com", userResponse.Body.Data.Attributes.Email)

	// The same endpoint requires a session when called through the full server.
	anonymousGetUser := ah.api.Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	anonymousGetUser.ExpectStatus(s.T().Context(), userRequest, http.StatusUnauthorized)
}
