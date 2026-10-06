package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/firebase/genkit/go/ai"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/jobs"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/pkg/situations"
)

func (s *BackendSuite) TestSituationDetection() {
	ah := s.newAppHarness(appTestOptions{ModelAction: incidentalInvestigationModel})
	ctx, owner := ah.NewIdentity("Responder")
	requestCtx := s.T().Context()
	pipeline, pipelineErr := ah.app.invoke[rez.ProviderEventPipelineService]()
	s.Require().NoError(pipelineErr, "resolve provider pipeline")

	api := ah.API(owner)
	listSituations := api.Operation[oapiv1.ListSituationsRequest, oapiv1.ListSituationsResponse](oapiv1.ListSituations)
	situation := func() oapiv1.Situation {
		response := listSituations.Call(requestCtx, oapiv1.ListSituationsRequest{})
		s.Require().Len(response.Body.Data, 1)
		return response.Body.Data[0]
	}
	processKind := jobs.ProcessSituationSignal{}.Kind()
	evaluateKind := jobs.EvaluateSituation{}.Kind()
	firedAt := ah.clock.Now()
	// deliver ingests a notification of the definition's alert about the checkout service, then runs the
	// signal processing and evaluations it triggers.
	deliver := func(definition string, resolved bool) {
		event := s.demoAlertEvent(definition, firedAt, ah.clock.Now(), resolved)
		s.Require().NoError(pipeline.Ingest(ctx, event), "ingest %s alert", definition)
		ah.AwaitProjection(ctx, owner.Session.TenantID, event)
		ah.RunDueJobs(ctx, processKind, ah.clock.Now())
		ah.RunDueJobs(ctx, evaluateKind, ah.clock.Now())
	}

	// A firing alert about a service starts a candidate.
	deliver("checkout-errors", false)
	candidate := situation()
	s.Equal("candidate", candidate.Attributes.Stage)
	s.Equal(1, candidate.Attributes.SignalCount)

	// A firing alert from a second definition on the same service joins it.
	deliver("checkout-latency", false)
	joined := situation()
	s.Equal(candidate.Id, joined.Id)
	s.Equal("candidate", joined.Attributes.Stage)
	s.Equal(2, joined.Attributes.SignalCount)

	// Once the collection window has passed, its deadline evaluation raises it on breadth.
	ah.clock.Advance(situations.CollectionWindow)
	ah.RunDueJobs(ctx, evaluateKind, ah.clock.Now())
	raised := situation()
	s.Equal("raised", raised.Attributes.Stage)
	s.Require().NotNil(raised.Attributes.LatestJudgment)
	s.Equal([]string{"breadth"}, raised.Attributes.LatestJudgment.CitedReasons)

	// Raising starts real investigation work; allow it to succeed before advancing the evidence.
	s.Require().NotNil(raised.Attributes.Investigation)
	getInvestigation := api.Operation[
		oapiv1.GetInvestigationRequest,
		oapiv1.GetInvestigationResponse,
	](oapiv1.GetInvestigation)
	investigationRequest := oapiv1.GetInvestigationRequest{
		Id: raised.Attributes.Investigation.Investigation.Id,
	}
	investigationResponse := getInvestigation.Call(requestCtx, investigationRequest)
	ah.AwaitInvestigationTurn(ctx, investigationResponse.Body.Data.Attributes.SessionId, 1)

	// Both alerts resolve. Their episodes close after the flap grace, and the situation after the quiet period.
	deliver("checkout-errors", true)
	deliver("checkout-latency", true)
	ah.clock.Advance(s.Config().Alerts.FlapGrace)
	ah.RunDueJobs(ctx, jobs.SettleAlertEpisode{}.Kind(), ah.clock.Now())
	ah.RunDueJobs(ctx, processKind, ah.clock.Now())
	ah.RunDueJobs(ctx, evaluateKind, ah.clock.Now())
	s.Equal("raised", situation().Attributes.Stage, "the quiet period has not passed")

	ah.clock.Advance(situations.SituationQuietPeriod)
	ah.RunDueJobs(ctx, evaluateKind, ah.clock.Now())
	closed := situation()
	s.Equal("closed", closed.Attributes.Stage)
	s.Require().NotNil(closed.Attributes.CloseReason)
	s.Equal("stabilized", *closed.Attributes.CloseReason)
}

// demoAlertEvent is a demo provider notification of the definition's alert about the checkout service. The
// alert fired at firedAt; a resolved notification ends it when delivered.
func (s *BackendSuite) demoAlertEvent(definition string, firedAt, deliveredAt time.Time, resolved bool) rez.ProviderEvent {
	checkout := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "demo",
			ProviderNamespace: "app-test",
			ResourceRef:       "demo:component:checkout",
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: "Checkout",
	}
	state := projections.AlertStateFiring
	if resolved {
		state = projections.AlertStateResolved
	}
	ref := definition + "-" + state
	payload := map[string]any{
		"definition_ref":    definition,
		"title":             definition,
		"state":             state,
		"instance_id":       "checkout",
		"severity":          "critical",
		"started_at":        firedAt,
		"occurred_at":       deliveredAt,
		"notification_ref":  ref,
		"observed_entities": []projections.EntityObservation{checkout},
	}
	if resolved {
		payload["ended_at"] = deliveredAt
	}
	attributes, encodeErr := json.Marshal(payload)
	s.Require().NoError(encodeErr, "encode demo alert")
	return rez.ProviderEvent{
		Provider:            "demo",
		ProviderNamespace:   "app-test",
		ProviderEventSource: "alerts",
		ProviderEventRef:    ref,
		ReceivedAt:          deliveredAt,
		Attributes:          attributes,
	}
}

// Detection permits incidental investigation turns without scripting investigation outputs.
func incidentalInvestigationModel(context.Context, *ai.ModelRequest, any, ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	return &ai.ModelResponse{
		Message:      ai.NewModelTextMessage("Recorded."),
		FinishReason: ai.FinishReasonStop,
	}, nil
}
