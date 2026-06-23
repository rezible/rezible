import {
	type ErrorModel,
	getInstallableIntegrationsOptions,
	type InstallableIntegration,
	listIntegrationInstallationsOptions,
	installIntegrationMutation,
	type InstallIntegrationRequestAttributes,
	type IntegrationInstallation,
	listIntegrationInstallTargetsOptions,
	type IntegrationInstallTarget,
	installIntegrationFromTargetsMutation,
	updateIntegrationInstallationMutation,
	type UpdateIntegrationInstallationRequestAttributes,
} from "$lib/api";

import { useUserSessionState } from "$lib/user-session.svelte";

import { SvelteMap } from "svelte/reactivity";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export type IntegrationProviderDisplayInfo = {
	displayName: string;
	description: string;
}

export type IntegrationProvider = IntegrationProviderDisplayInfo & {
	name: string;
	integrations: InstallableIntegration[];
}

export const providerDisplays = new Map<string, IntegrationProviderDisplayInfo>([
	["demo", {displayName: "Demo", description: "Demo Data"}],
	["slack", {displayName: "Slack", description: "Slack Integration"}],
	["google", {displayName: "Google", description: "Google Integration"}],
	["github", {displayName: "Github", description: "Github Integration"}],
]);

const makeIntegrationProviderDisplay = (intg: InstallableIntegration): IntegrationProviderDisplayInfo => ({
	displayName: intg.displayName, 
	description: intg.description,
});

export class IntegrationsController {
	session = useUserSessionState();

	private listAvailableQuery = createQuery(() => getInstallableIntegrationsOptions());
	available = $derived(this.listAvailableQuery.data?.data.toSorted((a, b) => a.name.localeCompare(b.name)) || []);
	availableByName = $derived(new Map(this.available.map(a => [a.name, a])));
	availableByProvider = $derived.by(() => {
		const grouped = new SvelteMap<string, InstallableIntegration[]>();
		for (const intg of this.available) {
			const curr = grouped.get(intg.provider) ?? [];
			grouped.set(intg.provider, [...curr, intg]);
		}
		return grouped;
	});

	providers = $derived.by<IntegrationProvider[]>(() => {
		if (!this.available || this.available.length === 0) return [];
		const nameMap = new Map<string, Set<InstallableIntegration>>();
		this.available.forEach(inst => {
			nameMap.getOrInsert(inst.provider, new Set()).add(inst)
		});
		const provs = nameMap.entries().flatMap(([name, intgs]) => {
			const knownDisp = providerDisplays.get(name);
			const integrations = intgs.values().toArray();
			const displayInfos = !!knownDisp ? [knownDisp] : integrations.map(makeIntegrationProviderDisplay);
			return displayInfos.map(disp => ({name, integrations, ...disp}));
		});
		return provs.toArray();
	});

	private listInstalledQuery = createQuery(() => listIntegrationInstallationsOptions());
	installed = $derived(this.listInstalledQuery.data?.data || []);
	installedById = $derived(new SvelteMap(this.installed.map((intg) => [intg.id, intg])));
	installationsByName = $derived.by(() => {
		const grouped = new SvelteMap<string, IntegrationInstallation[]>();
		for (const intg of this.installed) {
			const curr = grouped.get(intg.attributes.integrationName) ?? [];
			grouped.set(intg.attributes.integrationName, [...curr, intg]);
		}
		return grouped;
	});
	installationsByProvider = $derived.by(() => {
		const grouped = new Map<string, IntegrationInstallation[]>();
		for (const intg of this.installed) {
			const curr = grouped.get(intg.attributes.integrationName) ?? [];
			grouped.set(intg.attributes.providerName, [...curr, intg]);
		}
		return grouped;
	});

	refetchInstalled() {
		this.listInstalledQuery.refetch();
		this.installingName = undefined;
	}

	private onInstallationsMutated(installation?: IntegrationInstallation) {
		this.refetchInstalled();
	}

	private installMut = createMutation(() => ({
		...installIntegrationMutation({}),
		onSuccess: ({data}) => {
			this.onInstallationsMutated(data);
		},
	}));

	private updateInstalledMut = createMutation(() => ({
		...updateIntegrationInstallationMutation({}),
		onSuccess: ({data}) => {
			this.onInstallationsMutated(data);
		},
	}));

	private listInstallTargetsQuery = createQuery(() => listIntegrationInstallTargetsOptions());
	private installationTargets = $derived(this.listInstallTargetsQuery.data?.data || []);
	installationTargetsByName = $derived.by(() => {
		const nameTargets = new Map<string, IntegrationInstallTarget[]>();
		this.installationTargets.forEach(t => {
			const curr = nameTargets.get(t.integrationName) || [];
			nameTargets.set(t.integrationName, [...curr, t]);
		});
		return nameTargets;
	});

    private selectIntegrationInstallTargetMut = createMutation(() => ({
        ...installIntegrationFromTargetsMutation({}),
        onSuccess: () => {
            this.refetchInstalled();
			this.listInstallTargetsQuery.refetch();
        },
    }));

	installationPending = $derived(this.installMut.isPending || this.selectIntegrationInstallTargetMut.isPending);
	installingName = $derived(this.installMut.variables?.path?.name || this.selectIntegrationInstallTargetMut.variables?.path.name);
	installationErr = $derived(this.installMut.error || this.selectIntegrationInstallTargetMut.error);

	async installNew(name: string, attributes: InstallIntegrationRequestAttributes) {
		await this.installMut.mutateAsync({path: { name }, body: { attributes }});
	}

	async updateInstallation(id: string, attributes: UpdateIntegrationInstallationRequestAttributes) {
		await this.updateInstalledMut.mutateAsync({path: { id }, body: { attributes }});
	}

	async installFromTargets(name: string, externalRefs: string[]) {
		if (this.installationPending || externalRefs.length === 0) return;
		try {
			const attributes = { externalRefs }
			await this.selectIntegrationInstallTargetMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
		} catch (e) {
			console.error("failed to install targets", e);
		}
	}

	loading = $derived(this.listAvailableQuery.isPending || this.listInstalledQuery.isPending || this.listInstallTargetsQuery.isPending);
	error = $derived((this.listAvailableQuery.error ?? this.listInstalledQuery.error ?? this.listInstallTargetsQuery.error) as ErrorModel | null);
}

const ctx = new Context<IntegrationsController>("IntegrationsController");
export const initIntegrationsController = () => ctx.set(new IntegrationsController());
export const useIntegrationsController = () => ctx.get();
