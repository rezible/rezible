import { finishOrganizationSetupMutation, type ConfiguredIntegration } from "$lib/api";
import { useAuthSessionState } from "$lib/auth.svelte";
import { Context, watch } from "runed";
import { createQuery, createMutation } from "@tanstack/svelte-query";
import {
    listAvailableIntegrationsOptions,
    listConfiguredIntegrationsOptions,
    configureIntegrationMutation,
    type ConfigureIntegrationRequestBody,
} from "$lib/api";
import { IntegrationOAuthController } from "$src/features/settings/lib/integrationOAuthController.svelte";

type SetupStep = "org_name" | "required_integrations";
const RequiredDataKinds = new Set(["chat", "users"]);

const getEnabledDataKinds = (intg: ConfiguredIntegration) => 
    Object.entries(intg.attributes.dataKinds).
        filter(([_, enabled]) => (enabled)).
        map(([name, _]) => (name));

export class RequiredIntegrationsSetupController {
    oauth = new IntegrationOAuthController(() => { 
        this.listConfiguredQuery.refetch();
    });

    private listAvailableQuery = createQuery(() => listAvailableIntegrationsOptions());
    available = $derived(this.listAvailableQuery.data?.data || []);
    availableMap = $derived(new Map(this.available.map(intg => ([intg.name, intg]))));

    private listConfiguredQuery = createQuery(() => listConfiguredIntegrationsOptions());
    configured = $derived(this.listConfiguredQuery.data?.data || []);
	configuredMap = $derived(new Map(this.configured.map(intg => ([intg.name, intg]))));
	configuredDataKinds = $derived(new Set(this.configured.flatMap(getEnabledDataKinds)));

    remainingRequiredDataKinds = $derived(RequiredDataKinds.difference(this.configuredDataKinds).values().toArray());
    nextRequiredDataKind = $derived(this.remainingRequiredDataKinds.at(0));
    availableDataKindIntegrations = $derived.by(() => {
        const reqKind = this.nextRequiredDataKind;
        if (!reqKind) return [];
        return this.available.filter(intg => intg.dataKinds.includes(reqKind));
    });

    private configureMut = createMutation(() => ({
        ...configureIntegrationMutation({}),
        onSuccess: () => { this.listConfiguredQuery.refetch() }
    }));
    configureMutErr = $derived(this.configureMut.error);

    async doConfigure(name: string, attributes: ConfigureIntegrationRequestBody["attributes"]) {
        this.configureMut.mutateAsync({
            path: { name },
            body: { attributes }
        })
    }

    isLoading = $derived(this.listAvailableQuery.isPending || this.listConfiguredQuery.isPending);
    isConfiguring = $derived(this.configureMut.isPending);
};

export class InitialSetupViewController {
    session = useAuthSessionState();

    step = $state<SetupStep>("required_integrations");

    integrations = new RequiredIntegrationsSetupController();

    canFinish = $derived.by(() => {
        if (!this.integrations) return false;
        if (this.integrations.remainingRequiredDataKinds.length === 0) return true;
        return false;
    });

    constructor() {
        watch(() => this.canFinish, ok => {
            if (ok) this.doFinishOrganizationSetup();
        })
    }

    integrationsLoading = $derived(this.integrations?.isLoading || this.integrations.isConfiguring);
    loading = $derived(this.integrationsLoading);

    private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    async doFinishOrganizationSetup() {
        if (this.finishOrgSetupMut.isPending) return;
        const id = this.session.org?.id;
        if (!id) return;
        await this.finishOrgSetupMut.mutateAsync({ path: { id } });
        this.session.refetch();
    }
    finishingOrgSetup = $derived(this.finishOrgSetupMut.isPending);
}

const ctx = new Context<InitialSetupViewController>("InitialSetupViewController");
export const initInitialSetupViewController = () => ctx.set(new InitialSetupViewController());
export const useInitialSetupViewController = () => ctx.get();