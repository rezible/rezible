package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	at "github.com/rezible/rezible/ent/agentturn"
	ale "github.com/rezible/rezible/ent/alertepisode"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

const (
	investigationReportText  = "Checkout errors need investigation."
	investigationQuestion    = "What should we check next?"
	investigationAnswerTitle = "Next check"
	investigationAnswerBody  = "Check the checkout error rate."
)

func (s *BackendSuite) TestInvestigation() {
	t := s.T()
	model := &investigationModel{}
	t.Cleanup(func() { model.AssertComplete(t) })
	// The application reads the global tracer provider as it is built.
	spans := tracetest.NewSpanRecorder()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	t.Cleanup(func() { otel.SetTracerProvider(previousProvider) })
	ah := s.newAppHarness(appTestOptions{ModelAction: model.generate})
	ctx, alice := ah.NewIdentity("Alice")
	_, bob := ah.NewIdentity("Bob")
	aliceAPI := ah.API(alice)
	requestCtx := s.T().Context()

	client := ah.Client(ctx)
	situations, situationsErr := ah.app.invoke[rez.SituationService]()
	s.Require().NoError(situationsErr, "resolve situation service")

	// Record an alert and create a situation from its episode without starting an investigation.
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
	episodeEntityID := s.recordAlert(ah, ctx, event)

	observationGroup := rez.SituationObservationGroupParams{
		Title:           "Checkout errors",
		SignalEntityIDs: []uuid.UUID{episodeEntityID},
	}
	params := rez.CreateSituationParams{
		Title:             "Checkout errors",
		SeedEntityID:      episodeEntityID,
		ObservationGroups: []rez.SituationObservationGroupParams{observationGroup},
	}
	situation, situationErr := situations.CreateSituation(ctx, params)
	s.Require().NoError(situationErr, "create checkout situation")

	// Start through HTTP and capture the investigation's linked identities.
	startInvestigation := aliceAPI.Operation[
		oapiv1.RaiseSituationRequest,
		oapiv1.RaiseSituationResponse,
	](oapiv1.RaiseSituation)
	startRequest := oapiv1.RaiseSituationRequest{Id: situation.ID}
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
	s.requireTurnLinkedToRequest(spans, oapiv1.SubmitInvestigationUserInput.OperationID, secondTurn.ID, situation.ID)

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
}

// requireTurnLinkedToRequest requires the turn to run in its own trace: its agent.turn span names the
// situation, Genkit's spans nest beneath it, and its job's work span links back, through the job that
// inserted it, to the API request that started the chain.
func (s *BackendSuite) requireTurnLinkedToRequest(spans *tracetest.SpanRecorder, operationID string, turnID, situationID uuid.UUID) {
	s.T().Helper()
	var request, turn sdktrace.ReadOnlySpan
	byID := make(map[trace.SpanID]sdktrace.ReadOnlySpan)
	for _, span := range spans.Ended() {
		byID[span.SpanContext().SpanID()] = span
		if span.Name() == operationID {
			request = span
		}
		if span.Name() == "agent.turn" && slices.Contains(span.Attributes(), attribute.String("agent_turn_id", turnID.String())) {
			turn = span
		}
	}
	s.Require().NotNil(request, "request span")
	s.Require().NotNil(turn, "agent.turn span")
	s.NotEqual(request.SpanContext().TraceID(), turn.SpanContext().TraceID(), "the turn runs in its job's trace")
	s.Subset(turn.Attributes(), []attribute.KeyValue{
		attribute.String("situation_id", situationID.String()),
		attribute.String("outcome", "completed"),
	})

	// ancestors returns span's ancestors within its trace, nearest first.
	ancestors := func(span sdktrace.ReadOnlySpan) []sdktrace.ReadOnlySpan {
		var found []sdktrace.ReadOnlySpan
		for parent, known := byID[span.Parent().SpanID()]; known && span.Parent().IsValid(); parent, known = byID[span.Parent().SpanID()] {
			found = append(found, parent)
			span = parent
		}
		return found
	}

	hasGenkitChild := false
	for _, span := range spans.Ended() {
		if span.InstrumentationScope().Name == "genkit-tracer" && slices.Contains(ancestors(span), turn) {
			s.Equal(turn.SpanContext().TraceID(), span.SpanContext().TraceID())
			hasGenkitChild = true
		}
	}
	s.True(hasGenkitChild, "Genkit's spans nest beneath the turn")

	// Follow each job's link to the span that inserted it until the request's trace is reached.
	work := turn
	for hop := 0; work.SpanContext().TraceID() != request.SpanContext().TraceID(); hop++ {
		s.Require().Less(hop, 5, "the turn's trace links back to the request")
		if chain := ancestors(work); len(chain) > 0 {
			work = chain[len(chain)-1]
		}
		s.Require().True(strings.HasPrefix(work.Name(), "river.work/"), "a trace starts at a job's work span, not %s", work.Name())
		s.Require().Len(work.Links(), 1, "%s links to the span that inserted its job", work.Name())
		insert, known := byID[work.Links()[0].SpanContext.SpanID()]
		s.Require().True(known, "%s links to a recorded span", work.Name())
		s.Equal("river.insert_many", insert.Name())
		work = insert
	}
	s.Contains(ancestors(work), request, "the request inserted the first job")
}

// recordAlert records the event as a firing notification of a checkout alert definition and returns its
// episode's knowledge entity.
func (s *BackendSuite) recordAlert(ah *appHarness, ctx context.Context, event *ent.NormalizedEvent) uuid.UUID {
	knowledge, knowledgeErr := ah.app.invoke[rez.KnowledgeGraphIngestionService]()
	s.Require().NoError(knowledgeErr, "resolve knowledge ingestion")
	alerts, alertsErr := ah.app.invoke[rez.AlertService]()
	s.Require().NoError(alertsErr, "resolve alert service")

	definitionRef := rez.ProviderResourceRef{Provider: event.Provider, ProviderNamespace: event.ProviderNamespace, ResourceRef: event.ProviderResourceRef}
	definitionEntityRef := rez.KnowledgeEntityRef{Category: kne.CategorySignal, Kind: "alert", ProviderResourceRef: definitionRef}
	definitionEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "alert_definition_observed",
		EffectiveAt: event.OccurredAt,
		Subject:     rez.KnowledgeSubjectRef{Entity: &definitionEntityRef},
	}
	ingestErr := knowledge.IngestEvidence(ctx, event, definitionEvidence)
	s.Require().NoError(ingestErr, "ingest alert definition evidence")

	queryAlias := ah.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderResourceRef(definitionRef.ResourceRef))
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr, "read alert definition alias")

	recordParams := rez.RecordAlertInstanceParams{
		Event: event,
		Definition: rez.AlertDefinitionValues{
			KnowledgeEntityID: *alias.EntityID,
			Title:             "Checkout errors",
		},
		Instance: rez.AlertInstanceValues{
			InstanceID: "checkout",
			Severity:   schematypes.SignalSeverityCritical,
			Firing:     true,
			StartedAt:  event.OccurredAt,
		},
	}
	definition, recordErr := alerts.RecordAlertInstance(ctx, recordParams)
	s.Require().NoError(recordErr, "record checkout alert")

	queryEpisode := ah.Client(ctx).AlertEpisode.Query().
		Where(ale.AlertDefinitionID(definition.ID))
	episode, episodeErr := queryEpisode.Only(ctx)
	s.Require().NoError(episodeErr, "read checkout alert episode")
	return *episode.KnowledgeEntityID
}

// The script exercises real tools across two turns; it never writes domain data.
type investigationModel struct {
	mu   sync.Mutex
	step int
}

func (m *investigationModel) AssertComplete(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	assert.Equal(t, 4, m.step, "investigation model must consume all four calls")
}

func (m *investigationModel) generate(
	_ context.Context,
	request *ai.ModelRequest,
	_ any,
	_ ai.ModelStreamCallback,
) (*ai.ModelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.step++

	switch m.step {
	case 1:
		input := rezai.PublishInvestigationReportToolInput{
			Text:         investigationReportText,
			EvidenceRefs: []string{},
		}
		call := &ai.ToolRequest{
			Name:  rezai.PublishInvestigationReportTool.Name(),
			Ref:   "report-1",
			Input: input,
		}
		return m.requestTool(request, call)

	case 2:
		var result rezai.InvestigationReportToolResult
		responseErr := m.decodeToolResult(request, rezai.PublishInvestigationReportTool.Name(), "report-1", &result)
		if responseErr != nil {
			return nil, responseErr
		}
		if result.Text != investigationReportText {
			return nil, fmt.Errorf("step 2: report text=%q, want %q", result.Text, investigationReportText)
		}
		if len(result.EvidenceRefs) != 0 {
			return nil, fmt.Errorf("step 2: report evidence_refs=%v, want empty", result.EvidenceRefs)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage("Recorded."),
			FinishReason: ai.FinishReasonStop,
		}, nil

	case 3:
		expectedQuestion := "User question:\n" + investigationQuestion
		questionFound := false
		for _, message := range request.Messages {
			if message.Role == ai.RoleUser && message.Text() == expectedQuestion {
				questionFound = true
				break
			}
		}
		if !questionFound {
			return nil, fmt.Errorf("step 3: follow-up user question %q absent from model request", investigationQuestion)
		}
		input := rezai.PublishInvestigationAnswerToolInput{
			Title:        investigationAnswerTitle,
			Body:         investigationAnswerBody,
			EvidenceRefs: []string{},
		}
		call := &ai.ToolRequest{
			Name:  rezai.PublishInvestigationAnswerTool.Name(),
			Ref:   "answer-1",
			Input: input,
		}
		return m.requestTool(request, call)

	case 4:
		var result rezai.InvestigationFindingVersionToolResult
		responseErr := m.decodeToolResult(request, rezai.PublishInvestigationAnswerTool.Name(), "answer-1", &result)
		if responseErr != nil {
			return nil, responseErr
		}
		if result.Title != investigationAnswerTitle {
			return nil, fmt.Errorf("step 4: answer title=%q, want %q", result.Title, investigationAnswerTitle)
		}
		if result.Body != investigationAnswerBody {
			return nil, fmt.Errorf("step 4: answer body=%q, want %q", result.Body, investigationAnswerBody)
		}
		if !result.IsAnswer {
			return nil, fmt.Errorf("step 4: published finding is not an answer")
		}
		if len(result.EvidenceRefs) != 0 {
			return nil, fmt.Errorf("step 4: answer evidence_refs=%v, want empty", result.EvidenceRefs)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage("Recorded."),
			FinishReason: ai.FinishReasonStop,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected model call %d", m.step)
	}
}

func (m *investigationModel) requestTool(request *ai.ModelRequest, call *ai.ToolRequest) (*ai.ModelResponse, error) {
	for _, tool := range request.Tools {
		if tool.Name == call.Name {
			part := ai.NewToolRequestPart(call)
			return &ai.ModelResponse{
				Message: ai.NewModelMessage(part),
			}, nil
		}
	}
	return nil, fmt.Errorf("step %d: required tool %s missing", m.step, call.Name)
}

func (m *investigationModel) decodeToolResult(request *ai.ModelRequest, toolName, callID string, output any) error {
	for _, message := range request.Messages {
		for _, part := range message.Content {
			response := part.ToolResponse
			if response == nil || response.Name != toolName || response.Ref != callID {
				continue
			}
			encoded, encodeErr := json.Marshal(response.Output)
			if encodeErr != nil {
				return fmt.Errorf("step %d: encode %s result: %w", m.step, toolName, encodeErr)
			}
			if decodeErr := json.Unmarshal(encoded, output); decodeErr != nil {
				return fmt.Errorf("step %d: decode %s result: %w", m.step, toolName, decodeErr)
			}
			return nil
		}
	}
	return fmt.Errorf("step %d: response for tool %s call %s absent", m.step, toolName, callID)
}
