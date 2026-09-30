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

<Sidebar.Root bind:ref {collapsible} class="border-sidebar-border" {...restProps}>
	<Sidebar.Header
		class={cn(controller.isDefault && "h-14 shrink-0 justify-center border-b border-sidebar-border p-0")}
	>
		{#if controller.isDefault}
			<a
				href="/"
				aria-label="Rezible home"
				data-sveltekit-preload-data={controller.preloadHome}
				data-sveltekit-preload-code={controller.preloadHome}
				class={cn(
					"flex h-14 items-center gap-2 rounded-md focus-visible:outline-2 focus-visible:outline-offset-[-4px] focus-visible:outline-sidebar-ring",
					controller.expanded ? "px-4" : "justify-center"
				)}
			>
				<img src="/images/logo.svg" alt="" class="size-7" />
				{#if controller.expanded}
					<span class="text-base font-semibold">Rezible</span>
				{/if}
			</a>
		{:else}
			<Sidebar.Menu>
				<Sidebar.MenuItem>
					<Button href="/" variant="ghost" size="lg">
						<RiArrowLeftLine /> Back to app
					</Button>
				</Sidebar.MenuItem>
			</Sidebar.Menu>
		{/if}

		{#if controller.showSearch}
			{@const placeholder = controller.model.search?.placeholder}
			<Sidebar.Input bind:value={controller.searchQuery} {placeholder} aria-label={placeholder} />
		{/if}
	</Sidebar.Header>

	<Sidebar.Content>
		{#each controller.groups as group (group.label ?? group.items.map((item) => item.href).join("|"))}
			<Sidebar.Group>
				{#if group.label && controller.expanded}
					<Sidebar.GroupLabel>{group.label}</Sidebar.GroupLabel>
				{/if}
				<Sidebar.GroupContent>
					<Sidebar.Menu
						class={cn(
							"gap-1",
							!group.label && "pt-1",
							!controller.expanded && "items-center pt-1"
						)}
					>
						{#each group.items as item (item.href)}
							<NavMenuItem {item} />
						{/each}
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>
		{/each}
	</Sidebar.Content>

	<Sidebar.Footer class={cn(!controller.expanded && "items-center pb-3")}>
		{#if controller.showUserMenu}
			<NavUserMenu />
		{/if}
	</Sidebar.Footer>
</Sidebar.Root>
