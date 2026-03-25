import { finishOrganizationSetupMutation } from "$lib/api";
import { useAuthSessionState } from "$lib/auth.svelte";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { InitialIntegrationsSetupController } from "./initialIntegrationsSetupController.svelte";

export class InitialSetupViewController {
    session = useAuthSessionState();

    integrations: InitialIntegrationsSetupController;

    canFinish = $derived.by(() => {
        if (!this.integrations) return false;
        if (this.integrations.remainingRequiredDataKinds.length === 0) return true;
        return false;
    });

    constructor() {
        this.integrations = new InitialIntegrationsSetupController();
        watch(() => this.canFinish, ok => {
            if (ok) this.doFinishOrganizationSetup();
        })
    }

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