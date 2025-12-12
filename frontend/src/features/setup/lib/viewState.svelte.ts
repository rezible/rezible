import { completeIntegrationOauthMutation, finishOrganizationSetupMutation, listIntegrationsOptions, startIntegrationOauthMutation, type CompleteIntegrationOAuthRequestAttributes } from "$src/lib/api";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { SvelteSet } from "svelte/reactivity";
import { z } from "zod";
 
export const setupIntegrationParamsSchema = z.object({
	providerId: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

const getProviderKind = (id: string) => {
    if (id === "slack") return "chat";
    throw new Error("invalid provider id");
}

export class SetupViewState {
    session = useAuthSessionState();
    private searchParams = useSearchParams(setupIntegrationParamsSchema);
    private queryProviderId = $derived(this.searchParams.providerId);
    private queryCode = $derived(this.searchParams.code);
    private queryState = $derived(this.searchParams.state);

    settingUpProvider = $state<string>();

    private integrationsQuery = createQuery(() => listIntegrationsOptions({}));
    loading = $derived(this.integrationsQuery.isFetching);
    private integrations = $derived(this.integrationsQuery.data?.data);
    private enabledProviderIntegrationIds = $derived(this.integrations?.filter(intg => intg.attributes.enabled).map(intg => intg.attributes.provider_id) ?? []);
    private enabledProviderIdMap = $derived(new SvelteSet(this.enabledProviderIntegrationIds));

    constructor() {
        watch(() => this.queryProviderId, id => {this.onProviderIdSet(id)});
        watch(() => this.enabledProviderIdMap, ids => {this.onEnabledProvidersUpdated(ids)})
    }

    nextRequiredIntegrationId = $state<string>();
    private startIntegrationOAuthMut = createMutation(() => startIntegrationOauthMutation({}));
    nextRequiredIntegrationFlowUrl = $derived(this.startIntegrationOAuthMut.data?.data.flow_url);
    nextRequiredIntegrationFlowErr = $derived(this.startIntegrationOAuthMut.error);

    async onEnabledProvidersUpdated(ids: Set<string>) {
        this.nextRequiredIntegrationId = undefined;
        if (!ids.has("slack")) this.nextRequiredIntegrationId = "slack";
        // iterate in order

        try {
            if (this.nextRequiredIntegrationId) await this.doStartIntegrationOAuth(this.nextRequiredIntegrationId);
        } catch (e) {}
    }

    async doStartIntegrationOAuth(provider_id: string) {
        if (this.loading || !!this.queryProviderId || !!this.settingUpProvider) return;
        
        const resp = await this.startIntegrationOAuthMut.mutateAsync({body: {attributes: {provider_id, kind: getProviderKind(provider_id)}}});
        // console.log("start", resp.data);
    }

    private completeIntegrationOAuthMut = createMutation(() => completeIntegrationOauthMutation({}));
    completeIntegrationErr = $derived(this.completeIntegrationOAuthMut.error);

    async onProviderIdSet(provider_id?: string) {
        if (!provider_id) return;
        if (this.completeIntegrationOAuthMut.isPending) return;

        this.settingUpProvider = provider_id;

        const state = $state.snapshot(this.queryState);
        const code = $state.snapshot(this.queryCode);

        console.log("do complete", {state, code});

        this.searchParams.reset();
        if (!state || !code) return;

        try {
            const kind = getProviderKind(provider_id);
            await this.doCompleteIntegrationOAuth({provider_id, kind, state, code});
        } catch (e) {
            console.error("failed to complete", e);
        } finally {
            this.settingUpProvider = undefined;
        }
    }

    async doCompleteIntegrationOAuth(attributes: CompleteIntegrationOAuthRequestAttributes) {
        const resp = await this.completeIntegrationOAuthMut.mutateAsync({body: {attributes}});
        console.log("completed", resp);
        this.integrationsQuery.refetch();
    }

	private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    isFinishingSetup = $derived(this.finishOrgSetupMut.isPending);
    async doFinishOrganizationSetup() {
        const id = this.session.org?.id;
        if (!id) return;
		await this.finishOrgSetupMut.mutateAsync({path: {id}});
		this.session.refetch();
    }
}

const ctx = new Context<SetupViewState>("setupView");
export const setSetupViewState = () => ctx.set(new SetupViewState());
export const useSetupViewState = () => ctx.get();