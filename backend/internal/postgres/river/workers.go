package river

import (
	rez "github.com/rezible/rezible"
)

func RegisterJobWorkers(
	chat rez.ChatService,
	sync rez.ProviderSyncService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	debriefs rez.DebriefService,
) {
	RegisterWorkerFunc(chat.ProcessEvent)
	RegisterWorkerFunc(chat.HandleIncidentChatUpdate)

	RegisterPeriodicJob(sync.MakeSyncProviderDataPeriodicJob(), sync.SyncProviderData)

	RegisterPeriodicJob(shifts.MakeScanShiftsPeriodicJob(), shifts.HandlePeriodicScanShifts)
	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)
	RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)
}
