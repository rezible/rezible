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
	type IntegrationInstallTarget,
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

const configs: Record<string, Component> = {
	slack: SlackConfig,
	google: GoogleConfig,
	github: GithubConfig,
	fake: FakeConfig,
};

export type ConfigRequestAttributes = CreateInstalledIntegrationRequestBody["attributes"];

export class AvailableIntegrationCardController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	private integration = $state<AvailableIntegration>();

	constructor(integrationFn: Getter<AvailableIntegration>) {
		watch(integrationFn, (intg) => {
			this.integration = intg;
		});
	}

	name = $derived(this.integration?.name);
	ConfigComponent = $derived(this.name && this.name in configs ? configs[this.name] : PlaceholderConfig);

	installations = $derived<InstalledIntegration[]>(
		this.integrations.installationsByName.get(this.name || "") ?? []);
	hasInstalled = $derived(this.installations.length > 0);

	capabilities = $derived.by<Record<string, boolean>>(() => {
		const result: Record<string, boolean> = {};
		for (const installation of this.installations) {
			for (const [cap, enabled] of Object.entries(installation.attributes.capabilities)) {
				result[cap] = result[cap] || enabled;
			}
		}
		return result;
	});
	enabledCapabilities = $derived(
		Object.entries(this.capabilities)
			.filter(([_, enabled]) => !!enabled)
			.map(([name, _]) => name) ?? []
	);

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

	isInstallPending = $derived(!!this.name && this.integrations.installingName === this.name);
	isOAuthPending = $derived(this.oauth.inFlow && this.oauth.installingName === this.name);

	loading = $derived(this.isOAuthPending || this.isInstallPending);

	editingInstallation = $state<InstalledIntegration>();
	configAttrs = $state<ConfigRequestAttributes>();
	showConfig = $state(false);

	startConfig(ii?: InstalledIntegration) {
		if (this.loading || !this.name || !this.integration) return;
		this.editingInstallation = ii;
		this.showConfig = true;
	}

	setConfig(attrs: ConfigRequestAttributes) {
		this.configAttrs = attrs;
	}
	hasChanges = $derived(!!this.configAttrs);

	updateConfig(updateFn: (attrs?: ConfigRequestAttributes) => ConfigRequestAttributes) {
		this.configAttrs = updateFn(this.configAttrs);
	}

	clearConfig() {
		this.showConfig = false;
		this.configAttrs = undefined;
		this.editingInstallation = undefined;
	}

	async saveConfig() {
		if (!this.name || !this.hasChanges || !this.configAttrs) return;
		try {
			await this.integrations.installNew(this.name, this.configAttrs);
			this.clearConfig();
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}

	async startOAuthFlow() {
		if (this.loading || !this.name) return;
		try {
			const flowUrl = await this.oauth.getStartFlowUrl(this.name);
			window.location.assign(flowUrl);
		} catch (e) {
			this.setConfigError(e);
		}
	}

	installTargetOptions = $derived(this.integrations.installationTargetsByName.get(this.name ?? ""));
	installationTargetSelectionRequired = $derived(!!this.installTargetOptions && this.installTargetOptions.length > 0);
	selectedInstallTargetExternalRefs = new SvelteSet<string>();

	toggleInstallationTargetSelection(ref: string, selected: boolean) {
		if (selected) {
			this.selectedInstallTargetExternalRefs.add(ref);
		} else {
			this.selectedInstallTargetExternalRefs.delete(ref);
		}
	}

	async confirmSelectedInstallationTargets() {
		if (!this.name || this.integrations.installingName || this.selectedInstallTargetExternalRefs.size === 0) return;
		const refs = [...this.selectedInstallTargetExternalRefs];
		try {
			await this.integrations.installFromTargets(this.name, refs);
			this.selectedInstallTargetExternalRefs.clear();
		} catch (e) {
			
		}
	}
}

const ctx = new Context<AvailableIntegrationCardController>("AvailableIntegrationCardController");
export const initAvailableIntegrationCardController = (fn: Getter<AvailableIntegration>) =>
	ctx.set(new AvailableIntegrationCardController(fn));
export const useAvailableIntegrationCardController = () => ctx.get();
