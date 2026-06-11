import { Context, watch, type Getter } from "runed";

import {
	type ErrorModel,
	type AvailableIntegration,
	type InstalledIntegration,
	requestIntegrationEventSyncMutation,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";

import { getEnabledCapabilties, useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

export class AvailableIntegrationCardController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	private integration = $state<AvailableIntegration>();

	constructor(integrationFn: Getter<AvailableIntegration | undefined>) {
		watch(integrationFn, (intg) => {this.integration = intg});
		
	}

	name = $derived(this.integration?.name);
	displayName = $derived(this.integration?.displayName || "");
	description = $derived(this.integration?.description || "");
	provider = $derived(this.integration?.provider);
	supportedCapabilities = $derived(this.integration?.supportedCapabilities ?? []);

	installations = $derived<InstalledIntegration[]>(
		this.integrations.installationsByName.get(this.name || "") ?? []);
	hasInstalled = $derived(this.installations.length > 0);
	maxInstallsReached = $derived(
		typeof this.integration?.maxInstalls === "number" && this.installations.length >= this.integration.maxInstalls
	);
	canInstall = $derived(!this.maxInstallsReached);

	enabledCapabilities = $derived(getEnabledCapabilties(this.installations));

	openConfigDialog(installation?: InstalledIntegration) {
		this.integrations.configureDialogParams = {
			integration: $state.snapshot(this.integration),
			installation: $state.snapshot(installation),
		}
	}
}

const ctx = new Context<AvailableIntegrationCardController>("AvailableIntegrationCardController");
export const initAvailableIntegrationCardController = (fn: Getter<AvailableIntegration | undefined>) =>
	ctx.set(new AvailableIntegrationCardController(fn));
export const useAvailableIntegrationCardController = () => ctx.get();


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

export class InstalledIntegrationDataSyncController {
	private ctrl = useAvailableIntegrationCardController();

	private installation = $state<InstalledIntegration>();
	constructor(fn: Getter<InstalledIntegration>) {
		watch(fn, (intg) => {this.installation = intg});
	}
	id = $derived(this.installation?.id);

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
		if (!id || !this.ctrl.hasInstalled || this.requestDataSyncMutation.isPending) return;
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