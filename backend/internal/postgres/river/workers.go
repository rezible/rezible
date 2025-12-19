package river

import (
	rez "github.com/rezible/rezible"
)

func RegisterJobWorkers(
	syncer rez.IntegrationsDataSyncService,
	chat rez.ChatService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	debriefs rez.DebriefService,
) {
	RegisterWorkerFunc(chat.ProcessEvent)
	RegisterWorkerFunc(chat.HandleIncidentChatUpdate)

	RegisterPeriodicJob(syncer.MakeSyncIntegrationsDataPeriodicJob(), syncer.SyncIntegrationsData)
	RegisterPeriodicJob(shifts.MakeScanShiftsPeriodicJob(), shifts.HandlePeriodicScanShifts)

	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)
	RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)
}
