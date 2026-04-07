import { useAuthSessionState } from "$src/lib/auth.svelte";
import { Context } from "runed";

export class GeneralSettingsController {
    session = useAuthSessionState();

    constructor() {

    }
}

const ctx = new Context<GeneralSettingsController>("GeneralSettingsController");
export const initGeneralSettingsController = () => ctx.set(new GeneralSettingsController());
export const useGeneralSettingsController = () => ctx.get();