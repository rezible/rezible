import { page } from "$app/state";
import { useAuthSessionState } from "$src/lib/auth.svelte";
import { convertSettingsViewParam } from "$src/params/settingsView";
import { Context } from "runed";

export class SettingsViewController {
    viewParam = $derived(convertSettingsViewParam(page.params.view));
    session = useAuthSessionState();
    showInitialSetup = $derived(!this.session.isSetup);

    constructor() {

    }
}

const ctx = new Context<SettingsViewController>("SettingsViewController");
export const initSettingsViewController = () => ctx.set(new SettingsViewController());
export const useSettingsViewController = () => ctx.get();