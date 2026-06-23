<script lang="ts">
	import { resolve } from "$app/paths";
	import * as Collapsible from "$components/ui/collapsible";
	import * as Sidebar from "$components/ui/sidebar";
    import RiArrowRightSLine from 'remixicon-svelte/icons/arrow-right-s-line';
	import { useAppSidebarController, type SidebarItem } from "./controller.svelte";

    type Props = {
        item: SidebarItem;
    };
    const { item }: Props = $props();

    const controller = useAppSidebarController();

    const isActive= $derived(controller.activeStatus.deepestActiveItem === item.href);
    const activeSubs = $derived(controller.activeStatus.activeSubItems.get(item.href));
</script>

{#if !item.subItems}
    <Sidebar.MenuItem>
        <Sidebar.MenuButton {isActive} tooltipContent={item.label}>
            {#snippet child({ props })}
                <a href={resolve(item.href as any)} {...props}>
                    {#if !!item.icon && typeof item.icon === "function"}
                        <item.icon />
                    {/if}
                    {#if !controller.collapsed}
                        {item.label}
                    {/if}
                </a>
            {/snippet}
        </Sidebar.MenuButton>
    </Sidebar.MenuItem>
{:else}
    <Collapsible.Root open={isActive} class="group/collapsible">
        {#snippet child({ props })}
            <Sidebar.MenuItem {...props}>
                <Collapsible.Trigger>
                    {#snippet child({ props })}
                        <Sidebar.MenuButton {...props} {isActive} tooltipContent={item.label}>
                            {#snippet child({ props })}
                                <a href={resolve(item.href as any)} {...props}>
                                    {#if item.icon}
                                        <item.icon />
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
                        {#each item.subItems as subItem (subItem.label)}
                            <Sidebar.MenuSubItem>
                                <Sidebar.MenuSubButton isActive={!!activeSubs && activeSubs.has(subItem.href)}>
                                    {#snippet child({ props })}
                                        <a href={resolve(subItem.href as any)} {...props}>
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
