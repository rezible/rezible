import { Context } from "runed";

import {
	type ErrorModel,
	requestIntegrationEventSyncMutation,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { useIntegrationsController } from "../../lib/integrationsController.svelte";

type SyncStatusDisplay = {
	label: string;
	variant: "default" | "secondary" | "destructive" | "outline";
	class?: string;
};

const syncStatusDisplays: Record<string, SyncStatusDisplay> = {
	queued: { label: "Queued", variant: "outline" },
	started: { label: "Started", variant: "secondary", class: "text-blue-700" },
	complete: { label: "Complete", variant: "secondary", class: "text-green-700" },
	success: { label: "Success", variant: "secondary", class: "text-green-700" },
	error: { label: "Error", variant: "destructive" },
	failed: { label: "Failed", variant: "destructive" },
	skipped: { label: "Skipped", variant: "outline" },
};

const pollAfterRequestMs = 10_000;
const pollIntervalMs = 3_000;

export class IntegrationDataSyncController {
	private ctrl = useIntegrationsController();

	installation = $derived(this.ctrl.dataSyncDialogInstallation);
	isOpen = $derived(!!this.installation);

	id = $derived(this.installation?.id);

	setOpen(open: boolean) {
		if (!open) this.ctrl.dataSyncDialogInstallation = undefined;
	}

	private syncRequestPolling = $state(false);
	private syncPollTimeout: ReturnType<typeof setTimeout> | undefined;
	private syncRequestError = $state<ErrorModel>();

    syncStatusError = $derived(
		(this.syncRequestError/* ?? this.syncStatusQuery.error*/) as ErrorModel | undefined
	);
    latestSyncStatusDisplay = $derived<SyncStatusDisplay | undefined>(
		/*this.formatSyncStatus(this.latestSyncStatus)*/
        undefined
	);
	isSyncing = $derived(false);

	private requestDataSyncMutation = createMutation(() => ({
		...requestIntegrationEventSyncMutation(),
		onSuccess: async () => {
			this.syncRequestError = undefined;
		},
		onError: (err) => {
			this.syncRequestError = err;
		},
	}));

	async requestSync() {
		const id = this.id;
		if (!id || this.requestDataSyncMutation.isPending) return;
		await this.requestDataSyncMutation.mutateAsync({
			path: { id },
			body: { attributes: {} },
		});
	};

    disabled = $derived(this.requestDataSyncMutation.isPending);
    /*
	private syncStatusQueryOptions = $derived(
		listIntegrationEventSyncRunsOptions({
			path: { id: this.id ?? "" },
		})
	);
	private syncStatusQuery = createQuery(() => ({
		...this.syncStatusQueryOptions,
		enabled: !!this.ctrl.name && this.ctrl.hasInstalled,
		refetchInterval: this.syncRequestPolling ? pollIntervalMs : false,
	}));

	syncStatusRuns = $derived<IntegrationProviderDataSyncStatus[]>(this.syncStatusQuery.data?.data ?? []);
	syncStatusError = $derived(
		(this.syncRequestError ?? this.syncStatusQuery.error) as ErrorModel | undefined
	);
	latestSyncStatus = $derived<string | undefined>(this.syncStatusRuns[0]?.attributes.status);
	latestSyncStatusDisplay = $derived<SyncStatusDisplay | undefined>(
		this.formatSyncStatus(this.latestSyncStatus)
	);
	isSyncing = $derived(this.requestDataSyncMutation.isPending || this.syncRequestPolling);

	private formatSyncStatus(status?: string): SyncStatusDisplay | undefined {
		if (!status) return undefined;
		return syncStatusDisplays[status] ?? { label: status, variant: "outline" };
	}

	private startSyncStatusPolling() {
		this.syncRequestPolling = true;
		if (this.syncPollTimeout) {
			clearTimeout(this.syncPollTimeout);
		}
		this.syncPollTimeout = setTimeout(() => {
			this.syncRequestPolling = false;
			this.syncPollTimeout = undefined;
		}, pollAfterRequestMs);
	}

	async refetchSyncStatus() {
		this.syncRequestError = undefined;
		await this.syncStatusQuery.refetch();
	}
    */
}

const dataSyncCtx = new Context<IntegrationDataSyncController>("IntegrationDataSyncController");
export const initIntegrationDataSyncController = () => dataSyncCtx.set(new IntegrationDataSyncController());
export const useIntegrationDataSyncController = () => dataSyncCtx.get();