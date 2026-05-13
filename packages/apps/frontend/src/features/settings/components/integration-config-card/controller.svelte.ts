import { Context, watch, type Getter } from "runed";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import {
	startIntegrationOauthFlowMutation,
	completeIntegrationOauthFlowMutation,
	selectIntegrationOauthFlowMutation,
	type ErrorModel,
	listConfiguredIntegrationsOptions,
	type AvailableIntegration,
	type ExternalIntegrationOption,
	type ConfiguredIntegration,
} from "$lib/api";
import { page } from "$app/state";

import SlackConfig from "./config-components/SlackConfig.svelte";
import PlaceholderConfig from "./config-components/PlaceholderConfig.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import type { Component } from "svelte";
import { resolve } from "$app/paths";
import { clearQueryParams } from "$src/lib/utils";
import { SvelteSet } from "svelte/reactivity";

const configs: Record<string, Component> = {
	slack: SlackConfig,
	google: GoogleConfig,
	github: GithubConfig,
};

export class IntegrationConfigController {
	integrations = useIntegrationsController();
	private queryClient = useQueryClient();

	private inOAuthFlow = $state(false);

	constructor(integrationFn: Getter<AvailableIntegration>) {
		this.inOAuthFlow = false;
		watch(integrationFn, (intg) => {
			this.setIntegration(intg);
		});
	}

	private integration = $state<AvailableIntegration>();
	name = $derived(this.integration?.name);
	Component = $derived(this.name && this.name in configs ? configs[this.name] : PlaceholderConfig);

	private setIntegration(intg: AvailableIntegration) {
		this.integration = intg;
		if (intg.oauthRequired) this.checkOAuthCallback(intg.name);
	}

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
		this.inOAuthFlow = false;
		if (!err) {
			this.configError = undefined;
			return;
		}
		this.configError = {
			title: "Integration Setup Failed",
			detail: err instanceof Error ? err.message : "An unknown issue occurred",
		};
	}

	private startOAuthFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
	private completeOAuthFlowMut = createMutation(() => ({
		...completeIntegrationOauthFlowMutation({}),
		onSuccess: () => {
			void this.queryClient.invalidateQueries(listConfiguredIntegrationsOptions());
		},
	}));
	private selectOAuthFlowMut = createMutation(() => ({
		...selectIntegrationOauthFlowMutation({}),
		onSuccess: () => {
			void this.queryClient.invalidateQueries(listConfiguredIntegrationsOptions());
		},
	}));

	private oauthMutPending = $derived(this.startOAuthFlowMut.isPending ||
			this.completeOAuthFlowMut.isPending ||
			this.selectOAuthFlowMut.isPending);
	private isConfiguring = $derived(!!this.name && this.integrations.configuringProviderName === this.name);
	
	loading = $derived(this.inOAuthFlow || this.oauthMutPending || this.isConfiguring);

	selectionToken = $state<string>();
	selectionOptions = $state<ExternalIntegrationOption[]>([]);
	selectedExternalRefs = new SvelteSet<string>();
	selectionRequired = $derived(!!this.selectionToken && this.selectionOptions.length > 0);

	isSelected(ref: string) {
		return this.selectedExternalRefs.has(ref);
	}

	toggleSelection(ref: string, selected: boolean) {
		if (selected) {
			this.selectedExternalRefs.add(ref);
		} else {
			this.selectedExternalRefs.delete(ref);
		}
	}

	private setSelection(token?: string, options: ExternalIntegrationOption[] = []) {
		this.selectionToken = token;
		this.selectionOptions = options;
		this.selectedExternalRefs.clear();
		for (const option of options) {
			this.selectedExternalRefs.add(option.externalRef);
		}
	}

	async startOAuthFlow() {
		if (this.loading || !this.name) return;
		const name = this.name;
		try {
			const attributes = {
				callbackPath: resolve(`/settings/integration-callback/${name}`),
			}
			const resp = await this.startOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			this.inOAuthFlow = true;
			window.location.assign(new URL(resp.data.flow_url));
		} catch (e) {
			this.setConfigError(e);
		}
	}

	private async checkOAuthCallback(name: string) {
		if (this.completeOAuthFlowMut.isPending) return;

		const params = page.url.searchParams;
		const callbackName = params.get("name");
		if (!callbackName || callbackName !== this.name) return;

		const code = params.get("code");
		const state = params.get("state");

		if (!state || !code) return;

		try {
			const attributes = { state, code };
			const resp = await this.completeOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			if (resp.data.status === "selection_required") {
				if (!resp.data.selectionToken) {
					console.error("no selection token returned for oauth response");
				} else {
					this.setSelection(resp.data.selectionToken, resp.data.options);
				}
			} else {
				this.setSelection();
			}
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
		await clearQueryParams();
	}

	async selectOAuthOptions() {
		if (!this.name || !this.selectionToken || this.selectedExternalRefs.size === 0) return;
		try {
			const attributes = {
				selectionToken: this.selectionToken,
				externalRefs: [...this.selectedExternalRefs],
			}
			await this.selectOAuthFlowMut.mutateAsync({
				path: { name: this.name },
				body: { attributes },
			});
			this.setSelection();
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}

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
}

const ctx = new Context<IntegrationConfigController>("IntegrationConfigController");
export const initIntegrationConfigController = (fn: Getter<AvailableIntegration>) =>
	ctx.set(new IntegrationConfigController(fn));
export const useIntegrationConfigController = () => ctx.get();
