package river

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
)

func (s *JobService) RegisterWorkers(oncall rez.OncallService, debriefs rez.DebriefService) error {
	return errors.Join(
		s.registerSendDebriefRequests(debriefs),
		s.registerGenerateDebriefResponse(debriefs),
		s.registerEnsureShiftHandovers(oncall),
		s.registerOncallHandoverScanPeriodicJob(time.Hour, oncall),
	)
}

func (s *JobService) registerSendDebriefRequests(debriefs rez.DebriefService) error {
	workFn := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.SendIncidentDebriefRequests]) error {
		return debriefs.SendUserDebriefRequests(ctx, j.Args.IncidentId)
	})
	return river.AddWorkerSafely(s.clientCfg.Workers, workFn)
}

func (s *JobService) registerGenerateDebriefResponse(debriefs rez.DebriefService) error {
	workFn := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.GenerateIncidentDebriefResponse]) error {
		return debriefs.GenerateResponse(ctx, j.Args.DebriefId)
	})
	return river.AddWorkerSafely(s.clientCfg.Workers, workFn)
}

func (s *JobService) registerEnsureShiftHandovers(oncall rez.OncallService) error {
	workFn := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.EnsureShiftHandover]) error {
		return oncall.EnsureShiftHandover(ctx, j.Args.ShiftId)
	})
	return river.AddWorkerSafely(s.clientCfg.Workers, workFn)
}
