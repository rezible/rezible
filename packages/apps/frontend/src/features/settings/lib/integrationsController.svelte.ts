import {
	type ApiError,
	getInstallableIntegrationsOptions,
	type InstallableIntegration,
	listIntegrationInstallationsOptions,
	installIntegrationMutation,
	type IntegrationInstallation,
	listIntegrationInstallTargetsOptions,
	type IntegrationInstallTarget,
	installIntegrationFromTargetsMutation,
	updateIntegrationInstallationMutation,
	deleteIntegrationInstallationMutation,
} from "$lib/api";

import { useUserSessionState } from "$lib/user-session.svelte";

import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { capabilityPills, connectionStatus } from "./integrationCatalogue";
import { getIntegrationProviderInfo, type IntegrationProviderInfo } from "./integrationProviders";

type UserSettings = Record<string, unknown>;

export type IntegrationProviderSummary = IntegrationProviderInfo & {
	installable: InstallableIntegration[];
	installations: IntegrationInstallation[];
	capabilityPills: string[];
	// What is connected, or undefined when nothing is.
	status?: string;
};

export class IntegrationsController {
	private session = useUserSessionState();

	private installableQuery = createQuery(() => getInstallableIntegrationsOptions());
	private installedQuery = createQuery(() => listIntegrationInstallationsOptions());
	private installTargetsQuery = createQuery(() => listIntegrationInstallTargetsOptions());

	installable = $derived(this.installableQuery.data?.data ?? []);
	installed = $derived(this.installedQuery.data?.data ?? []);
	private installTargets = $derived(this.installTargetsQuery.data?.data ?? []);

	installedCapabilities = $derived(new Set(this.installed.flatMap((intg) => intg.attributes.capabilities)));

	providers = $derived.by<IntegrationProviderSummary[]>(() => {
		const providerNames = new Set<string>();
		for (const intg of this.installable) {
			providerNames.add(intg.provider);
		}
		for (const installation of this.installed) {
			providerNames.add(installation.attributes.provider);
		}

		const summaries = Array.from(providerNames, (name) => {
			const installable = this.installable.filter((intg) => intg.provider === name);
			const installations = this.installed.filter((intg) => intg.attributes.provider === name);
			const info = getIntegrationProviderInfo(name, installable);
			const catalogueInput = { ...info, installable, installations };
			return {
				...catalogueInput,
				capabilityPills: capabilityPills(catalogueInput),
				status: connectionStatus(catalogueInput),
			};
		});
		return summaries.toSorted((a, b) => a.displayName.localeCompare(b.displayName));
	});
	installedProviders = $derived(this.providers.filter((provider) => provider.installations.length > 0));
	availableProviders = $derived(this.providers.filter((provider) => provider.installations.length === 0));

	loading = $derived(this.installableQuery.isPending || this.installedQuery.isPending);
	error = $derived((this.installableQuery.error ?? this.installedQuery.error) as ApiError | null);
	installTargetsError = $derived(this.installTargetsQuery.error as ApiError | null);

	constructor() {
		// Some integrations are only installable while incident management is enabled.
		watch(
			() => this.session.orgPreferences?.enableIncidentManagement,
			() => {
				this.installableQuery.refetch();
			},
			{ lazy: true }
		);
	}

	getProvider(name: string) {
		return this.providers.find((provider) => provider.name === name);
	}

	installableIntegration(name: string) {
		return this.installable.find((intg) => intg.name === name);
	}

	installationsFor(name: string) {
		return this.installed.filter((intg) => intg.attributes.name === name);
	}

	installTargetsFor(name: string): IntegrationInstallTarget[] {
		return this.installTargets.filter((target) => target.resourceRef.providerNamespace === name);
	}

	canInstall(name: string) {
		const intg = this.installableIntegration(name);
		if (!intg) {
			return false;
		}
		if (intg.maxInstalls === undefined) {
			return true;
		}
		return this.installationsFor(name).length < intg.maxInstalls;
	}

	async refresh() {
		await Promise.all([
			this.installableQuery.refetch(),
			this.installedQuery.refetch(),
			this.installTargetsQuery.refetch(),
		]);
	}

	private installMut = createMutation(() => installIntegrationMutation());
	private installFromTargetsMut = createMutation(() => installIntegrationFromTargetsMutation());
	private updateMut = createMutation(() => updateIntegrationInstallationMutation());
	private deleteMut = createMutation(() => deleteIntegrationInstallationMutation());

	async install(name: string, config: Record<string, unknown>) {
		const resp = await this.installMut.mutateAsync({ path: { name }, body: { attributes: { config } } });
		await this.refresh();
		return resp.data;
	}

	async installFromTargets(name: string, resourceRefs: string[]) {
		const body = { attributes: { resourceRefs } };
		const resp = await this.installFromTargetsMut.mutateAsync({ path: { name }, body });
		await this.refresh();
		return resp.data;
	}

	async updateSettings(id: string, userSettings: UserSettings) {
		const resp = await this.updateMut.mutateAsync({
			path: { id },
			body: { attributes: { userSettings } },
		});
		await this.installedQuery.refetch();
		return resp.data;
	}

	async remove(id: string) {
		await this.deleteMut.mutateAsync({ path: { id } });
		await this.refresh();
	}
}

const ctx = new Context<IntegrationsController>("IntegrationsController");
export const initIntegrationsController = () => ctx.set(new IntegrationsController());
export const useIntegrationsController = () => ctx.get();
