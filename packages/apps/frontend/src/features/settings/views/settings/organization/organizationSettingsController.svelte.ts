import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { useUserSessionState } from "$lib/user-session.svelte";
import { Context, watch } from "runed";

export class OrganizationSettingsViewController {
	session = useUserSessionState();

	orgName = $derived(this.session.org?.attributes.name ?? "");
	isOrgAdmin = $derived(this.session.isAdmin);

	// constructor() {
	// 	watch(
	// 		() => this.session.ready && !this.session.isAdmin,
	// 		(shouldRedirect) => {
	// 			if (shouldRedirect) goto(resolve("/settings/user"));
	// 		}
	// 	);
	// }
}

const ctx = new Context<OrganizationSettingsViewController>("OrganizationSettingsViewController");

export const initOrganizationSettingsViewController = () => ctx.set(new OrganizationSettingsViewController());
