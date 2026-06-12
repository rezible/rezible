import {
	listAvailableIntegrationsOptions,
	listInstalledIntegrationsOptions,
	createInstalledIntegrationMutation,
	startIntegrationOauthFlowMutation, 
	completeIntegrationOauthFlowMutation,
	type CreateInstalledIntegrationRequestBody,
	type InstalledIntegration,
	type ErrorModel,
	listIntegrationInstallTargetsOptions,
	type IntegrationInstallTarget,
	type AvailableIntegration,
	installIntegrationTargetsMutation,
	type IntegrationOAuthInstallResult,
} from "$lib/api";

import { page } from "$app/state";
import { clearQueryParams } from "$lib/utils";
import { useUserSessionState } from "$lib/user-session.svelte";

import { SvelteMap } from "svelte/reactivity";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { type ConfigureIntegrationDialogParams } from "../components/configure-integration-dialog/controller.svelte";
import { tick } from "svelte";

export class IntegrationOAuthController {
    inFlowForName = $state<string>();
    error = $state<ErrorModel>();

	private onSuccess: (res: IntegrationOAuthInstallResult) => void;

    constructor(onSuccess: (res: IntegrationOAuthInstallResult) => void) {
		this.onSuccess = onSuccess;
        watch(() => page.url.search, search => {
            this.checkOAuthCallback(new URLSearchParams(search));
        });
    }

    private setError(err: unknown) {
        this.error = {
            title: "Integration Setup Failed",
            detail: err instanceof Error ? err.message : "An unknown issue occurred",
        };
    };

    clearFlow() {
        this.inFlowForName = undefined;
        this.error = undefined;
    }

    private startOAuthFlowMut = createMutation(() => ({
		...startIntegrationOauthFlowMutation({}),
	}));

    async startFlowFor(name: string) {
        this.inFlowForName = name;
        const resp = await this.startOAuthFlowMut.mutateAsync({path: { name }});
        window.location.assign(new URL(resp.data.flow_url));
    }

    private completeOAuthFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: async ({data}) => {
			await tick();
			this.onSuccess?.(data);
        },
    }));

	private async checkOAuthCallback(params: URLSearchParams) {
		const name = params.get("name");
		const code = params.get("code");
		const state = params.get("state");

		if (this.completeOAuthFlowMut.isPending) return;
		if (!name || !state || !code) return;

        this.inFlowForName = name;

		await clearQueryParams();

		try {
			const attributes = { state, code };
			await this.completeOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			this.error = undefined;
		} catch (e) {
			this.setError(e);
		} finally {
            this.inFlowForName = undefined;
        }
	}

    inFlow = $derived(this.startOAuthFlowMut.isPending || this.completeOAuthFlowMut.isPending);
};

export const getEnabledCapabilties = (installed: InstalledIntegration[]) =>
	installed.flatMap(intg => Object.entries(intg.attributes.capabilities)
		.filter(([_, enabled]) => enabled)
		.map(([name, _]) => name));

export class IntegrationsController {
	oauth = new IntegrationOAuthController(res => {this.onOAuthResult(res)});;
	session = useUserSessionState();

	dataSyncDialogInstallation = $state<InstalledIntegration>();
	configureDialogParams = $state.raw<ConfigureIntegrationDialogParams>();

	openConfigureDialog(integration: AvailableIntegration, installation?: InstalledIntegration) {
		this.configureDialogParams = {integration, installation};
	}

	private listAvailableQuery = createQuery(() => listAvailableIntegrationsOptions());
	available = $derived(this.listAvailableQuery.data?.data.toSorted((a, b) => a.name.localeCompare(b.name)) || []);
	availableByName = $derived(new Map(this.available.map(a => [a.name, a])));
	availableByProvider = $derived.by(() => {
		const grouped = new SvelteMap<string, AvailableIntegration[]>();
		for (const intg of this.available) {
			const curr = grouped.get(intg.provider) ?? [];
			grouped.set(intg.provider, [...curr, intg]);
		}
		return grouped;
	});
	availableProviders = $derived(this.availableByProvider.keys().toArray());

	private oauthTarget = $derived(!!this.oauth.inFlowForName ? this.availableByName.get(this.oauth.inFlowForName) : undefined);
	constructor() {
		watch(() => this.oauthTarget, oauthIntegration => {
			if (!oauthIntegration) return;
			this.openConfigureDialog(oauthIntegration);
		});
	}

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

	private refetchInstalled() {
		this.listInstalledQuery.refetch();
		this.installingName = undefined;
		this.configureDialogParams = undefined;
	}

	private onOAuthResult(res: IntegrationOAuthInstallResult) {
		const name = res.installed?.at(0)?.attributes.integrationName || res.installTargetOptions?.at(0)?.integrationName;
		this.refetchInstalled();
		if (!name || (!!res.installed && res.installed.length > 1)) return;
		const integration = this.availableByName.get(name);
		const installed = res.installed?.length === 1 ? res.installed.at(0) : undefined;
		if (integration) this.openConfigureDialog(integration, installed);
	}

	private installMut = createMutation(() => ({
		...createInstalledIntegrationMutation({}),
		onSuccess: ({data}) => {
			const dialogParams = $state.snapshot(this.configureDialogParams);
			this.refetchInstalled();
			if (!!dialogParams?.integration) this.openConfigureDialog(dialogParams.integration, data);
		},
	}));

    private selectIntegrationInstallTargetMut = createMutation(() => ({
        ...installIntegrationTargetsMutation({}),
        onSuccess: () => {
            this.refetchInstalled();
			this.listInstallTargetsQuery.refetch();
        },
    }));

	installationPending = $derived(this.installMut.isPending || this.selectIntegrationInstallTargetMut.isPending);
	installingName = $derived(this.installMut.variables?.path?.name || this.selectIntegrationInstallTargetMut.variables?.path.name);
	installationErr = $derived(this.installMut.error || this.selectIntegrationInstallTargetMut.error);

	async installNew(name: string, attributes: CreateInstalledIntegrationRequestBody["attributes"]) {
		await this.installMut.mutateAsync({path: { name }, body: { attributes }});
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
