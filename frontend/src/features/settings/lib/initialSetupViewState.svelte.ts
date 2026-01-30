import {
    listSupportedIntegrationsOptions,
    listConfiguredIntegrationsOptions,
    type ConfiguredIntegration,
    type ConfiguredIntegrationAttributes,
    configureIntegrationMutation,
    finishOrganizationSetupMutation, 
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
    type SupportedIntegration,
    type ConfigureIntegrationRequestBody,
} from "$lib/api";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { SvelteMap, SvelteSet } from "svelte/reactivity";
import { z } from "zod";

export const oauthCallbackParamsSchema = z.object({
    name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

class OAuthIntegrationSetupState {
    private callbackParams = useSearchParams(oauthCallbackParamsSchema);
    private callbackName = $derived(this.callbackParams.name);

    private onCompleted: () => void;

    constructor(onCompleted: () => void) {
        watch(() => this.callbackName, name => { this.onCallbackSet(name) });
        this.onCompleted = onCompleted;
    }

    private startFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
    loadingFlowUrl = $derived(this.startFlowMut.isPending);
    startFlowUrl = $derived(this.startFlowMut.data?.data.flow_url);
    startFlowErr = $derived(this.startFlowMut.error);

    async startFlow(name: string) {
        try {
            const resp = await this.startFlowMut.mutateAsync({ path: { name } });
            window.location.assign(new URL(resp.data.flow_url));
        } catch (e) {
            console.error("failed to complete", e);
        }
    }

    private completeFlowMut = createMutation(() => completeIntegrationOauthFlowMutation({}));
    completingFlow = $derived(this.completeFlowMut.isPending);
    completeFlowErr = $derived(this.completeFlowMut.error);

    private async onCallbackSet(name?: string) {
        if (!name || this.completingFlow) return;

        const { state, code } = $state.snapshot(this.callbackParams);
        this.callbackParams.reset();

        if (!state || !code) return;

        try {
            const resp = await this.completeFlowMut.mutateAsync({ path: { name }, body: { attributes: { state, code } } });
            console.log("completed", resp);
            this.onCompleted();
        } catch (e) {
            console.error("failed to complete", e);
        }
    }
};

const dataKinds = [
    {name: "Chat", kind: "chat", required: true},
    {name: "Users", kind: "users", required: true},
];

const getDataKindIntegrationsMap = (intgs: SupportedIntegration[]) => {
    const kindsMap = new SvelteMap<string, Set<SupportedIntegration>>();
    intgs.forEach(intg => {
        intg.supportedDataKinds.forEach(kind => {
            const intgNames = (kindsMap.get(kind) || new Set<SupportedIntegration>());
            kindsMap.set(kind, intgNames.add(intg));
        });
    })
    return kindsMap;
}

export class InitialSetupViewState {
    session = useAuthSessionState();

    oauth: OAuthIntegrationSetupState;

    private supportedIntegrationsQuery = createQuery(() => listSupportedIntegrationsOptions());
    supportedIntegrations = $derived(this.supportedIntegrationsQuery.data?.data || []);
    supportedIntegrationsMap = $derived(new SvelteMap(this.supportedIntegrations.map(intg => ([intg.name, intg]))));

    private dataKindIntegrationSupportMap = $derived(getDataKindIntegrationsMap(this.supportedIntegrations));

    private configuredIntegrationsQuery = createQuery(() => listConfiguredIntegrationsOptions());
    configuredIntegrations = $derived(this.configuredIntegrationsQuery.data?.data || []);
	configuredIntegrationsMap = $derived(new SvelteMap(this.configuredIntegrations.map(intg => ([intg.name, intg]))));

    loadingIntegrations = $derived(this.supportedIntegrationsQuery.isPending || this.configuredIntegrationsQuery.isPending);

    requiredDataKinds = new Set(dataKinds.filter(k => k.required).map(k => k.kind));
	configuredEnabledDataKinds = $derived(new Set(this.configuredIntegrations.flatMap(intg => intg.attributes.enabledDataKinds)));
    remainingRequiredDataKinds = $derived(this.requiredDataKinds.difference(this.configuredEnabledDataKinds));
    nextRequiredDataKind = $derived(this.remainingRequiredDataKinds.values().next().value);
    nextRequiredSupportedIntegrations = $derived.by(() => {
        if (!this.nextRequiredDataKind) return [];
        const intgs = this.dataKindIntegrationSupportMap.get(this.nextRequiredDataKind);
        if (!intgs) return []; // this shouldn't happen - we don't have any supported integrations for this (required) data kind?
        return intgs.values().toArray();
    });
    // are there configured integrations that support this data kind?
    nextRequiredSupportedConfiguredIntegrations = $derived.by(() => {
        const reqKind = $state.snapshot(this.nextRequiredDataKind);
        if (!reqKind) return;
        return this.configuredIntegrations.filter(intg => intg.attributes.enabledDataKinds.includes(reqKind));
    });

    constructor() {
        this.oauth = new OAuthIntegrationSetupState(() => { this.configuredIntegrationsQuery.refetch() });
    }

    private configureIntegrationMut = createMutation(() => ({
        ...configureIntegrationMutation({}),
        onSuccess: () => {
            this.configuredIntegrationsQuery.refetch();
        }
    }));
    configuringIntegration = $derived(this.configureIntegrationMut.isPending);
    configureIntegrationMutErr = $derived(this.configureIntegrationMut.error);

    async doConfigureIntegration(name: string, attributes: ConfigureIntegrationRequestBody["attributes"]) {
        this.configureIntegrationMut.mutateAsync({
            path: { name },
            body: { attributes }
        })
    }

    private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    finishingSetup = $derived(this.finishOrgSetupMut.isPending);
    async doFinishOrganizationSetup() {
        const id = this.session.org?.id;
        if (!id) return;
        await this.finishOrgSetupMut.mutateAsync({ path: { id } });
        this.session.refetch();
    }

    loading = $derived(this.finishingSetup || this.configuringIntegration || this.loadingIntegrations)
}

const ctx = new Context<InitialSetupViewState>("initialSetupView");
export const setInitialSetupViewState = () => ctx.set(new InitialSetupViewState());
export const useInitialSetupViewState = () => ctx.get();