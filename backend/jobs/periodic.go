package jobs

import "time"

var SyncAllTenantIntegrationsDataPeriodicJob = NewPeriodicJob(
	PeriodicInterval(time.Hour),
	func() (JobArgs, *InsertOpts) {
		return &SyncIntegrationsData{}, &InsertOpts{
			UniqueOpts: UniqueOpts{
				ByState: NonCompletedJobStates,
			},
		}
	},
	&PeriodicJobOpts{RunOnStart: true},
)

var ScanOncallShiftsPeriodicJob = NewPeriodicJob(
	PeriodicInterval(time.Hour),
	func() (JobArgs, *InsertOpts) {
		return &ScanOncallShifts{}, nil
	},
	&PeriodicJobOpts{RunOnStart: true},
)
