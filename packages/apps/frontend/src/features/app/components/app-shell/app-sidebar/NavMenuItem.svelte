<script lang="ts">
	import * as Collapsible from "$components/ui/collapsible";
	import * as Sidebar from "$components/ui/sidebar";
	import Icon from "$src/components/icon/Icon.svelte";
    import RiArrowRightSLine from 'remixicon-svelte/icons/arrow-right-s-line';
	import { getActiveStatus, type SidebarItem } from "./sidebar";
	import { page } from "$app/state";

    type Props = {
        item: SidebarItem;
    };
    const { item }: Props = $props();

    const status = $derived(getActiveStatus(page.route.id, item));
    const anyActiveSubItems = $derived(status.subItemsActive.values().some(Boolean));
    const isActive = $derived((status.active && !anyActiveSubItems) ? true : undefined);
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
    <Collapsible.Root open={status.active} class="group/collapsible">
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
                                    <RiArrowRightSLine class="ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                </a>
                            {/snippet}
                        </Sidebar.MenuButton>
                    {/snippet}
                </Collapsible.Trigger>
                <Collapsible.Content>
                    <Sidebar.MenuSub>
                        {#each item.subItems as subItem, idx (subItem.label)}
                            <Sidebar.MenuSubItem>
                                <Sidebar.MenuSubButton isActive={status.subItemsActive.get(idx)}>
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