import {
	listAvailableIntegrationsOptions,
	listInstalledIntegrationsOptions,
	createInstalledIntegrationMutation,
	type CreateInstalledIntegrationRequestBody,
	type InstalledIntegration,
	type ErrorModel,
	listIntegrationInstallTargetsOptions,
	type IntegrationInstallTarget,
	type AvailableIntegration,
	installIntegrationTargetsMutation,
} from "$lib/api";
import { useAuthSessionState } from "$lib/auth-session.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteMap } from "svelte/reactivity";

export class IntegrationsController {
	session = useAuthSessionState();

	private listAvailableQuery = createQuery(() => listAvailableIntegrationsOptions());
	available = $derived(this.listAvailableQuery.data?.data || []);
	availableByProvider = $derived.by(() => {
		const grouped = new SvelteMap<string, AvailableIntegration[]>();
		for (const intg of this.available) {
			const curr = grouped.get(intg.provider) ?? [];
			grouped.set(intg.provider, [...curr, intg]);
		}
		return grouped;
	});

	private listInstalledQuery = createQuery(() => listInstalledIntegrationsOptions());
	installed = $derived(this.listInstalledQuery.data?.data || []);
	installedById = $derived(new SvelteMap(this.installed.map((intg) => [intg.id, intg])));
	installationsByName = $derived.by(() => {
		const grouped = new SvelteMap<string, InstalledIntegration[]>();
		for (const intg of this.installed) {
			const curr = grouped.get(intg.attributes.integrationName) ?? [];
			grouped.set(intg.attributes.integrationName, [...curr, intg]);
		}
		return grouped;
	});

	refetchInstalled() {
		this.listInstalledQuery.refetch();
	}

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

	private installMut = createMutation(() => ({
		...createInstalledIntegrationMutation({}),
		onSuccess: () => {
			this.refetchInstalled();
		},
	}));

    private selectIntegrationInstallTargetMut = createMutation(() => ({
        ...installIntegrationTargetsMutation({}),
        onSuccess: () => {
            this.refetchInstalled();
        },
    }));

	installingName = $derived(this.installMut.variables?.path?.name 
		|| this.selectIntegrationInstallTargetMut.variables?.path.name);
	private installationErr = $derived(this.installMut.error 
		|| this.selectIntegrationInstallTargetMut.error);
	isInstalling = $derived(this.installMut.isPending 
		|| this.selectIntegrationInstallTargetMut.isPending);

	installError = $derived(this.installationErr?.detail || this.installationErr?.title || "");

	async installNew(name: string, attributes: CreateInstalledIntegrationRequestBody["attributes"]) {
		await this.installMut.mutateAsync({path: { name }, body: { attributes }});
	}

	async installFromTargets(name: string, externalRefs: string[]) {
		if (this.isInstalling || externalRefs.length === 0) return;
		try {
			const attributes = { externalRefs }
			await this.selectIntegrationInstallTargetMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
            this.installingName = undefined;
		} catch (e) {
			console.error("failed to install targets", e);
		}
	}

	errorFor(name: string) {
		if (this.installingName !== name) return "";
		return this.installError;
	}

	loading = $derived(this.listAvailableQuery.isPending || this.listInstalledQuery.isPending);
	error = $derived((this.listAvailableQuery.error ?? this.listInstalledQuery.error) as ErrorModel | null);
}

const ctx = new Context<IntegrationsController>("IntegrationsController");
export const initIntegrationsController = () => ctx.set(new IntegrationsController());
export const useIntegrationsController = () => ctx.get();
