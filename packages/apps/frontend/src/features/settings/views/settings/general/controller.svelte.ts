import { useUserSessionState } from "$src/lib/user-session.svelte";
import { Context } from "runed";

export class GeneralSettingsController {
    session = useUserSessionState();

    constructor() {

    }
}

const ctx = new Context<GeneralSettingsController>("GeneralSettingsController");
export const initGeneralSettingsController = () => ctx.set(new GeneralSettingsController());
export const useGeneralSettingsController = () => ctx.get();