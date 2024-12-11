package postgres

import (
	"errors"
)

func (s *AlertsService) RegisterJobs() error {
	//s.jobClient.AddPeriodicJob(jobs.PeriodicJob{
	//	Schedule: jobs.PeriodicInterval(time.Hour),
	//	Constructor: func() (jobs.JobArgs, *jobs.InsertOpts) {
	//		return &oncallDataSyncJobArgs{}, nil
	//	},
	//	Options: &jobs.PeriodicJobOpts{
	//		RunOnStart: true,
	//	},
	//})

	return errors.Join(
	//jobs.RegisterWorker(s.jobClient, &oncallDataSyncJobWorker{ds: s.dataSyncer}),
	//jobs.RegisterWorker(s.jobClient, &oncallHandoverPrepJobWorker{svc: s}),
	)
}
