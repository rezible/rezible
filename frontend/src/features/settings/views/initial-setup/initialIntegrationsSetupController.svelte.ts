import { watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { SvelteMap } from "svelte/reactivity";
import { createQuery, createMutation } from "@tanstack/svelte-query";
import {
    listAvailableIntegrationsOptions,
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
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
			const callbackPath = `/setup/callback/${name}`;
			const resp = await this.startFlowMut.mutateAsync({ path: { name }, body: { attributes: { callbackPath } } });
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

        const { state, code } = this.callbackParams;
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

const requiredDataKinds = [
    {name: "Chat", kind: "chat"},
    {name: "Users", kind: "users"},
];
const requiredKinds = new Set(requiredDataKinds.map(k => k.kind));

const getEnabledDataKinds = (s: {[name: string]: boolean}) => {
    return Object.entries(s).filter(([_, enabled]) => (enabled)).map(([name, _]) => (name));
}

export class InitialIntegrationsSetupController {
    oauth: IntegrationOAuthSetupController;
    constructor() {
        this.oauth = new IntegrationOAuthSetupController(() => { 
            this.listConfiguredQuery.refetch();
        });
    }

    private listAvailableQuery = createQuery(() => listAvailableIntegrationsOptions());
    available = $derived(this.listAvailableQuery.data?.data || []);
    availableMap = $derived(new SvelteMap(this.available.map(intg => ([intg.name, intg]))));

    private listConfiguredQuery = createQuery(() => listConfiguredIntegrationsOptions());
    configured = $derived(this.listConfiguredQuery.data?.data || []);
	configuredMap = $derived(new SvelteMap(this.configured.map(intg => ([intg.name, intg]))));
	configuredDataKinds = $derived(new Set(this.configured.flatMap(intg => getEnabledDataKinds(intg.attributes.dataKinds))));

    remainingRequiredDataKinds = $derived(requiredKinds.difference(this.configuredDataKinds).values().toArray());
    nextRequiredDataKind = $derived(this.remainingRequiredDataKinds.at(0));
    availableDataKindIntegrations = $derived.by(() => {
        const reqKind = this.nextRequiredDataKind;
        if (!reqKind) return [];
        return this.available.filter(intg => intg.dataKinds.includes(reqKind));
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

    isLoading = $derived(this.listAvailableQuery.isPending || this.listConfiguredQuery.isPending);
    isConfiguring = $derived(this.configureMut.isPending);
};