import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import {
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
    type ErrorModel,
    listConfiguredIntegrationsOptions,
} from "$lib/api";
import { goto } from "$app/navigation";
import { page } from "$app/state";

export const oauthCallbackParamsSchema = z.object({
    name: z.string().default(""),
    code: z.string().default(""),
    state: z.string().default(""),
});

export class IntegrationOAuthController {
    private callbackParams = useSearchParams(oauthCallbackParamsSchema);
    private callbackName = $derived(this.callbackParams.name);

    private inFlow = $state(false);

    constructor() {
        this.inFlow = false;
        watch(() => this.callbackName, name => {
            if (name) this.onCallback(name);
        });
    }

    private queryClient = useQueryClient();
    private startFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
    private completeFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: () => {
            this.queryClient.invalidateQueries(listConfiguredIntegrationsOptions());
        }
    }));

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
        if (this.completeFlowMut.isPending) return;
        console.log("callback", name);

        const { state, code } = this.callbackParams;
        if (!state || !code) return;

        try {
            const resp = await this.completeFlowMut.mutateAsync({ path: { name }, body: { attributes: { state, code } } });
            this.callbackParams.reset();
            await goto(page.url.pathname);
        } catch (e) {
            this.setFlowError(e);
        }
    }
};

const ctx = new Context<IntegrationOAuthController>("IntegrationOAuthController");
export const initIntegrationOAuthController = () => ctx.set(new IntegrationOAuthController());
export const useIntegrationOAuthController = () => ctx.get();