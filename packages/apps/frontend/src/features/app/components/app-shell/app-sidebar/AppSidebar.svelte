<script lang="ts">
    import type { ComponentProps } from "svelte";
    import * as Sidebar from "$components/ui/sidebar";
	import NavHeader from "./NavHeader.svelte";
	import NavPrimaryMenu from "./NavPrimaryMenu.svelte";
	import NavUserMenu from "./NavUserMenu.svelte";
	import NavSecondaryMenu from "./NavSecondaryMenu.svelte";
    
    let {
        ref = $bindable(null),
        collapsible = "icon",
        ...restProps
    }: ComponentProps<typeof Sidebar.Root> = $props();
</script>

<Sidebar.Root bind:ref {collapsible} {...restProps}>
    <Sidebar.Header>
        <NavHeader />
    </Sidebar.Header>

    <Sidebar.Content>
        <NavPrimaryMenu />
        <NavSecondaryMenu class="mt-auto" />
    </Sidebar.Content>

    <Sidebar.Footer>
        <NavUserMenu />
    </Sidebar.Footer>
</Sidebar.Root>