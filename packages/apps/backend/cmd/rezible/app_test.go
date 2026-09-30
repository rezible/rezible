package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	at "github.com/rezible/rezible/ent/agentturn"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	usr "github.com/rezible/rezible/ent/user"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

type BackendSuite struct {
	test.Suite
}

func TestBackendSuite(t *testing.T) {
	suite.Run(t, &BackendSuite{Suite: test.NewSuite()})
}

func (s *BackendSuite) TestIngestAndQuery() {
	ah := s.newAppHarness(nil)
	ctx, owner := ah.NewIdentity("Provider owner")
	api := ah.API(owner)
	requestCtx := s.T().Context()
	client := ah.Client(ctx)
	pipeline, pipelineErr := ah.app.invoke[rez.ProviderEventPipelineService]()
	s.Require().NoError(pipelineErr, "resolve provider pipeline")

	event := rez.ProviderEvent{
		Provider:            "demo",
		ProviderNamespace:   "app-test",
		ProviderEventSource: "users",
		ProviderEventRef:    "alice-delivery",
		ReceivedAt:          time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
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
	getUser := api.Operation[
		oapiv1.GetUserRequest,
		oapiv1.GetUserResponse,
	](oapiv1.GetUser)
	userRequest := oapiv1.GetUserRequest{Id: projectedUser.ID}
	userResponse := getUser.Call(requestCtx, userRequest)
	s.Equal(projectedUser.ID, userResponse.Body.Data.Id)
	s.Equal("Alice", userResponse.Body.Data.Attributes.Name)
	s.Equal("alice@example.com", userResponse.Body.Data.Attributes.Email)

	// The same endpoint requires a session when called through the full server.
	anonymousGetUser := ah.api.Operation[
		oapiv1.GetUserRequest,
		oapiv1.GetUserResponse,
	](oapiv1.GetUser)
	anonymousGetUser.ExpectStatus(requestCtx, userRequest, http.StatusUnauthorized)

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

func (s *BackendSuite) TestInvestigation() {
	model := &investigationModel{}
	ah := s.newAppHarness(model.generate)
	ctx, alice := ah.NewIdentity("Alice")
	_, bob := ah.NewIdentity("Bob")
	aliceAPI := ah.API(alice)
	requestCtx := s.T().Context()

	client := ah.Client(ctx)
	situations, situationsErr := ah.app.invoke[rez.SituationService]()
	s.Require().NoError(situationsErr, "resolve situation service")

	// Seed evidence and create a situation without starting an investigation.
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := client.NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("investigations").
		SetProviderResourceRef("checkout-alert").
		SetKind("alert").
		SetProviderEventSource("app-test").
		SetProviderEventRef("checkout-alert-1").
		SetAttributes([]byte(`{"message":"checkout errors increased"}`)).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt)
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr, "seed checkout alert evidence")

	observationGroup := rez.SituationObservationGroupParams{
		Title:              "Source evidence",
		NormalizedEventIDs: []uuid.UUID{event.ID},
	}
	params := rez.CreateSituationParams{
		Title:              "Checkout errors",
		StartInvestigation: false,
		ObservationGroups:  []rez.SituationObservationGroupParams{observationGroup},
	}
	situation, situationErr := situations.CreateSituation(ctx, params)
	s.Require().NoError(situationErr, "create checkout situation")

	// Start through HTTP and capture the investigation's linked identities.
	startInvestigation := aliceAPI.Operation[
		oapiv1.RequestSituationInvestigationRequest,
		oapiv1.RequestSituationInvestigationResponse,
	](oapiv1.RequestSituationInvestigation)
	startRequest := oapiv1.RequestSituationInvestigationRequest{Id: situation.ID}
	startResponse := startInvestigation.Call(requestCtx, startRequest)
	s.Require().NotNil(startResponse.Body.Data.Attributes.Investigation)
	investigationID := startResponse.Body.Data.Attributes.Investigation.Investigation.Id
	s.Require().NotEqual(uuid.Nil, investigationID)

	getInvestigation := aliceAPI.Operation[
		oapiv1.GetInvestigationRequest,
		oapiv1.GetInvestigationResponse,
	](oapiv1.GetInvestigation)
	detailRequest := oapiv1.GetInvestigationRequest{Id: investigationID}
	detailResponse := getInvestigation.Call(requestCtx, detailRequest)
	s.Equal(investigationID, detailResponse.Body.Data.Id)
	sessionID := detailResponse.Body.Data.Attributes.SessionId
	analysisID := detailResponse.Body.Data.Attributes.AnalysisId
	s.Require().NotEqual(uuid.Nil, sessionID)
	s.Require().NotEqual(uuid.Nil, analysisID)

	// The initial turn publishes the exact completed report through the real tool.
	firstTurn := ah.AwaitInvestigationTurn(ctx, sessionID, 1)
	getReport := aliceAPI.Operation[
		oapiv1.GetInvestigationReportRequest,
		oapiv1.GetInvestigationReportResponse,
	](oapiv1.GetInvestigationReport)
	reportRequest := oapiv1.GetInvestigationReportRequest{
		Id:        investigationID,
		Selection: "completed",
	}
	reportResponse := getReport.Call(requestCtx, reportRequest)
	report := reportResponse.Body.Data.Attributes
	s.Equal(investigationReportText, report.Text)
	s.Equal(firstTurn.ID, report.AgentTurnId)
	s.Equal(string(at.StatusCompleted), report.TurnStatus)
	s.False(report.Provisional)

	// Repeating start preserves the investigation, analysis, session, and first turn.
	repeatedStartResponse := startInvestigation.Call(requestCtx, startRequest)
	s.Require().NotNil(repeatedStartResponse.Body.Data.Attributes.Investigation)
	s.Equal(investigationID, repeatedStartResponse.Body.Data.Attributes.Investigation.Investigation.Id)
	repeatedDetailResponse := getInvestigation.Call(requestCtx, detailRequest)
	s.Equal(sessionID, repeatedDetailResponse.Body.Data.Attributes.SessionId)
	s.Equal(analysisID, repeatedDetailResponse.Body.Data.Attributes.AnalysisId)
	queryTurns := client.AgentTurn.Query().
		Where(at.AgentSessionID(sessionID))
	initialTurnCount, initialCountErr := queryTurns.Count(ctx)
	s.Require().NoError(initialCountErr, "count initial investigation turns")
	s.Equal(1, initialTurnCount)

	// Submit a follow-up; the second turn publishes its answer through the real tool.
	submitInput := aliceAPI.Operation[
		oapiv1.SubmitInvestigationUserInputRequest,
		oapiv1.SubmitInvestigationUserInputResponse,
	](oapiv1.SubmitInvestigationUserInput)
	questionRequest := oapiv1.SubmitInvestigationUserInputRequest{Id: investigationID}
	questionRequest.Body.Text = investigationQuestion
	questionRequest.Body.SubmissionKey = "follow-up-1"
	submittedResponse := submitInput.Call(requestCtx, questionRequest)
	inputID := submittedResponse.Body.Data.Id
	s.Require().NotEqual(uuid.Nil, inputID)
	secondTurn := ah.AwaitInvestigationTurn(ctx, sessionID, 2)

	listFindings := aliceAPI.Operation[
		oapiv1.ListInvestigationFindingsRequest,
		oapiv1.ListInvestigationFindingsResponse,
	](oapiv1.ListInvestigationFindings)
	findingsRequest := oapiv1.ListInvestigationFindingsRequest{Id: investigationID}
	findingsResponse := listFindings.Call(requestCtx, findingsRequest)
	s.Require().Len(findingsResponse.Body.Data, 1)
	answer := findingsResponse.Body.Data[0]
	s.Equal(investigationAnswerTitle, answer.Attributes.Title)
	s.Equal(investigationAnswerBody, answer.Attributes.Body)
	s.Require().NotNil(answer.Attributes.UserInputId)
	s.Equal(inputID, *answer.Attributes.UserInputId)
	s.Equal(secondTurn.ID, answer.Attributes.AgentTurnId)
	s.Equal(string(at.StatusCompleted), answer.Attributes.TurnStatus)

	// The input points back to the exact answer version and producing turn.
	listInputs := aliceAPI.Operation[
		oapiv1.ListInvestigationUserInputsRequest,
		oapiv1.ListInvestigationUserInputsResponse,
	](oapiv1.ListInvestigationUserInputs)
	inputsRequest := oapiv1.ListInvestigationUserInputsRequest{Id: investigationID}
	inputsResponse := listInputs.Call(requestCtx, inputsRequest)
	s.Require().Len(inputsResponse.Body.Data, 1)
	input := inputsResponse.Body.Data[0]
	s.Equal(inputID, input.Id)
	s.Equal(investigationQuestion, input.Attributes.Text)
	s.Require().NotNil(input.Attributes.AnswerVersionId)
	s.Equal(answer.Id, *input.Attributes.AnswerVersionId)
	s.Require().NotNil(input.Attributes.AgentTurn)
	s.Equal(secondTurn.ID, input.Attributes.AgentTurn.Id)
	s.Equal(at.StatusCompleted, input.Attributes.AgentTurn.Status)

	// Bob's tenant cannot read or add input to Alice's investigation.
	bobAPI := ah.API(bob)
	bobGetInvestigation := bobAPI.Operation[
		oapiv1.GetInvestigationRequest,
		oapiv1.GetInvestigationResponse,
	](oapiv1.GetInvestigation)
	bobGetInvestigation.ExpectStatus(requestCtx, detailRequest, http.StatusNotFound)
	rejectedRequest := oapiv1.SubmitInvestigationUserInputRequest{Id: investigationID}
	rejectedRequest.Body.Text = "Unauthorized question"
	rejectedRequest.Body.SubmissionKey = "bob-1"
	bobSubmitInput := bobAPI.Operation[
		oapiv1.SubmitInvestigationUserInputRequest,
		oapiv1.SubmitInvestigationUserInputResponse,
	](oapiv1.SubmitInvestigationUserInput)
	bobSubmitInput.ExpectStatus(requestCtx, rejectedRequest, http.StatusNotFound)

	retainedInputsResponse := listInputs.Call(requestCtx, inputsRequest)
	s.Require().Len(retainedInputsResponse.Body.Data, 1)
	s.Equal(inputID, retainedInputsResponse.Body.Data[0].Id)
	finalTurnCount, finalCountErr := queryTurns.Count(ctx)
	s.Require().NoError(finalCountErr, "count turns after rejected input")
	s.Equal(2, finalTurnCount)
	model.AssertComplete(s.T())
}
