import { Context, watch, type Getter } from "runed";
import {
	type ErrorModel,
	type AvailableIntegration,
	type ExternalIntegrationOption,
	type ConfiguredIntegration,
	requestIntegrationDataSyncMutation,
} from "$lib/api";
import { page } from "$app/state";

import SlackConfig from "./config-components/SlackConfig.svelte";
import PlaceholderConfig from "./config-components/PlaceholderConfig.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";
import type { Component } from "svelte";
import { resolve } from "$app/paths";
import { clearQueryParams } from "$src/lib/utils";
import { SvelteSet } from "svelte/reactivity";
import { createMutation } from "@tanstack/svelte-query";

const configs: Record<string, Component> = {
	slack: SlackConfig,
	google: GoogleConfig,
	github: GithubConfig,
};

export class IntegrationCardController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	constructor(integrationFn: Getter<AvailableIntegration>) {
		watch(integrationFn, (intg) => {this.integration = intg});
	}

	private integration = $state<AvailableIntegration>();
	name = $derived(this.integration?.name);
	Component = $derived(this.name && this.name in configs ? configs[this.name] : PlaceholderConfig);

	configured = $derived<ConfiguredIntegration[]>(!!this.name ? (this.integrations.configuredByProvider.get(this.name) ?? []) : []);
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

	private requestDataSyncMutation = createMutation(() => requestIntegrationDataSyncMutation());
	async requestDataSync() {
		if (!this.name) return;
		this.requestDataSyncMutation.mutateAsync({ 
			path: { name: this.name },
			body: { attributes: {} }
		})
	}
}

const ctx = new Context<IntegrationCardController>("IntegrationCardController");
export const initIntegrationCardController = (fn: Getter<AvailableIntegration>) =>
	ctx.set(new IntegrationCardController(fn));
export const useIntegrationCardController = () => ctx.get();
