
import { Context, watch, type Getter } from "runed";
import type { Component } from "svelte";

import { createMutation, createQuery } from "@tanstack/svelte-query";
import {
	type ErrorModel,
	type AvailableIntegration,
	type InstalledIntegration,
	type IntegrationEventSyncRun,
	listIntegrationEventSyncRunsOptions,
	requestIntegrationEventSyncMutation,
	type CreateInstalledIntegrationRequestBody,
} from "$lib/api";

import { resolve } from "$app/paths";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

import SlackConfig from "./config-components/SlackConfig.svelte";
import PlaceholderConfig from "./config-components/PlaceholderConfig.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import FakeConfig from "./config-components/FakeConfig.svelte";
import { SvelteSet } from "svelte/reactivity";
import { useAvailableIntegrationCardController } from "./availableIntegrationController.svelte";

const configs: Record<string, Component> = {
	slack: SlackConfig,
	google: GoogleConfig,
	github: GithubConfig,
	fake: FakeConfig,
};

export type ConfigRequestAttributes = CreateInstalledIntegrationRequestBody["attributes"];

export class InstalledIntegrationCardController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	private installation = $state<InstalledIntegration>();

	constructor(fn: Getter<InstalledIntegration>) {
		watch(fn, (intg) => {this.installation = intg});
	}

	id = $derived(this.installation?.id);

}

const ctx = new Context<InstalledIntegrationCardController>("InstalledIntegrationCardController");
export const initInstalledIntegrationCardController = (fn: Getter<InstalledIntegration>) =>
	ctx.set(new InstalledIntegrationCardController(fn));
export const useInstalledIntegrationCardController = () => ctx.get();

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