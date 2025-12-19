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

type IntegrationTuple = {
    name: string;
    type: string;
}

const SlackIntegration: IntegrationTuple = {name: "slack", type: "chat"};

const getIntegrationByCallbackName = (name: string) => {
    switch (name) {
        case "slack": return SlackIntegration;
    }
}

export class SetupViewState {
    session = useAuthSessionState();

    private callbackParams = useSearchParams(callbackParamsSchema);
    private callbackName = $derived(this.callbackParams.name);
    private callbackIntegration = $derived(getIntegrationByCallbackName(this.callbackName));

    private integrationsQuery = createQuery(() => listIntegrationsOptions({}));
    loading = $derived(this.integrationsQuery.isFetching);
    private integrations = $derived(this.integrationsQuery.data?.data);
    private enabledIntegrations = $derived(this.integrations?.filter(intg => intg.attributes.enabled) ?? []);

    constructor() {
        watch(() => this.enabledIntegrations, intgs => {this.onEnabledIntegrationsUpdated(intgs)});
        watch(() => this.callbackIntegration, qi => {this.onCallbackIntegrationSet(qi)});
    }

    currentlyCompleting = $state<IntegrationTuple>();
    nextRequired = $state<IntegrationTuple>();

    private startIntegrationOAuthMut = createMutation(() => startIntegrationOauthMutation({}));
    nextRequiredIntegrationFlowUrl = $derived(this.startIntegrationOAuthMut.data?.data.flow_url);
    nextRequiredIntegrationFlowErr = $derived(this.startIntegrationOAuthMut.error);

    async onEnabledIntegrationsUpdated(intgs: Integration[]) {
        if (this.loading || !!this.currentlyCompleting) return;

        this.nextRequired = undefined;
        
        const isEnabled = (t: IntegrationTuple) => {
            return !!intgs.find(({attributes: attr}) => (attr.name === t.name && attr.type === t.type));
        }

        if (!isEnabled(SlackIntegration)) this.nextRequired = SlackIntegration;

        if (!this.nextRequired) return;

        try {
            const resp = await this.startIntegrationOAuthMut.mutateAsync({body: {attributes: this.nextRequired}});
            console.log("start", resp.data);
        } catch (e) {
            console.log("failed to start", e);
        }
    };

    private completeIntegrationOAuthMut = createMutation(() => completeIntegrationOauthMutation({}));
    completeIntegrationErr = $derived(this.completeIntegrationOAuthMut.error);

    async onCallbackIntegrationSet(intg?: IntegrationTuple) {
        if (!intg || this.completeIntegrationOAuthMut.isPending) return;

        this.currentlyCompleting = intg;

        const {state, code} = $state.snapshot(this.callbackParams);

        console.log("do complete", intg, {state, code});

        this.callbackParams.reset();
        if (!state || !code) return;

        try {
            await this.doCompleteIntegrationOAuth({type: intg.type, name: intg.name, state, code});
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