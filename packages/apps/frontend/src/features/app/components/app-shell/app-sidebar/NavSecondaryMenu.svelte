<script lang="ts">
	import { page } from "$app/state";
	import * as Sidebar from "$components/ui/sidebar";
	import { mdiCog, mdiFire, mdiHome, mdiLifebuoy, mdiSend } from "@mdi/js";
	import { getActiveSidebarItem, type SidebarGroup, type SidebarItem } from "./sidebar";
	import type { ComponentProps } from "svelte";
	import Icon from "$src/components/icon/Icon.svelte";

	let { ref = $bindable(null), ...restProps }: ComponentProps<typeof Sidebar.Group> = $props();
	const items: SidebarItem[] = [
		{ label: "Support", icon: mdiLifebuoy, route: "/" },
		{ label: "Feedback", icon: mdiSend, route: "/" },
		{ label: "Settings", icon: mdiCog, route: "/settings" },
	];
</script>

<Sidebar.Group bind:ref {...restProps}>
	<Sidebar.GroupContent>
		<Sidebar.Menu>
            {#each items as item (item.label)}
                <Sidebar.MenuItem>
                    <Sidebar.MenuButton size="sm">
                        {#snippet child({ props })}
                            <a href={item.route} {...props}>
                                {#if item.icon}
                                    <Icon data={item.icon} />
                                {/if}
                                <span>{item.label}</span>
                            </a>
                        {/snippet}
                    </Sidebar.MenuButton>
                </Sidebar.MenuItem>
            {/each}
		</Sidebar.Menu>
	</Sidebar.GroupContent>
</Sidebar.Group>
