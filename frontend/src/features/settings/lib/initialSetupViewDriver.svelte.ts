import {
    listSupportedIntegrationsOptions,
    listConfiguredIntegrationsOptions,
    configureIntegrationMutation,
    finishOrganizationSetupMutation,
    type SupportedIntegration,
    type ConfigureIntegrationRequestBody,
} from "$lib/api";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteMap } from "svelte/reactivity";
import { IntegrationOAuthSetupDriver } from "./integrationOAuthSetupDriver.svelte";

const dataKinds = [
    {name: "Chat", kind: "chat", required: true},
    {name: "Users", kind: "users", required: true},
];

class InitialIntegrationsSetupDriver {
    constructor() {
        this.oauth = new IntegrationOAuthSetupDriver(() => { 
            this.listConfiguredQuery.refetch();
        });
    }
    oauth: IntegrationOAuthSetupDriver;

    private listSupportedQuery = createQuery(() => listSupportedIntegrationsOptions());
    supported = $derived(this.listSupportedQuery.data?.data || []);
    supportedMap = $derived(new SvelteMap(this.supported.map(intg => ([intg.name, intg]))));

    private listConfiguredQuery = createQuery(() => listConfiguredIntegrationsOptions());
    configured = $derived(this.listConfiguredQuery.data?.data || []);
	configuredMap = $derived(new SvelteMap(this.configured.map(intg => ([intg.name, intg]))));

    requiredDataKinds = new Set(dataKinds.filter(k => k.required).map(k => k.kind));
	configuredEnabledDataKinds = $derived(new Set(this.configured.flatMap(intg => intg.attributes.enabledDataKinds)));
    remainingRequiredDataKinds = $derived(this.requiredDataKinds.difference(this.configuredEnabledDataKinds).values().toArray());
    nextRequiredDataKind = $derived(this.remainingRequiredDataKinds.at(0));
    nextRequiredSupportedIntegrations = $derived.by(() => {
        const reqKind = this.nextRequiredDataKind;
        if (!reqKind) return [];
        const intgSupport = this.supported.filter(intg => 
            intg.supportedDataKinds.includes(reqKind));
        if (!intgSupport) return []; // this shouldn't happen - we don't have any supported integrations for this (required) data kind?
        return intgSupport;
    });
    // are there configured integrations that support this data kind?
    nextRequiredSupportedConfiguredIntegrations = $derived.by(() => {
        const reqKind = $state.snapshot(this.nextRequiredDataKind);
        if (!reqKind) return;
        return this.configured.filter(intg => intg.attributes.enabledDataKinds.includes(reqKind));
    });

    private configureMut = createMutation(() => ({
        ...configureIntegrationMutation({}),
        onSuccess: () => { this.listConfiguredQuery.refetch() }
    }));
    configureMutErr = $derived(this.configureMut.error);

    async doConfigure(name: string, attributes: ConfigureIntegrationRequestBody["attributes"]) {
        this.configureMut.mutateAsync({
            path: { name },
            body: { attributes }
        })
    }

    isLoading = $derived(this.listSupportedQuery.isPending || this.listConfiguredQuery.isPending);
    isConfiguring = $derived(this.configureMut.isPending);
}

export class InitialSetupViewDriver {
    session = useAuthSessionState();

    constructor() {
        this.integrations = new InitialIntegrationsSetupDriver();
    }

    integrations: InitialIntegrationsSetupDriver;

    canFinish = $derived.by(() => {
        if (!this.integrations) return false;
        if (this.integrations.remainingRequiredDataKinds.length === 0) return true;
        return false;
    });

    private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    async doFinishOrganizationSetup() {
        const id = this.session.org?.id;
        if (!id) return;
        await this.finishOrgSetupMut.mutateAsync({ path: { id } });
        this.session.refetch();
    }
    finishingOrgSetup = $derived(this.finishOrgSetupMut.isPending);
}

const ctx = new Context<InitialSetupViewDriver>("InitialSetupViewDriver");
export const initInitialSetupViewDriver = () => ctx.set(new InitialSetupViewDriver());
export const useInitialSetupViewDriver = () => ctx.get();