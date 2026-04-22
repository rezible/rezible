<script lang="ts">
	import * as Sidebar from "$components/ui/sidebar";
	import {
		mdiFire,
		mdiHome,
	} from "@mdi/js";
	import { type SidebarGroup } from "./sidebar";
	import NavMenuItem from "./NavMenuItem.svelte";

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
				{ label: "Oncall", icon: mdiFire, route: "/oncall" },
			],
		},
	];
	const sidebar = Sidebar.useSidebar();
</script>

{#each sidebarGroups as group}
	<Sidebar.Group>
		{#if group.label && sidebar.open}
			<Sidebar.GroupLabel>{group.label}</Sidebar.GroupLabel>
		{/if}
		<Sidebar.Menu>
			{#each group.items as item (item.label)}
				<NavMenuItem {item} />
			{/each}
		</Sidebar.Menu>
	</Sidebar.Group>
{/each}