<script lang="ts">
	import { page } from "$app/state";
	import * as Sidebar from "$components/ui/sidebar";
	import {
		mdiFire,
		mdiHome,
		mdiPhone,
		mdiTimelineText,
	} from "@mdi/js";
	import { getActiveSidebarItem, type SidebarGroup } from "./sidebar";
	import NavRouteMenuItem from "./NavRouteMenuItem.svelte";

	const sidebarGroups: SidebarGroup[] = [
		{
			items: [
				{ label: "Home", icon: mdiHome, route: "/" }
			]
		},
		{
			label: "System",
			items: [
				{ label: "Incidents", icon: mdiFire, route: "/incidents" },
			],
		},
	];
	
	const active = $derived(getActiveSidebarItem(sidebarGroups, page.route.id));
	const sidebar = Sidebar.useSidebar();
</script>

{#each sidebarGroups as group, i}
	<Sidebar.Group>
		{#if group.label && sidebar.open}
			<Sidebar.GroupLabel>{group.label}</Sidebar.GroupLabel>
		{/if}
		<Sidebar.Menu>
			{#each group.items as item, i (item.label)}
				<NavRouteMenuItem {item} {active} />
			{/each}
		</Sidebar.Menu>
	</Sidebar.Group>
{/each}