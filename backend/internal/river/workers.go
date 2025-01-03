package river

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
)

func (s *JobService) RegisterWorkers(
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
	debriefs rez.DebriefService,
) error {
	generateDebriefResponse := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.GenerateIncidentDebriefResponse]) error {
		return debriefs.GenerateResponse(ctx, j.Args.DebriefId)
	})
	sendDebriefRequests := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.SendIncidentDebriefRequests]) error {
		return debriefs.SendUserDebriefRequests(ctx, j.Args.IncidentId)
	})
	ensureShiftHandovers := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.EnsureShiftHandover]) error {
		return oncall.EnsureShiftHandover(ctx, j.Args.ShiftId)
	})
	return errors.Join(
		river.AddWorkerSafely(s.clientCfg.Workers, sendDebriefRequests),
		river.AddWorkerSafely(s.clientCfg.Workers, generateDebriefResponse),
		river.AddWorkerSafely(s.clientCfg.Workers, ensureShiftHandovers),
		s.registerOncallHandoverScanPeriodicJob(time.Hour, oncall),
		s.registerProviderDataSyncPeriodicJob(time.Hour, users, incidents, oncall, alerts),
	)
}
