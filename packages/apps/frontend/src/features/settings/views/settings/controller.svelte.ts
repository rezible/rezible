import { page } from "$app/state";
import { useAppShell, type AppSidebarItem, type AppSidebarModel } from "$lib/app-shell.svelte";
import { useUserSessionState } from "$lib/user-session.svelte";
import { initIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { Context, watch } from "runed";
import { onDestroy } from "svelte";

import RiPlugLine from "remixicon-svelte/icons/plug-line";
import RiBuilding2Line from "remixicon-svelte/icons/building-2-line";
import RiUserSettingsLine from "remixicon-svelte/icons/user-settings-line";
import RiNotification3Line from "remixicon-svelte/icons/notification-3-line";
import RiTeamLine from "remixicon-svelte/icons/team-line";
import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";

const orgPaths = ["/settings/organization", "/settings/integrations", "/settings/incidents"];
const isOrgPath = (currPath: string) => orgPaths.some(p => currPath.startsWith(p));

const makeSettingsSidebar = (isAdmin: boolean, hasIncidentManagement: boolean): AppSidebarModel => {
	const personalItems: AppSidebarItem[] = [
		{ label: "Preferences", href: "/settings/user", icon: RiUserSettingsLine },
		{ label: "Notifications", href: "/settings/user/notifications", icon: RiNotification3Line },
	];

	let groups: AppSidebarModel["groups"] = [
		{ label: "Personal", items: personalItems },
	]

	if (isAdmin) {
		const organizationItems: AppSidebarItem[] = [
			{ label: "Preferences", href: "/settings/organization", icon: RiBuilding2Line },
			{ label: "Members", href: "/settings/organization/members", icon: RiTeamLine },
			{ label: "Integrations", href: "/settings/integrations", icon: RiPlugLine },
		];
		if (hasIncidentManagement) {
			organizationItems.push({
				label: "Incident management",
				href: "/settings/incidents",
				icon: RiAlarmWarningLine,
			});
		}
		groups.push({ label: "Organization", items: organizationItems });
	}

	return {
		search: { placeholder: "Filter settings" },
		groups,
	};
};

export class SettingsViewController {
	shell = useAppShell();
	session = useUserSessionState();

	integrations = initIntegrationsController();

	showInitialSetup = $derived(!this.session.isSetup);
	provider = $derived(page.params.provider);
	isOrganizationRoute = $derived(isOrgPath(page.url.pathname));
	showAccessDenied = $derived(this.session.ready && this.isOrganizationRoute && !this.session.isAdmin);
	hasIncidentManagement = $derived(this.integrations.installedCapabilities.has("incident_management"));

	sidebar = $derived(makeSettingsSidebar(this.session.isAdmin, this.hasIncidentManagement));

	constructor() {
		watch(
			() => this.sidebar,
			(sb) => {
				this.shell.setChildSidebar(sb);
			}
		);
		onDestroy(() => {
			this.shell.clearChildSidebar();
		});
	}
}

const ctx = new Context<SettingsViewController>("SettingsViewController");
export const initSettingsViewController = () => ctx.set(new SettingsViewController());
export const useSettingsViewController = () => ctx.get();
