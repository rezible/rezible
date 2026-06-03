import { finishOrganizationSetupMutation, type AvailableIntegration, type InstalledIntegration } from "$lib/api";
import { useAuthSessionState } from "$src/lib/auth-session.svelte";
import { Context, watch } from "runed";
import { createMutation } from "@tanstack/svelte-query";
import { useIntegrationsController } from "$src/features/settings/lib/integrationsController.svelte";

type SetupStep = "org_name" | "required_integrations";
const RequiredCapabilities = new Set(["chat", "users"]);

const getEnabledCapabilities = (intg: InstalledIntegration) => 
    Object.entries(intg.attributes.capabilities).
        filter(([_, enabled]) => (enabled)).
        map(([name, _]) => (name));

export class InitialSetupViewController {
    session = useAuthSessionState();
    private integrations = useIntegrationsController();

    step = $state<SetupStep>("required_integrations");

    availableOptions = $derived(this.integrations.available.filter(intg => (!this.integrations.installationsByName.has(intg.name))));
	installedCapabilities = $derived(new Set(this.integrations.installed.flatMap(getEnabledCapabilities)));
    remainingRequiredCapabilities = $derived(RequiredCapabilities.difference(this.installedCapabilities).values().toArray());
    availableIntegrationsForCapabilities = $derived.by(() => {
        const capMap = new Map<string, AvailableIntegration[]>();
        this.availableOptions.forEach(intg => {
            intg.supportedCapabilities.forEach(cap => {
                capMap.set(cap, [...(capMap.get(cap) || []), intg]);
            });
        });
        return capMap;
    });

    canFinish = $derived.by(() => {
        if (!this.integrations) return false;
        if (this.remainingRequiredCapabilities.length === 0) return true;
        return false;
    });

    constructor() {
        $inspect(this.availableOptions);
        watch(() => this.canFinish, ok => {
            if (ok) this.doFinishOrganizationSetup();
        })
    }

    private finishingSetup = $state(false);
    private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
    async doFinishOrganizationSetup() {
        if (this.finishOrgSetupMut.isPending) return;
        const id = this.session.org?.id;
        if (!id) return;
        this.finishingSetup = true;
        try {
            await this.finishOrgSetupMut.mutateAsync({ path: { id } });
            this.session.refetch();
        } catch (e) {
            console.error("failed to finish setup", e);
            this.finishingSetup = false;
        }
    }

    loading = $derived(this.finishingSetup || this.integrations.loading);
}

const ctx = new Context<InitialSetupViewController>("InitialSetupViewController");
export const initInitialSetupViewController = () => ctx.set(new InitialSetupViewController());
export const useInitialSetupViewController = () => ctx.get();