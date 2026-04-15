import { afterNavigate, beforeNavigate, goto } from "$app/navigation";
import { page } from "$app/state";
import { setPageBreadcrumbs } from "$lib/appShell.svelte";
import { useAuthSessionState } from "$lib/auth.svelte";
import { convertSettingsViewParam } from "$src/params/settingsView";
import { Context } from "runed";
import { initIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

export class SettingsViewController {
    session = useAuthSessionState();
    showInitialSetup = $derived(!this.session.isSetup);
    oauth = initIntegrationOAuthController();

    viewParam = $derived(convertSettingsViewParam(page.params.view));

    constructor() {
        this.preventInitialSetupNavigation();
        setPageBreadcrumbs(() => ([
            { label: "Settings", href: "/settings" },
            { label: this.viewParam }
        ]));
    }

    private preventInitialSetupNavigation() {
        afterNavigate(nav => {
            if (!this.showInitialSetup) return;
            const routeId = nav.to?.route.id;
            if (!!routeId && routeId !== "/settings" && !routeId.startsWith("/settings/integration-callback")) {
                goto("/settings");
            }
        });
        beforeNavigate(nav => {
            if (nav.willUnload || !nav.to) return;
            if (this.showInitialSetup) nav.cancel();
        })
    }
}

const ctx = new Context<SettingsViewController>("SettingsViewController");
export const initSettingsViewController = () => ctx.set(new SettingsViewController());
export const useSettingsViewController = () => ctx.get();