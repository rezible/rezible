package river

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
)

func RegisterJobWorkers(
	syncer rez.IntegrationsDataSyncer,
	chat rez.ChatService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	debriefs rez.DebriefService,
) {
	RegisterPeriodicJob(jobs.SyncAllTenantIntegrationsDataPeriodicJob, syncer.SyncIntegrationsData)
	RegisterPeriodicJob(jobs.ScanOncallShiftsPeriodicJob, shifts.HandlePeriodicScanShifts)

	RegisterWorkerFunc(chat.HandleIncidentChatUpdate)

	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)
	RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)
}
