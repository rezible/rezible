import { page } from "$app/state";
import { useAppShell } from "$lib/app-shell.svelte";
import { useUserSessionState } from "$lib/user-session.svelte";
import { initIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { Context } from "runed";
import { onDestroy } from "svelte";

import RiPlugLine from "remixicon-svelte/icons/plug-line";
import RiSettings3Line from "remixicon-svelte/icons/settings-3-line";

export class SettingsViewController {
    shell = useAppShell();
    session = useUserSessionState();
    integrations = initIntegrationsController();

    showInitialSetup = $derived(!this.session.isSetup);
    provider = $derived(page.params.provider);

    constructor() {
        this.shell.setChildSidebar({
			search: { placeholder: "Search settings" },
			groups: [
				{
					items: [
						{ label: "General", href: "/settings", icon: RiSettings3Line },
						{ label: "Integrations", href: "/settings/integrations", icon: RiPlugLine },
					],
				},
			],
		});

        onDestroy(() => {
            this.shell.clearChildSidebar();
        });
    }
}

const ctx = new Context<SettingsViewController>("SettingsViewController");
export const initSettingsViewController = () => ctx.set(new SettingsViewController());
export const useSettingsViewController = () => ctx.get();
