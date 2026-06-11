import { page } from "$app/state";
import { 
	listInstalledIntegrationsOptions,
	startIntegrationOauthFlowMutation, 
	completeIntegrationOauthFlowMutation,
	type ErrorModel,
	listIntegrationInstallTargetsOptions,
} from "$lib/api";
import { clearQueryParams } from "$src/lib/utils";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { tick } from "svelte";

export class IntegrationOAuthController {
    inFlowForName = $state<string>();
    error = $state<ErrorModel>();

    constructor() {
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

    private queryClient = useQueryClient();
    private async reloadInstalledIntegrations() {
        await tick();
        this.queryClient.invalidateQueries(listInstalledIntegrationsOptions());
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
        onSuccess: () => {
            this.reloadInstalledIntegrations();
        },
    }));

	private async checkOAuthCallback(params: URLSearchParams) {
		if (this.completeOAuthFlowMut.isPending) return;

		const name = params.get("name");
		const code = params.get("code");
		const state = params.get("state");

		if (!name || !state || !code) return;

        this.inFlowForName = name;

		await clearQueryParams();

		try {
			const attributes = { state, code };
			const resp = await this.completeOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			if (resp.data.targetSelectionRequired) {
        		this.queryClient.invalidateQueries(listIntegrationInstallTargetsOptions());
			}
			this.error = undefined;
		} catch (e) {
			this.setError(e);
		} finally {
            this.inFlowForName = undefined;
        }
	}

    inFlow = $derived(this.startOAuthFlowMut.isPending || this.completeOAuthFlowMut.isPending);
}

const ctx = new Context<IntegrationOAuthController>("IntegrationOAuthController");
export const initIntegrationOAuthController = () => ctx.set(new IntegrationOAuthController());
export const useIntegrationOAuthController = () => ctx.get();