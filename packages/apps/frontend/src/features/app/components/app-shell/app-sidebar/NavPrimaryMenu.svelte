<script lang="ts">
	import * as Sidebar from "$components/ui/sidebar";
    import RiHome from 'remixicon-svelte/icons/home-2-fill';
    import RiFire from 'remixicon-svelte/icons/fire-fill';
	import { type SidebarGroup } from "./sidebar";
	import NavMenuItem from "./NavMenuItem.svelte";

	const sidebarGroups: SidebarGroup[] = [
		{
			items: [
				{ label: "Home", icon: RiHome, route: "/" }
			]
		},
		{
			label: "System",
			items: [
				{ label: "Incidents", icon: RiFire, route: "/incidents" },
				// { label: "Oncall", icon: mdiFire, route: "/oncall" },
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