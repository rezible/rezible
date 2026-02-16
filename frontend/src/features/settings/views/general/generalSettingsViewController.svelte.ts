import { useAuthSessionState } from "$src/lib/auth.svelte";
import { Context } from "runed";

export class GeneralSettingsViewController {
    session = useAuthSessionState();

    constructor() {

    }
}

const ctx = new Context<GeneralSettingsViewController>("GeneralSettingsViewController");
export const initGeneralSettingsViewController = () => ctx.set(new GeneralSettingsViewController());
export const useGeneralSettingsViewController = () => ctx.get();