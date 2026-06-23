<script lang="ts">
    import type { ComponentProps } from "svelte";
    import * as Sidebar from "$components/ui/sidebar";
	import { cn } from "$lib/utils";
	import NavMenuItem from "./NavMenuItem.svelte";
	import NavUserMenu from "./NavUserMenu.svelte";
	import RiArrowLeftLine from "remixicon-svelte/icons/arrow-left-line";
	import Button from "$src/components/ui/button/button.svelte";
	import { initAppSidebarController } from "./controller.svelte";
    
    let {
        ref = $bindable(null),
        collapsible = "icon",
        ...restProps
    }: ComponentProps<typeof Sidebar.Root> = $props();

	const controller = initAppSidebarController();
</script>

<Sidebar.Root bind:ref {collapsible} {...restProps}>
	<Sidebar.Header>
        <Sidebar.Menu>
            <Sidebar.MenuItem 
				data-sveltekit-preload-data={controller.preloadHome} 
				data-sveltekit-preload-code={controller.preloadHome}
			>
				{#if controller.isDefault}
					<Sidebar.MenuButton size="lg">
						{#snippet child({ props })}
							<a {...props} href="/" class="text-2xl text-base flex items-center gap-2">
								<img src="/images/logo.svg" alt="logo" class={cn("fill-neutral size-10")} />
								<span data-open={controller.isOpen ? true : undefined} class="hidden data-open:inline">
									Rezible
								</span>
							</a>
						{/snippet}
					</Sidebar.MenuButton>
				{:else}
					<Button href="/" variant="ghost" size="lg">
						<RiArrowLeftLine /> Back to app
					</Button>
				{/if}
            </Sidebar.MenuItem>
        </Sidebar.Menu>

		{#if controller.showSearch}
			<Sidebar.Input
				bind:value={controller.searchQuery}
				placeholder={controller.model?.search?.placeholder}
			/>
		{/if}
	</Sidebar.Header>

	<Sidebar.Content>
		{#each controller.groups as group (group.label ?? group.items.map((item) => item.href).join("|"))}
			<Sidebar.Group>
				{#if group.label && controller.isOpen}
					<Sidebar.GroupLabel>{group.label}</Sidebar.GroupLabel>
				{/if}
				<Sidebar.GroupContent>
					<Sidebar.Menu class={cn("gap-1", !group.label && "pt-1")}>
						{#each group.items as item (item.href)}
							<NavMenuItem {item} />
						{/each}
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>
		{/each}
	</Sidebar.Content>

	{#if controller.isDefault}
		<Sidebar.Footer>
			<NavUserMenu />
		</Sidebar.Footer>
	{/if}
</Sidebar.Root>
