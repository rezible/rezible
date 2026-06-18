import {
	startIntegrationOauthFlowMutation, 
	completeIntegrationOauthFlowMutation,
	type ErrorModel,
	type IntegrationOAuthInstallResult,
} from "$lib/api";

import { page } from "$app/state";
import { clearQueryParams } from "$lib/utils";
import { createMutation } from "@tanstack/svelte-query";
import { watch } from "runed";
import { tick } from "svelte";

export class IntegrationOAuthController {
    inFlowForName = $state<string>();
    error = $state<ErrorModel>();

    private onSuccess: (res: IntegrationOAuthInstallResult) => void;

    constructor(onSuccess: (res: IntegrationOAuthInstallResult) => void) {
        this.onSuccess = onSuccess;
        watch(() => page.url.search, search => {
            this.checkOAuthCallback(new URLSearchParams(search));
        });
    }

    private setError(err: unknown) {
        this.error = {
            title: "Integration Setup Failed",
            detail: err instanceof Error ? err.message : "An unknown issue occurred",
        };
    };

    clearFlow() {
        this.inFlowForName = undefined;
        this.error = undefined;
    }

    private startOAuthFlowMut = createMutation(() => ({
        ...startIntegrationOauthFlowMutation({}),
    }));

    async startFlowFor(name: string) {
        this.inFlowForName = name;
        const resp = await this.startOAuthFlowMut.mutateAsync({path: { name }});
        window.location.assign(new URL(resp.data.flow_url));
    }

    private completeOAuthFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: async ({data}) => {
            await tick();
            this.onSuccess?.(data);
        },
    }));

    private async checkOAuthCallback(params: URLSearchParams) {
        const name = params.get("name");
        const code = params.get("code");
        const state = params.get("state");

        if (this.completeOAuthFlowMut.isPending) return;
        if (!name || !state || !code) return;

        this.inFlowForName = name;

        await clearQueryParams();

        try {
            const attributes = { state, code };
            await this.completeOAuthFlowMut.mutateAsync({
                path: { name },
                body: { attributes },
            });
            this.error = undefined;
        } catch (e) {
            this.setError(e);
        } finally {
            this.inFlowForName = undefined;
        }
    }

    inFlow = $derived(this.startOAuthFlowMut.isPending || this.completeOAuthFlowMut.isPending);
};