import { completeIntegrationOauthMutation, createIntegrationMutation, finishOrganizationSetupMutation, listConfiguredIntegrationsOptions, listSupportedIntegrationsOptions, startIntegrationOauthMutation, type CompleteIntegrationOAuthRequestAttributes, type Integration } from "$src/lib/api";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
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
        watch(() => this.callbackName, name => {this.onCallbackSet(name)});
        this.onCompleted = onCompleted;
    }

    private startIntegrationOAuthMut = createMutation(() => startIntegrationOauthMutation({}));
    loadingFlowUrl = $derived(this.startIntegrationOAuthMut.isPending);
    startFlowUrl = $derived(this.startIntegrationOAuthMut.data?.data.flow_url);
    startFlowErr = $derived(this.startIntegrationOAuthMut.error);

    async startFlow(name: string) {
        try {
            const resp = await this.startIntegrationOAuthMut.mutateAsync({body: {attributes: {name}}});
            const flowUrl = new URL(resp.data.flow_url);
            alert(`oauth navigation to ${resp.data.flow_url}`);
            window.location.assign(flowUrl);
        } catch (e) {
            console.error("failed to complete", e);
        }
    }

    private completeIntegrationOAuthMut = createMutation(() => completeIntegrationOauthMutation({}));
    completingFlow = $derived(this.completeIntegrationOAuthMut.isPending);
    completeIntegrationErr = $derived(this.completeIntegrationOAuthMut.error);

    private async onCallbackSet(name?: string) {
        if (!name || this.completingFlow) return;

        const {state, code} = $state.snapshot(this.callbackParams);

        this.callbackParams.reset();
        if (!state || !code) return;

        try {
            await this.doCompleteIntegrationOAuth({name, state, code});
        } catch (e) {
            console.error("failed to complete", e);
        }
    }

    private async doCompleteIntegrationOAuth(attributes: CompleteIntegrationOAuthRequestAttributes) {
        const resp = await this.completeIntegrationOAuthMut.mutateAsync({body: {attributes}});
        console.log("completed", resp);
        this.onCompleted();
    }
};

export class SetupViewState {
    session = useAuthSessionState();

    oauth: OAuthIntegrationSetupState;

    private supportedIntegrationsQuery = createQuery(() => listSupportedIntegrationsOptions());
    private loadingSupported = $derived(this.supportedIntegrationsQuery.isFetching);
    supportedIntegrations = $derived(this.supportedIntegrationsQuery.data?.data || []);

    private configuredIntegrationsQuery = createQuery(() => listConfiguredIntegrationsOptions());
    private loadingConfigured = $derived(this.configuredIntegrationsQuery.isFetching);
    configuredIntegrations = $derived(this.configuredIntegrationsQuery.data?.data || []);

    loadingIntegrations = $derived(this.loadingSupported || this.loadingConfigured);

    constructor() {
        this.oauth = new OAuthIntegrationSetupState(() => {this.configuredIntegrationsQuery.refetch()});
    }

    private configureIntegrationMut = createMutation(() => ({
        ...createIntegrationMutation({}),
        onSuccess: () => {
            this.configuredIntegrationsQuery.refetch();
        }
    }));
    configuringIntegration = $derived(this.configureIntegrationMut.isPending);
    configureIntegrationMutErr = $derived(this.configureIntegrationMut.error);
    
    async doConfigureIntegration(name: string, config: any) {
        this.configureIntegrationMut.mutateAsync({
            body: {
                attributes: {name, config}
            }
        })
    }

	private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    finishingSetup = $derived(this.finishOrgSetupMut.isPending);
    async doFinishOrganizationSetup() {
        const id = this.session.org?.id;
        if (!id) return;
		await this.finishOrgSetupMut.mutateAsync({path: {id}});
		this.session.refetch();
    }

    loading = $derived(this.finishingSetup || this.configuringIntegration || this.loadingIntegrations)
}

const ctx = new Context<SetupViewState>("setupView");
export const setSetupViewState = () => ctx.set(new SetupViewState());
export const useSetupViewState = () => ctx.get();