import { page } from "$app/state";
import { 
	listInstalledIntegrationsOptions,
	startIntegrationOauthFlowMutation, 
	completeIntegrationOauthFlowMutation, 
	installIntegrationTargetsMutation,
	type ErrorModel, 
	type IntegrationInstallTarget,
	listIntegrationInstallTargetsOptions,
} from "$lib/api";
import { clearQueryParams } from "$src/lib/utils";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { tick } from "svelte";
import { SvelteSet } from "svelte/reactivity";

export class IntegrationOAuthController {
    constructor() {
        watch(() => page.url.search, search => {
            this.checkOAuthCallback(new URLSearchParams(search));
        })
    }

    installingName = $state<string>();
    error = $state<ErrorModel>();
    private setError(err?: unknown) {
        if (!err) {
            this.error = undefined;
            return;
        }
        this.error = {
            title: "Integration Setup Failed",
            detail: err instanceof Error ? err.message : "An unknown issue occurred",
        };
    }

    private queryClient = useQueryClient();
    private async reloadInstalledIntegrations() {
        await tick();
        this.queryClient.invalidateQueries(listInstalledIntegrationsOptions());
    }

    private startOAuthFlowMut = createMutation(() => ({
		...startIntegrationOauthFlowMutation({}),
	}));

    async getStartFlowUrl(name: string) {
        const resp = await this.startOAuthFlowMut.mutateAsync({path: { name }});
        return new URL(resp.data.flow_url);
    }

    private completeOAuthFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: () => {
            this.reloadInstalledIntegrations();
        },
    }));

    inFlow = $derived(this.startOAuthFlowMut.isPending ||
            this.completeOAuthFlowMut.isPending);

	private async checkOAuthCallback(params: URLSearchParams) {
		if (this.completeOAuthFlowMut.isPending) return;

		const name = params.get("name");
		const code = params.get("code");
		const state = params.get("state");

		if (!name || !state || !code) return;
        this.installingName = name;

		await clearQueryParams();

		try {
			const attributes = { state, code };
			const resp = await this.completeOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			if (resp.data.targetSelectionRequired) {
        		this.queryClient.invalidateQueries(listIntegrationInstallTargetsOptions());
			} else {
                this.installingName = undefined;
			}
			this.setError();
		} catch (e) {
			this.setError(e);
		}
	}
}

const ctx = new Context<IntegrationOAuthController>("IntegrationOAuthController");
export const initIntegrationOAuthController = () => ctx.set(new IntegrationOAuthController());
export const useIntegrationOAuthController = () => ctx.get();