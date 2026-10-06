package main

import (
	"net/http"
	"time"

	rez "github.com/rezible/rezible"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
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

	// Ingest the delivery and wait for its real process and projection workers.
	ingestErr := pipeline.Ingest(ctx, event)
	s.Require().NoError(ingestErr, "ingest demo user delivery")

	projection := ah.AwaitProjection(ctx, owner.Session.TenantID, event)
	s.Require().Len(projection.ProcessJobIDs, 1, "delivery should have one process job")

	client := ah.Client(ctx)

	// Independently read the projected user, its alias, and its event evidence.
	queryUser := client.User.Query().
		Where(usr.Email("alice@example.com"))
	projectedUser, userErr := queryUser.Only(ctx)
	s.Require().NoError(userErr, "read projected Alice user")

	queryAlias := client.KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider(event.Provider),
			ksa.ProviderNamespace(event.ProviderNamespace),
			ksa.ProviderResourceRef("demo:user:alice"),
		)
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr, "read demo user alias")
	s.Require().NotNil(alias.EntityID)
	s.Equal(projectedUser.KnowledgeEntityID, alias.EntityID)

	queryEvidence := client.KnowledgeEvidence.Query().
		Where(ke.EventID(projection.Event.ID))
	evidence, evidenceErr := queryEvidence.Only(ctx)
	s.Require().NoError(evidenceErr, "read delivery evidence")
	s.Equal(alias.ID, evidence.SubjectAliasID)
	s.Equal("UALICE", projectedUser.ChatID)

	// Query the projected user through the registered, authenticated v1 operation.
	getUser := api.Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	userRequest := oapiv1.GetUserRequest{Id: projectedUser.ID}
	userResponse := getUser.Call(baseCtx, userRequest)
	s.Equal(projectedUser.ID, userResponse.Body.Data.Id)
	s.Equal("Alice", userResponse.Body.Data.Attributes.Name)
	s.Equal("alice@example.com", userResponse.Body.Data.Attributes.Email)

	// The same endpoint requires a session when called through the full server.
	anonymousGetUser := ah.api.Operation[oapiv1.GetUserRequest, oapiv1.GetUserResponse](oapiv1.GetUser)
	anonymousGetUser.ExpectStatus(baseCtx, userRequest, http.StatusUnauthorized)

	// A repeated delivery reuses the completed process job and persisted records.
	repeatErr := pipeline.Ingest(ctx, event)
	s.Require().NoError(repeatErr, "repeat demo user delivery")

	repeatedJobIDs := ah.ProcessJobIDs(ctx, owner.Session.TenantID, event)
	s.Equal(projection.ProcessJobIDs, repeatedJobIDs)
	for _, jobID := range repeatedJobIDs {
		ah.AwaitJob(ctx, jobID)
	}

	repeatedUser, repeatedUserErr := queryUser.Only(ctx)
	s.Require().NoError(repeatedUserErr, "read user after repeated delivery")
	s.Equal(projectedUser.ID, repeatedUser.ID)

	repeatedEvidence, repeatedEvidenceErr := queryEvidence.Only(ctx)
	s.Require().NoError(repeatedEvidenceErr, "read evidence after repeated delivery")
	s.Equal(evidence.ID, repeatedEvidence.ID)

	queryEvent := client.NormalizedEvent.Query().
		Where(
			ne.Provider(event.Provider),
			ne.ProviderNamespace(event.ProviderNamespace),
			ne.ProviderEventSource(event.ProviderEventSource),
			ne.ProviderEventRef(event.ProviderEventRef),
		)
	repeatedEvent, repeatedEventErr := queryEvent.Only(ctx)
	s.Require().NoError(repeatedEventErr, "read normalized event after repeated delivery")
	s.Equal(projection.Event.ID, repeatedEvent.ID)

	queryReceipt := client.NormalizedEventProjection.Query().
		Where(nep.EventID(projection.Event.ID))
	repeatedReceipt, repeatedReceiptErr := queryReceipt.Only(ctx)
	s.Require().NoError(repeatedReceiptErr, "read receipt after repeated delivery")
	s.Equal(projection.Receipt.ID, repeatedReceipt.ID)
}
