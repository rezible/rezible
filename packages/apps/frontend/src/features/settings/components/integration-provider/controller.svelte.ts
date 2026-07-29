import { Context, watch, type Getter } from "runed";
import type { Component } from "svelte";

import {
	type ErrorModel,
	type InstallIntegrationRequestAttributes,
	type IntegrationInstallation,
	type IntegrationOAuthInstallResult,
	type UpdateIntegrationInstallationRequestAttributes,
} from "$lib/api";

import { IntegrationOAuthController } from "$features/settings/lib/integrationsOAuthController.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

import SlackProvider from "./slack/SlackProvider.svelte";
import GoogleProvider from "./google/GoogleProvider.svelte";
import GithubProvider from "./github/GithubProvider.svelte";
import DemoProvider from "./demo/DemoProvider.svelte";

type ConfigMap = Record<string, unknown>;

const providerComponents: Record<string, Component> = {
	slack: SlackProvider,
	google: GoogleProvider,
	github: GithubProvider,
	demo: DemoProvider,
};

export class IntegrationProviderConfigController {
	integrations = useIntegrationsController();
	oauth = new IntegrationOAuthController((res) => {
		this.onOAuthResult(res);
	});

	private name = $state.raw<string>();
	nameValid = $derived(!!this.name);

	installations = $derived(
		!!this.name ? this.integrations.installationsByProvider.get(this.name) || [] : []
	);
	availableIntegrations = $derived(
		!!this.name ? this.integrations.availableByProvider.get(this.name) || [] : []
	);

	ProviderComponent = $derived(
		!!this.name && this.name in providerComponents ? providerComponents[this.name] : undefined
	);

	loading = $derived(this.oauth.inFlow || this.integrations.loading);

	constructor(nameFn: Getter<string>) {
		watch(nameFn, (name) => {
			this.name = name;
		});
	}

	async startOAuthFlow(integrationName: string) {
		if (this.loading) return;
		try {
			await this.oauth.startFlowFor(integrationName);
		} catch (e) {
			this.setConfigError(e);
		}
	}

	private onOAuthResult(res: IntegrationOAuthInstallResult) {
		this.integrations.refetchInstalled();
		this.integrations.refetchInstallTargets();
	}

	oauthError = $derived(this.oauth.inFlow ? this.oauth.error : undefined);

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

	editingIntegrationName = $state<string>();
	editingInstallation = $state.raw<IntegrationInstallation>();

	userSettings = $state.raw<ConfigMap>({});
	userSettingsValid = $state(false);

	installConfig = $state.raw<ConfigMap>({});
	installConfigValid = $state(false);

	setEditing(name: string, installation?: IntegrationInstallation) {
		this.editingIntegrationName = name;
		this.editingInstallation = installation;
	}

	clearEditing() {
		this.editingIntegrationName = undefined;
		this.editingInstallation = undefined;

		this.userSettings = {};
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
		this.installConfig = config;
		this.installConfigValid = valid;
	}

	async saveInstall() {
		if (!this.editingIntegrationName) return;
		try {
			if (!this.editingInstallation) {
				const attributes: InstallIntegrationRequestAttributes = {
					config: this.installConfig,
				};
				await this.integrations.installNew(this.editingIntegrationName, attributes);
			} else {
				const attributes: UpdateIntegrationInstallationRequestAttributes = {
					userSettings: this.userSettings,
				};
				await this.integrations.updateInstallation(this.editingInstallation.id, attributes);
			}
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}

	availableIntegration(name: string) {
		return this.integrations.availableByName.get(name);
	}

	installationsFor(name: string) {
		return this.integrations.installationsByName.get(name) ?? [];
	}

	installTargetOptionsFor(name: string) {
		return this.integrations.installationTargetsByName.get(name) ?? [];
	}

	isInstalled(name: string) {
		return this.installationsFor(name).length > 0;
	}

	async disconnect(id: string) {
		try {
			await this.integrations.deleteInstallation(id);
			this.setConfigError();
		} catch (e) {
			this.setConfigError(e);
		}
	}
}

const ctx = new Context<IntegrationProviderConfigController>("IntegrationProviderConfigController");
export const initIntegrationProviderConfigController = (nameFn: Getter<string>) =>
	ctx.set(new IntegrationProviderConfigController(nameFn));
export const useIntegrationProviderConfigController = () => ctx.get();
