<script lang="ts">
	import { appShell } from "$features/app";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import OrganizationTeamsTab from "./tabs/OrganizationTeamsTab.svelte";
	import OrganizationUsersTab from "./tabs/OrganizationUsersTab.svelte";
	import { initOrganizationSettingsViewController } from "./organizationSettingsViewController.svelte";
	import type { OrganizationSettingsViewParam } from "$src/params/organizationSettingsView";

	const view = initOrganizationSettingsViewController();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Organization", href: "/settings/organization" },
	]);

	const tabs: Tab<OrganizationSettingsViewParam>[] = [
		{ label: "Teams", view: undefined, component: OrganizationTeamsTab },
		{ label: "Users", view: "users", component: OrganizationUsersTab },
	];
</script>

<div class="mb-2 rounded border p-3">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<div>
			<h2 class="text-lg font-semibold">{view.orgName || "Organization"}</h2>
			<span class="text-sm text-surface-content/80">
				{#if view.isOrgAdmin}
					You are an organization admin.
				{:else}
					Read-only access. Organization admin is required for team and membership changes.
				{/if}
			</span>
		</div>
	</div>
</div>

<TabbedViewContainer {tabs} path="/settings/organization" />
