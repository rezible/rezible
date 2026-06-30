import { page } from "$app/state";
import { useAppShell, type AppSidebarModel, type AppSidebarGroup } from "$lib/app-shell.svelte";
import { useUserSessionState } from "$lib/user-session.svelte";
import { initIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { Context, watch } from "runed";
import { onDestroy } from "svelte";

import RiPlugLine from "remixicon-svelte/icons/plug-line";
import RiBuilding2Line from "remixicon-svelte/icons/building-2-line";
import RiUserSettingsLine from "remixicon-svelte/icons/user-settings-line";

const sidebarAdminGroups: AppSidebarGroup[] = [{
    label: "Administration",
    items: [
        { label: "Organization", href: "/settings/organization", icon: RiBuilding2Line },
    ],
}];

const makeSettingsSidebar = (isAdmin: boolean): AppSidebarModel => ({
    search: { placeholder: "Search settings" },
    groups: [
        {
            label: "User",
            items: [

                { label: "Preferences", href: "/settings/user/preferences", icon: RiUserSettingsLine },
            ]
        },
        {
            label: "App",
            items: [
                { label: "Integrations", href: "/settings/integrations", icon: RiPlugLine },
            ],
        },
        ...(isAdmin ? sidebarAdminGroups : []),
    ],
});

export class SettingsViewController {
    shell = useAppShell();
    session = useUserSessionState();

    integrations = initIntegrationsController();

    showInitialSetup = $derived(!this.session.isSetup);
    provider = $derived(page.params.provider);

    sidebar = $derived(makeSettingsSidebar(this.session.isAuthenticated))

    constructor() {
        watch(() => this.sidebar, sb => {
            this.shell.setChildSidebar(sb);
        });
        onDestroy(() => {
            this.shell.clearChildSidebar();
        });
    }
}

const ctx = new Context<SettingsViewController>("SettingsViewController");
export const initSettingsViewController = () => ctx.set(new SettingsViewController());
export const useSettingsViewController = () => ctx.get();
