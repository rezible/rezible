import { goto } from "$app/navigation";
import { page } from "$app/state";
import { Context, watch, type Getter } from "runed";
import type { Component } from "svelte";
import { toast } from "svelte-sonner";

import type { ErrorModel, IntegrationInstallation, IntegrationOAuthInstallResult } from "$lib/api";

import { IntegrationOAuthController } from "$features/settings/lib/integrationsOAuthController.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

import SlackProvider from "./slack/SlackProvider.svelte";
import GoogleProvider from "./google/GoogleProvider.svelte";
import GithubProvider from "./github/GithubProvider.svelte";
import DemoProvider from "./demo/DemoProvider.svelte";
import AlertmanagerProvider from "./alertmanager/AlertmanagerProvider.svelte";
import GrafanaProvider from "./grafana/GrafanaProvider.svelte";

const providerComponents: Record<string, Component> = {
	slack: SlackProvider,
	google: GoogleProvider,
	github: GithubProvider,
	demo: DemoProvider,
	alertmanager: AlertmanagerProvider,
	grafana: GrafanaProvider,
};

export class IntegrationProviderController {
	integrations = useIntegrationsController();
	oauth = new IntegrationOAuthController((name, result) => {
		this.onOAuthComplete(name, result);
	});

	private providerName = $state("");

	provider = $derived(this.integrations.getProvider(this.providerName));
	ProviderComponent = $derived(providerComponents[this.providerName]);
	loading = $derived(this.integrations.loading);
	error = $derived(this.integrations.error);

	constructor(providerNameFn: Getter<string>) {
		watch(providerNameFn, (name) => {
			this.providerName = name;
		});
	}

	// The connection whose settings are open when a page lists several connections, kept in the URL.
	openConnectionId = $derived(page.url.searchParams.get("connection") ?? undefined);

	isOpen(id: string) {
		return this.openConnectionId === id;
	}

	toggleOpen = (id: string) => {
		if (this.isOpen(id)) {
			this.setOpenConnection(undefined);
		} else {
			this.setOpenConnection(id);
		}
	};

	private async setOpenConnection(id: string | undefined) {
		const url = new URL(page.url);
		if (id) {
			url.searchParams.set("connection", id);
		} else {
			url.searchParams.delete("connection");
		}
		await goto(url, { replaceState: true, noScroll: true, keepFocus: true });
	}

	private async onOAuthComplete(name: string, result: IntegrationOAuthInstallResult) {
		await this.integrations.refresh();

		const installed = result.installed ?? [];
		const firstInstalled = installed.at(0);
		if (firstInstalled) {
			toast.success(`Connected ${firstInstalled.attributes.displayName}.`);
		}
	}

	// Sign-in can return several targets, such as GitHub accounts, for the user to choose from.
	installTargetsPendingName = $state<string>();
	private installTargetsFailedName = $state<string>();
	private installTargetsError = $state.raw<ErrorModel>();

	installTargetsErrorFor(name: string) {
		if (this.installTargetsFailedName !== name) {
			return undefined;
		}
		return this.installTargetsError;
	}

	installFromTargets = async (name: string, resourceRefs: string[]) => {
		if (this.installTargetsPendingName || resourceRefs.length === 0) return;

		this.installTargetsPendingName = name;
		this.installTargetsFailedName = undefined;
		this.installTargetsError = undefined;
		try {
			const installed = await this.integrations.installFromTargets(name, resourceRefs);
			toast.success(
				installed.length === 1 ? "Connected." : `Connected ${installed.length} connections.`
			);
		} catch (e) {
			this.installTargetsFailedName = name;
			this.installTargetsError = e as ErrorModel;
		} finally {
			this.installTargetsPendingName = undefined;
		}
	};

	// Installation through a form, for integrations that do not use sign-in.
	installPendingName = $state<string>();
	private installFailedName = $state<string>();
	private installError = $state.raw<ErrorModel>();

	installErrorFor(name: string) {
		if (this.installFailedName !== name) {
			return undefined;
		}
		return this.installError;
	}

	// Installing again for an existing target, such as the same Google customer ID,
	// replaces its credentials and keeps its settings.
	async install(name: string, config: Record<string, unknown>, successMessage?: string) {
		if (this.installPendingName) return false;

		this.installPendingName = name;
		this.installFailedName = undefined;
		this.installError = undefined;
		try {
			const installation = await this.integrations.install(name, config);
			toast.success(successMessage ?? `Connected ${installation.attributes.displayName}.`);
			return true;
		} catch (e) {
			this.installFailedName = name;
			this.installError = e as ErrorModel;
			return false;
		} finally {
			this.installPendingName = undefined;
		}
	}

	// Disconnecting deletes the Rezible connection only; the provider's app stays installed.
	removalCandidate = $state.raw<IntegrationInstallation>();
	removing = $state(false);
	removeError = $state.raw<ErrorModel>();

	requestRemoval = (installation: IntegrationInstallation) => {
		this.removalCandidate = installation;
		this.removeError = undefined;
	};

	cancelRemoval = () => {
		if (this.removing) return;
		this.removalCandidate = undefined;
		this.removeError = undefined;
	};

	confirmRemoval = async () => {
		const installation = this.removalCandidate;
		if (!installation || this.removing) return;

		this.removing = true;
		this.removeError = undefined;
		try {
			await this.integrations.remove(installation.id);
			this.removalCandidate = undefined;
			toast.success(`Disconnected ${installation.attributes.displayName}.`);
		} catch (e) {
			this.removeError = e as ErrorModel;
		} finally {
			this.removing = false;
		}
	};
}

const ctx = new Context<IntegrationProviderController>("IntegrationProviderController");
export const initIntegrationProviderController = (providerNameFn: Getter<string>) =>
	ctx.set(new IntegrationProviderController(providerNameFn));
export const useIntegrationProviderController = () => ctx.get();
