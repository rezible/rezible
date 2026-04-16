<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import * as Sidebar from "$components/ui/sidebar";
	import Icon from "$src/components/icon/Icon.svelte";
	import ChevronRightIcon from "@lucide/svelte/icons/chevron-right";
	import { findActiveSidebarSubItem, type ActiveSidebarItem, type SidebarItem } from "./sidebar";

    type Props = {
        item: SidebarItem;
        active: ActiveSidebarItem;
    };
    const { item, active }: Props = $props();
    const itemRouteActive = $derived(active.itemRoute === item.route);
    const activeSubItemIdx = $derived(findActiveSidebarSubItem(item.subItems, active.subItemRoute));
    const isActive = $derived((itemRouteActive && (activeSubItemIdx < 0)) ? true : undefined);
</script>

{#if !item.subItems}
    <Sidebar.MenuItem>
        <Sidebar.MenuButton {isActive} tooltipContent={item.label}>
            {#snippet child({ props })}
                <a href={item.route} {...props}>
                    {#if item.icon}
                        <Icon data={item.icon} />
                    {/if}
                    {item.label}
                </a>
            {/snippet}
        </Sidebar.MenuButton>
    </Sidebar.MenuItem>
{:else}
    <Collapsible.Root open={itemRouteActive} class="group/collapsible">
        {#snippet child({ props })}
            <Sidebar.MenuItem {...props}>
                <Collapsible.Trigger>
                    {#snippet child({ props })}
                        <Sidebar.MenuButton {...props} {isActive} tooltipContent={item.label}>
                            {#snippet child({ props })}
                                <a href={item.route} {...props}>
                                    {#if item.icon}
                                        <Icon data={item.icon} />
                                    {/if}
                                    {item.label}
                                    <ChevronRightIcon class="ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                </a>
                            {/snippet}
                        </Sidebar.MenuButton>
                    {/snippet}
                </Collapsible.Trigger>
                <Collapsible.Content>
                    <Sidebar.MenuSub>
                        {#each item.subItems as subItem, si (subItem.label)}
                            <Sidebar.MenuSubItem>
                                <Sidebar.MenuSubButton isActive={(activeSubItemIdx === si) ? true : undefined}>
                                    {#snippet child({ props })}
                                        <a href={subItem.route} {...props}>
                                            <span>{subItem.label}</span>
                                        </a>
                                    {/snippet}
                                </Sidebar.MenuSubButton>
                            </Sidebar.MenuSubItem>
                        {/each}
                    </Sidebar.MenuSub>
                </Collapsible.Content>
            </Sidebar.MenuItem>
        {/snippet}
    </Collapsible.Root>
{/if}