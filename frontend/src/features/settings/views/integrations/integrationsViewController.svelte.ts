import { useAuthSessionState } from "$src/lib/auth.svelte";
import { Context } from "runed";

export class IntegrationsViewController {
    session = useAuthSessionState();

    constructor() {

    }
}

const ctx = new Context<IntegrationsViewController>("IntegrationsViewController");
export const initIntegrationsViewController = () => ctx.set(new IntegrationsViewController());
export const useIntegrationsViewController = () => ctx.get();