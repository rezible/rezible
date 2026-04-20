import { finishOrganizationSetupMutation, type ConfiguredIntegration } from "$lib/api";
import { useAuthSessionState } from "$lib/auth.svelte";
import { Context, watch } from "runed";
import { createMutation } from "@tanstack/svelte-query";
import { initIntegrationsController } from "$src/features/settings/lib/integrationsController.svelte";

type SetupStep = "org_name" | "required_integrations";
const RequiredDataKinds = new Set(["chat", "users"]);

const getEnabledDataKinds = (intg: ConfiguredIntegration) => 
    Object.entries(intg.attributes.dataKinds).
        filter(([_, enabled]) => (enabled)).
        map(([name, _]) => (name));

export class InitialSetupViewController {
    session = useAuthSessionState();

    step = $state<SetupStep>("required_integrations");

    integrations = initIntegrationsController();

	configuredDataKinds = $derived(new Set(this.integrations.configured.flatMap(getEnabledDataKinds)));
    remainingRequiredDataKinds = $derived(RequiredDataKinds.difference(this.configuredDataKinds).values().toArray());
    nextRequiredDataKind = $derived(this.remainingRequiredDataKinds.at(0));
    availableDataKindIntegrations = $derived.by(() => {
        const reqKind = this.nextRequiredDataKind;
        if (!reqKind) return [];
        return this.integrations.available.filter(intg => intg.dataKinds.includes(reqKind));
    });

    canFinish = $derived.by(() => {
        if (!this.integrations) return false;
        if (this.remainingRequiredDataKinds.length === 0) return true;
        return false;
    });

    constructor() {
        watch(() => this.canFinish, ok => {
            if (ok) this.doFinishOrganizationSetup();
        })
    }

    loading = $derived(this.integrations.loading || this.integrations.isConfiguring);

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