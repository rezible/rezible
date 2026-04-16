<script lang="ts">
	import * as Sidebar from "$components/ui/sidebar";
	import { mdiCog, mdiLifebuoy, mdiSend } from "@mdi/js";
	import { type SidebarItem } from "./sidebar";
	import type { ComponentProps } from "svelte";
	import NavMenuItem from "./NavMenuItem.svelte";

	let { ref = $bindable(null), ...restProps }: ComponentProps<typeof Sidebar.Group> = $props();
	const items: SidebarItem[] = [
		// { label: "Support", icon: mdiLifebuoy, route: "/" },
		// { label: "Feedback", icon: mdiSend, route: "/" },
		{ label: "Settings", icon: mdiCog, route: "/settings" },
	];
</script>

<Sidebar.Group bind:ref {...restProps}>
	<Sidebar.GroupContent>
		<Sidebar.Menu class="gap-1">
            {#each items as item (item.label)}
                <NavMenuItem {item} />
            {/each}
		</Sidebar.Menu>
	</Sidebar.GroupContent>
</Sidebar.Group>
