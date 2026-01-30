import { watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { createMutation } from "@tanstack/svelte-query";
import {
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
} from "$lib/api";

export const oauthCallbackParamsSchema = z.object({
    name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

export class IntegrationOAuthSetupDriver {
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