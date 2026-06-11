import { Context, watch } from "runed";
import type { Component } from "svelte";
import { SvelteSet } from "svelte/reactivity";

import {
	type ErrorModel,
	type AvailableIntegration,
	type InstalledIntegration,
	type CreateInstalledIntegrationRequestBody,
} from "$lib/api";

import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

import SlackAgent from "./config-components/SlackAgent.svelte";
import PlaceholderConfig from "./config-components/PlaceholderConfig.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import FakeConfig from "./config-components/FakeConfig.svelte";
import SlackIncidents from "./config-components/SlackIncidents.svelte";

const configs: Record<string, Component> = {
	"fake": FakeConfig,
	"slack_agent": SlackAgent,
	"slack_incidents": SlackIncidents,
	"google": GoogleConfig,
	"github": GithubConfig,
};

export type ConfigRequestAttributes = CreateInstalledIntegrationRequestBody["attributes"];

export type ConfigureIntegrationDialogParams = {
    integration?: AvailableIntegration;
    installation?: InstalledIntegration;
}

export class ConfigureIntegrationDialogController {
	integrations = useIntegrationsController();
	oauth = useIntegrationOAuthController();

	private oauthTarget = $derived.by(() => {
		if (!this.oauth.inFlowForName) return; 
		return this.integrations.availableByName.get(this.oauth.inFlowForName);
	});
	constructor() {
		watch(() => this.oauthTarget, oauthIntegration => {
			if (!oauthIntegration) return;
			console.log("opening dialog for oauth target", oauthIntegration);
			this.integrations.openConfigureDialog(oauthIntegration);
		});
	}

	private params = $derived(this.integrations.configureDialogParams);
    integration = $derived(this.params?.integration);
    installation = $derived(this.params?.installation);

    name = $derived(this.integration?.name);
    ConfigComponent = $derived((!!this.name && this.name in configs) ? (configs[this.name] ?? PlaceholderConfig) : undefined);

	async startOAuthFlow() {
		if (this.loading || !this.name) return;
		try {
			await this.oauth.startFlowFor(this.name);
		} catch (e) {
			this.setConfigError(e);
		}
	}

	oauthPending = $derived(this.oauth.inFlowForName === this.name);
    oauthError = $derived(this.oauthPending ? this.oauth.error : undefined);

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

	installPending = $derived(!!this.name && this.integrations.installingName === this.name);

	configAttrs = $state<ConfigRequestAttributes>();

    isOpen = $derived(!!this.integration);

	setOpen(open: boolean) {
		if (!open && !this.oauthPending) this.close();
	}

	open(params: ConfigureIntegrationDialogParams) {
		if (this.loading || !this.name || !this.integration) return;
		this.integrations.configureDialogParams = params;
	}

	close() {
		this.integrations.configureDialogParams = undefined;
	}

	setConfig(attrs: ConfigRequestAttributes) {
		this.configAttrs = attrs;
	}

	configValid = $derived(!!this.configAttrs);

	async saveConfig() {
		if (!this.name || !this.configValid || !this.configAttrs) return;
		try {
			await this.integrations.installNew(this.name, this.configAttrs);
			// this.clearConfig();
			this.setConfigError();
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

	loading = $derived(this.oauthPending || this.installPending);
}

const ctx = new Context<ConfigureIntegrationDialogController>("ConfigureIntegrationDialogController");
export const initConfigureIntegrationDialogController = () => ctx.set(new ConfigureIntegrationDialogController());
export const useConfigureIntegrationDialogController = () => ctx.get();
