import { watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { SvelteMap } from "svelte/reactivity";
import { createQuery, createMutation } from "@tanstack/svelte-query";
import {
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
    listSupportedIntegrationsOptions,
    listConfiguredIntegrationsOptions,
    configureIntegrationMutation,
    type ConfigureIntegrationRequestBody,
} from "$lib/api";

export const oauthCallbackParamsSchema = z.object({
    name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

class IntegrationOAuthSetupController {
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

export class InitialIntegrationsSetupController {
    constructor() {
        this.oauth = new IntegrationOAuthSetupController(() => { 
            this.listConfiguredQuery.refetch();
        });
    }
    oauth: IntegrationOAuthSetupController;

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
};