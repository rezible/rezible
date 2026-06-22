import { Context, watch } from "runed";
import { SvelteSet } from "svelte/reactivity";

import {
	type AvailableIntegration,
	type CreateInstalledIntegrationRequestAttributes,
	type ErrorModel,
	type InstalledIntegration,
	type UpdateInstalledIntegrationRequestAttributes,
} from "$lib/api";

import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

import SlackAgent from "./config-components/SlackAgent.svelte";
import GoogleConfig from "./config-components/GoogleConfig.svelte";
import GithubConfig from "./config-components/GithubConfig.svelte";
import DemoConfig from "./config-components/DemoConfig.svelte";
import SlackIncidents from "./config-components/SlackIncidents.svelte";
import type { Component } from "svelte";

type ConfigMap = Record<string, unknown>;

const configs: Record<string, Component> = {
	demo: DemoConfig,
	slack_agent: SlackAgent,
	slack_incidents: SlackIncidents,
	google: GoogleConfig,
	github: GithubConfig,
};

export type ConfigureIntegrationDialogParams = {
	integration?: AvailableIntegration;
	installation?: InstalledIntegration;
};

export class ConfigureIntegrationDialogController {
	integrations = useIntegrationsController();

	private params = $derived(this.integrations.configureDialogParams);
	integration = $derived(this.params?.integration);
	installation = $derived(this.params?.installation);

	name = $derived(this.integration?.name);
	ConfigComponent = $derived(!!this.name && this.name in configs ? configs[this.name] : undefined);

	isEditMode = $derived(!!this.installation);
	isInstallMode = $derived(!this.isEditMode);

	userSettings = $state.raw<ConfigMap>({});
	userSettingsValid = $state(false);

	installConfig = $state.raw<ConfigMap>({});
	installConfigValid = $state(false);

	configValid = $derived(this.userSettingsValid && (this.isEditMode || this.installConfigValid));

	hasChanged = $derived(this.isInstallMode || (this.userSettingsValid));

	constructor() {
		this.resetDraft(this.params);
		watch(
			() => this.params,
			(params) => this.resetDraft(params)
		);
	}

	async startOAuthFlow() {
		if (this.loading || !this.name) return;
		try {
			await this.integrations.oauth.startFlowFor(this.name);
		} catch (e) {
			this.setConfigError(e);
		}
	}

	oauthPending = $derived(this.integrations.oauth.inFlowForName === this.name);
	oauthError = $derived(this.oauthPending ? this.integrations.oauth.error : undefined);

	configError = $state.raw<ErrorModel>();

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

	installPending = $derived(
		this.integrations.installationPending && this.integrations.installingName === this.name
	);

	private resetDraft(params?: ConfigureIntegrationDialogParams) {
		this.userSettings = params?.installation?.attributes.settings ?? {};
		this.userSettingsValid = false;

		this.installConfig = {};
		this.installConfigValid = false;
		
		this.setConfigError();
	}

	setUserSettings(settings: ConfigMap, valid = true) {
		this.userSettings = settings;
		this.userSettingsValid = valid;
	}

	setInstallConfig(config: ConfigMap, valid = true) {
		if (!this.isInstallMode) return;
		this.installConfig = config;
		this.installConfigValid = valid;
	}

	isOpen = $derived(!!this.integration);

	setOpen(open: boolean) {
		if (!open && !this.oauthPending) this.close();
	}

	open(params: ConfigureIntegrationDialogParams) {
		if (this.loading) return;
		this.integrations.configureDialogParams = params;
	}

	close() {
		this.integrations.configureDialogParams = undefined;
	}

	async saveConfig() {
		if (!this.name || !this.configValid) return;
		try {
			if (!!this.installation) {
				const attributes: UpdateInstalledIntegrationRequestAttributes = {
					userSettings: this.userSettings,
				};
				await this.integrations.updateInstallation(this.installation.id, attributes);
			} else {
				const attributes: CreateInstalledIntegrationRequestAttributes = {
					config: this.installConfig,
					userSettings: this.userSettings,
				};
				await this.integrations.installNew(this.name, attributes);
			}
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}

	installTargetOptions = $derived(this.integrations.installationTargetsByName.get(this.name ?? ""));
	installationTargetSelectionRequired = $derived(
		!!this.installTargetOptions && this.installTargetOptions.length > 0
	);
	selectedInstallTargetExternalRefs = new SvelteSet<string>();

	toggleInstallationTargetSelection(ref: string, selected: boolean) {
		if (selected) {
			this.selectedInstallTargetExternalRefs.add(ref);
		} else {
			this.selectedInstallTargetExternalRefs.delete(ref);
		}
	}

	async confirmSelectedInstallationTargets() {
		if (
			!this.name ||
			this.integrations.installingName ||
			this.selectedInstallTargetExternalRefs.size === 0
		)
			return;
		const refs = [...this.selectedInstallTargetExternalRefs];
		try {
			await this.integrations.installFromTargets(this.name, refs);
			this.selectedInstallTargetExternalRefs.clear();
		} catch {}
	}

	loading = $derived(this.oauthPending || this.installPending);
}

const ctx = new Context<ConfigureIntegrationDialogController>("ConfigureIntegrationDialogController");
export const initConfigureIntegrationDialogController = () =>
	ctx.set(new ConfigureIntegrationDialogController());
export const useConfigureIntegrationDialogController = () => ctx.get();
