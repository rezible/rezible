import { Context, watch, type Getter } from "runed";
import {
	type ErrorModel,
	type AvailableIntegration,
	type ConfiguredIntegration,
	type IntegrationProviderDataSyncStatus,
	getIntegrationDataSyncStatusOptions,
	requestIntegrationDataSyncMutation,
} from "$lib/api";

import SlackConfig from "./config-components/SlackConfig.svelte";
import PlaceholderConfig from "./config-components/PlaceholderConfig.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";
import type { Component } from "svelte";
import { resolve } from "$app/paths";
import { createMutation, createQuery } from "@tanstack/svelte-query";

const configs: Record<string, Component> = {
	slack: SlackConfig,
	google: GoogleConfig,
	github: GithubConfig,
};

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

export class IntegrationCardController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	constructor(integrationFn: Getter<AvailableIntegration>) {
		watch(integrationFn, (intg) => {
			this.integration = intg;
		});
	}

	private integration = $state<AvailableIntegration>();
	name = $derived(this.integration?.name);
	Component = $derived(this.name && this.name in configs ? configs[this.name] : PlaceholderConfig);

	configured = $derived<ConfiguredIntegration[]>(
		!!this.name ? (this.integrations.configuredByProvider.get(this.name) ?? []) : []
	);
	hasConfigured = $derived(this.configured.length > 0);
	primaryConfigured = $derived(this.configured[0]);

	dataKinds = $derived.by<Record<string, boolean>>(() => {
		const result: Record<string, boolean> = {};
		for (const intg of this.configured) {
			for (const [kind, enabled] of Object.entries(intg.attributes.dataKinds)) {
				result[kind] = result[kind] || enabled;
			}
		}
		return result;
	});
	enabledDataKinds = $derived(
		Object.entries(this.dataKinds)
			.filter(([_, enabled]) => !!enabled)
			.map(([name, _]) => name) ?? []
	);

	hasChanges = $state(false);
	private configDraft = $state<Record<string, unknown>>({});
	setConfig(key: string, val: unknown) {
		this.configDraft[key] = val;
		this.hasChanges = true;
	}

	clearConfig(key?: string) {
		if (key) {
			delete this.configDraft[key];
		} else {
			this.configDraft = {};
		}
		this.hasChanges = Object.keys(this.configDraft).length > 0;
	}

	configError = $state<ErrorModel>();
	private setConfigError(err?: unknown) {
		if (!err) {
			this.configError = undefined;
			return;
		}
		this.configError = {
			title: "Integration Setup Failed",
			detail: err instanceof Error ? err.message : "An unknown issue occurred",
		};
	}

	private isConfiguring = $derived(!!this.name && this.integrations.configuringProviderName === this.name);
	private isOAuthPending = $derived(this.oauth.pending && this.oauth.name === this.name);

	loading = $derived(this.isOAuthPending || this.isConfiguring);

	async startOAuthFlow() {
		if (this.loading || !this.name) return;
		const name = this.name;
		try {
			const callbackPath = resolve(`/settings/integration-callback/${name}`);
			const flowUrl = await this.oauth.getStartFlowUrl(this.name, callbackPath);
			window.location.assign(flowUrl);
		} catch (e) {
			this.setConfigError(e);
		}
	}

	oauthSelectionRequired = $derived(this.oauth.selectionRequired && this.oauth.name === this.name);

	async save() {
		if (!this.name || !this.hasChanges) return;
		const currConfig = this.primaryConfigured?.attributes.config ?? {};
		const externalRef = this.primaryConfigured?.attributes.externalRef ?? "default";
		const displayName = this.primaryConfigured?.attributes.displayName ?? this.name;
		try {
			await this.integrations.configure(this.name, {
				displayName,
				externalRef,
				config: { ...currConfig, ...$state.snapshot(this.configDraft) },
			});
			this.clearConfig();
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}

	private syncRequestPolling = $state(false);
	private syncPollTimeout: ReturnType<typeof setTimeout> | undefined;
	private syncRequestError = $state<ErrorModel>();

	private syncStatusQueryOptions = $derived(
		getIntegrationDataSyncStatusOptions({
			path: { name: this.name ?? "" },
		})
	);
	private syncStatusQuery = createQuery(() => ({
		...this.syncStatusQueryOptions,
		enabled: !!this.name && this.hasConfigured,
		refetchInterval: this.syncRequestPolling ? pollIntervalMs : false,
	}));
	private requestDataSyncMutation = createMutation(() => ({
		...requestIntegrationDataSyncMutation(),
		onSuccess: async () => {
			this.syncRequestError = undefined;
			this.startSyncStatusPolling();
			await this.syncStatusQuery.refetch();
		},
		onError: (err) => {
			this.syncRequestError = err;
		},
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

	async requestSync() {
		if (!this.name || !this.hasConfigured || this.requestDataSyncMutation.isPending) return;
		await this.requestDataSyncMutation.mutateAsync({
			path: { name: this.name },
			body: { attributes: {} },
		});
	}

	async refetchSyncStatus() {
		this.syncRequestError = undefined;
		await this.syncStatusQuery.refetch();
	}
}

const ctx = new Context<IntegrationCardController>("IntegrationCardController");
export const initIntegrationCardController = (fn: Getter<AvailableIntegration>) =>
	ctx.set(new IntegrationCardController(fn));
export const useIntegrationCardController = () => ctx.get();
