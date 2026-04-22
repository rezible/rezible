import { Context, watch, type Getter } from "runed";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import {
    startIntegrationOauthFlowMutation, 
    completeIntegrationOauthFlowMutation,
    type ErrorModel,
    listConfiguredIntegrationsOptions,
    type AvailableIntegration,
} from "$lib/api";
import { goto } from "$app/navigation";
import { page } from "$app/state";

import SlackConfig from './config-components/SlackConfig.svelte';
import PlaceholderConfig from './config-components/PlaceholderConfig.svelte';
import GoogleConfig from './config-components/GoogleConfig.svelte';
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import type { Component } from "svelte";
import { resolve } from "$app/paths";
import { clearQueryParams } from "$src/lib/utils";

const configs: Record<string, Component> = {
    slack: SlackConfig,
    google: GoogleConfig,
};

export class IntegrationConfigController {
	integrations = useIntegrationsController();

    private inOAuthFlow = $state(false);

    constructor(integrationFn: Getter<AvailableIntegration>) {
        this.inOAuthFlow = false;
        watch(integrationFn, intg => {this.setIntegration(intg)});
    }

    private integration = $state<AvailableIntegration>();
    name = $derived(this.integration?.name);
    Component = $derived((this.name && this.name in configs) ? configs[this.name] : PlaceholderConfig);

    private setIntegration(intg: AvailableIntegration) {
        this.integration = intg;
        if (intg.oauthRequired) this.checkOAuthCallback(intg.name);
    }

	configured = $derived(!!this.name ? this.integrations.configuredMap.get(this.name) : undefined);

	dataKinds = $derived<Record<string, boolean>>(this.configured?.attributes.dataKinds ?? {});
	enabledDataKinds = $derived(Object.entries(this.dataKinds).filter(([_, enabled]) => (!!enabled)).map(([name, _]) => name) ?? []);

	hasChanges = $state(false);
	setConfig(key: string, val: any) {

	};

    clearConfig(key?: string) {
        
    }

    configError = $state<ErrorModel>();
    private setConfigError(err?: unknown) {
        this.inOAuthFlow = false;
        if (!err) {
            this.configError = undefined;
            return;
        }
        this.configError = {
            title: "Integration Setup Failed",
            detail: (err instanceof Error) ? err.message : "An unknown issue occurred",
        };
    }

    private startOAuthFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
    private completeOAuthFlowMut = createMutation(() => ({
        ...completeIntegrationOauthFlowMutation({}),
        onSuccess: () => {
            // this.queryClient.invalidateQueries(listConfiguredIntegrationsOptions());
        }
    }));

    loading = $derived(this.inOAuthFlow || this.startOAuthFlowMut.isPending || this.completeOAuthFlowMut.isPending);

    async startOAuthFlow() {
        if (this.loading || !this.name) return;
        const name = this.name;
        try {
            const callbackPath = resolve(`/settings/integration-callback/${name}`);
            const resp = await this.startOAuthFlowMut.mutateAsync({ path: { name }, body: { attributes: { callbackPath } } });
            this.inOAuthFlow = true;
            window.location.assign(new URL(resp.data.flow_url));
        } catch (e) {
            this.setConfigError(e);
        }
    }

    private async checkOAuthCallback(name: string) {
        if (this.completeOAuthFlowMut.isPending) return;

        const params = page.url.searchParams;
        const callbackName = params.get("name");
        if (!callbackName || callbackName !== this.name) return;

        const code = params.get("code");
        const state = params.get("state");
        
        if (!state || !code) return;

        try {
            const resp = await this.completeOAuthFlowMut.mutateAsync({ path: { name }, body: { attributes: { state, code } } });
            console.log(resp);
        } catch (e) {
            this.setConfigError(e);
        }
        await clearQueryParams();
    }
};

const ctx = new Context<IntegrationConfigController>("IntegrationConfigController");
export const initIntegrationConfigController = (fn: Getter<AvailableIntegration>) => ctx.set(new IntegrationConfigController(fn));
export const useIntegrationConfigController = () => ctx.get();