package river

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
)

func (s *JobService) RegisterWorkers(
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
	debriefs rez.DebriefService,
) error {
	generateDebriefResponse := river.WorkFunc(func(ctx context.Context, j *river.Job[generateIncidentDebriefResponseJobArgs]) error {
		return debriefs.GenerateResponse(ctx, j.Args.debriefId)
	})
	sendDebriefRequests := river.WorkFunc(func(ctx context.Context, j *river.Job[sendIncidentDebriefRequestsJobArgs]) error {
		return debriefs.SendUserDebriefRequests(ctx, j.Args.incidentId)
	})
	ensureShiftHandovers := river.WorkFunc(func(ctx context.Context, j *river.Job[ensureShiftHandoverJobArgs]) error {
		return oncall.EnsureShiftHandover(ctx, j.Args.shiftId)
	})
	return errors.Join(
		river.AddWorkerSafely(s.clientCfg.Workers, sendDebriefRequests),
		river.AddWorkerSafely(s.clientCfg.Workers, generateDebriefResponse),
		river.AddWorkerSafely(s.clientCfg.Workers, ensureShiftHandovers),
		s.registerOncallHandoverScanPeriodicJob(time.Hour, oncall),
		s.registerProviderDataSyncPeriodicJob(time.Hour, users, incidents, oncall, alerts),
	)
}

// Send requests for users to complete debriefs
type sendIncidentDebriefRequestsJobArgs struct {
	incidentId uuid.UUID
}

func (sendIncidentDebriefRequestsJobArgs) Kind() string {
	return "send-incident-debrief-requests"
}

// Generate response to user debrief messages
type generateIncidentDebriefResponseJobArgs struct {
	debriefId uuid.UUID
}

func (generateIncidentDebriefResponseJobArgs) Kind() string {
	return "generate-incident-debrief-response"
}

// Generate Debrief Suggestions
type generateIncidentDebriefSuggestionsJobArgs struct {
	debriefId uuid.UUID
}

func (generateIncidentDebriefSuggestionsJobArgs) Kind() string {
	return "generate-incident-debrief-suggestions"
}

// Ensure Shift Handover (send user reminder, or auto-send fallback if unsent)
type ensureShiftHandoverJobArgs struct {
	shiftId uuid.UUID
}

func (ensureShiftHandoverJobArgs) Kind() string {
	return "send-shift-handover-reminder"
}
