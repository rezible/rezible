import { watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { createMutation } from "@tanstack/svelte-query";
import {
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
    type ErrorModel,
} from "$lib/api";

export const oauthCallbackParamsSchema = z.object({
    name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

export class IntegrationOAuthController {
    private callbackParams = useSearchParams(oauthCallbackParamsSchema);
    private callbackName = $derived(this.callbackParams.name);

    private inFlow = $state(false);
    private onCompleted: () => void;

    constructor(onCompletedFn: () => void) {
        this.inFlow = false;
        watch(() => this.callbackName, name => {
            if (name) this.onCallback(name);
        });
        this.onCompleted = onCompletedFn;
    }

    private startFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
    private completeFlowMut = createMutation(() => completeIntegrationOauthFlowMutation({}));

    loading = $derived(this.inFlow || this.startFlowMut.isPending || this.completeFlowMut.isPending);
    error = $state<ErrorModel>();

    private setFlowError(err?: unknown) {
        this.inFlow = false;
        if (err === undefined) {
            this.error = err;
        } else {
            this.error = {
                title: "Integration Setup Failed",
                detail: (err instanceof Error) ? err.message : "An unknown issue occurred",
            };
        }
    }

    async startFlow(name: string) {
        if (this.startFlowMut.isPending) return;
        try {
			const callbackPath = `/settings/integration-callback/${name}`;
			const resp = await this.startFlowMut.mutateAsync({ path: { name }, body: { attributes: { callbackPath } } });
            this.inFlow = true;
            window.location.assign(new URL(resp.data.flow_url));
        } catch (e) {
            this.setFlowError(e);
        }
    }

    private async onCallback(name: string) {
        console.log("callback", name);
        if (this.completeFlowMut.isPending) return;

        const { state, code } = this.callbackParams;
        this.callbackParams.reset();
        console.log("params", {state, code});

        if (!state || !code) return;

        try {
            const resp = await this.completeFlowMut.mutateAsync({ path: { name }, body: { attributes: { state, code } } });
            this.onCompleted();
            console.log("completed", resp);
        } catch (e) {
            this.setFlowError(e);
        }
    }
};