import { afterNavigate, beforeNavigate, goto } from "$app/navigation";
import { page } from "$app/state";
import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
import { useAuthSessionState } from "$src/lib/auth-session.svelte";
import { convertSettingsViewParam, type SettingsViewParam } from "$src/params/settingsView";
import { Context } from "runed";

const viewParamLabel = (p: SettingsViewParam): string => {
    switch (p) {
        case "incidents": return "Incidents"
        case "integrations": return "Integrations"
        default: return "General"
    }
}

export class SettingsViewController {
    session = useAuthSessionState();
    showInitialSetup = $derived(!this.session.isSetup);
    viewParam = $derived(convertSettingsViewParam(page.params.view));

    constructor() {
        this.preventInitialSetupNavigation();
        setPageBreadcrumbs(() => ([
            { label: "Settings", href: "/settings" },
            { label: viewParamLabel(this.viewParam), href: `/settings/${this.viewParam ?? ""}` }
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