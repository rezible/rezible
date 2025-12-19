import { completeIntegrationOauthMutation, finishOrganizationSetupMutation, listIntegrationsOptions, startIntegrationOauthMutation, type CompleteIntegrationOAuthRequestAttributes, type Integration } from "$src/lib/api";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
 
export const callbackParamsSchema = z.object({
	name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

export class SetupViewState {
    session = useAuthSessionState();

    private callbackParams = useSearchParams(callbackParamsSchema);
    private callbackName = $derived(this.callbackParams.name);

    private integrationsQuery = createQuery(() => listIntegrationsOptions({}));
    loading = $derived(this.integrationsQuery.isFetching);
    private integrations = $derived(this.integrationsQuery.data?.data);
    private enabledIntegrations = $derived(this.integrations?.filter(intg => intg.attributes.enabled) ?? []);

    constructor() {
        watch(() => this.enabledIntegrations, intgs => {this.onEnabledIntegrationsUpdated(intgs)});
        watch(() => this.callbackName, name => {this.onCallbackIntegrationSet(name)});
    }

    currentlyCompleting = $state<string>();
    nextRequired = $state<string>();

    private startIntegrationOAuthMut = createMutation(() => startIntegrationOauthMutation({}));
    nextRequiredIntegrationFlowUrl = $derived(this.startIntegrationOAuthMut.data?.data.flow_url);
    nextRequiredIntegrationFlowErr = $derived(this.startIntegrationOAuthMut.error);

    async onEnabledIntegrationsUpdated(intgs: Integration[]) {
        if (this.loading || !!this.currentlyCompleting) return;

        this.nextRequired = undefined;
        
        const enabledNames = new Set(intgs.map(int => int.attributes.name));

        if (!enabledNames.has("slack")) this.nextRequired = "slack";

        if (!this.nextRequired) return;

        try {
            const resp = await this.startIntegrationOAuthMut.mutateAsync({body: {attributes: {name: this.nextRequired}}});
            console.log("start", resp.data);
        } catch (e) {
            console.log("failed to start", e);
        }
    };

    private completeIntegrationOAuthMut = createMutation(() => completeIntegrationOauthMutation({}));
    completeIntegrationErr = $derived(this.completeIntegrationOAuthMut.error);

    async onCallbackIntegrationSet(name?: string) {
        if (!name || this.completeIntegrationOAuthMut.isPending) return;

        this.currentlyCompleting = name;

        const {state, code} = $state.snapshot(this.callbackParams);

        console.log("do complete", name, {state, code});

        this.callbackParams.reset();
        if (!state || !code) return;

        try {
            await this.doCompleteIntegrationOAuth({name, state, code});
        } catch (e) {
            console.error("failed to complete", e);
        } finally {
            this.currentlyCompleting = undefined;
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