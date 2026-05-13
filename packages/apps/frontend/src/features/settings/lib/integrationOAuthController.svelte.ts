import { page } from "$app/state";
import { startIntegrationOauthFlowMutation, completeIntegrationOauthFlowMutation, selectIntegrationOauthFlowMutation, type ErrorModel, type ExternalIntegrationOption, listConfiguredIntegrationsOptions } from "$lib/api";
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

    name = $state<string>();

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
    private async reloadConfigured() {
        await tick();
        this.queryClient.invalidateQueries(listConfiguredIntegrationsOptions());
    }

    private startOAuthFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
    private completeOAuthFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: () => {
            this.reloadConfigured();
        },
    }));
    private selectOAuthFlowMut = createMutation(() => ({
        ...selectIntegrationOauthFlowMutation({}),
        onSuccess: () => {
            this.reloadConfigured();
        },
    }));

    pending = $derived(this.startOAuthFlowMut.isPending ||
            this.completeOAuthFlowMut.isPending ||
            this.selectOAuthFlowMut.isPending);

    async getStartFlowUrl(name: string, callbackPath: string) {
        const resp = await this.startOAuthFlowMut.mutateAsync({
            path: { name },
            body: { attributes: { callbackPath } },
        });
        return new URL(resp.data.flow_url);
    }

	selectionToken = $state<string>();
	selectionOptions = $state<ExternalIntegrationOption[]>([]);
	selectedExternalRefs = new SvelteSet<string>();
	selectionRequired = $derived(!!this.selectionToken && this.selectionOptions.length > 0);

	private async checkOAuthCallback(params: URLSearchParams) {
		if (this.completeOAuthFlowMut.isPending) return;

		const name = params.get("name");
		const code = params.get("code");
		const state = params.get("state");

		if (!name || !state || !code) return;
        this.name = name;

		await clearQueryParams();

		try {
			const attributes = { state, code };
			const resp = await this.completeOAuthFlowMut.mutateAsync({
				path: { name },
				body: { attributes },
			});
			if (resp.data.status === "selection_required") {
				if (!resp.data.selectionToken) {
					console.error("no selection token returned for oauth response");
				} else {
					this.setSelection(resp.data.selectionToken, resp.data.options);
				}
			} else {
                this.name = undefined;
				this.setSelection();
			}
			this.setError();
		} catch (e) {
			this.setError(e);
		}
	}

	toggleSelection(ref: string, selected: boolean) {
		if (selected) {
			this.selectedExternalRefs.add(ref);
		} else {
			this.selectedExternalRefs.delete(ref);
		}
	}

	private setSelection(token?: string, options: ExternalIntegrationOption[] = []) {
		this.selectionToken = token;
		this.selectionOptions = options;
		this.selectedExternalRefs.clear();
		for (const option of options) {
			this.selectedExternalRefs.add(option.externalRef);
		}
	}

	async selectOAuthOptions() {
		if (!this.name || !this.selectionToken || this.selectedExternalRefs.size === 0) return;
		try {
			const attributes = {
				selectionToken: this.selectionToken,
				externalRefs: [...this.selectedExternalRefs],
			}
			await this.selectOAuthFlowMut.mutateAsync({
				path: { name: this.name },
				body: { attributes },
			});
            this.name = undefined;
			this.setSelection();
			this.setError();
		} catch (e) {
			this.setError(e);
		}
	}
}

const ctx = new Context<IntegrationOAuthController>("IntegrationOAuthController");
export const initIntegrationOAuthController = () => ctx.set(new IntegrationOAuthController());
export const useIntegrationOAuthController = () => ctx.get();