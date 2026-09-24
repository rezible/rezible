<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Checkbox from "$components/ui/checkbox";
	import * as Popover from "$components/ui/popover";
	import { getMapCategoryDisplay } from "$features/system/lib/system-map/category";
	import type { SystemMapHandle } from "$features/system/components/system-map";
	import { systemMapCategoryOptions, useSystemMapViewController } from "../../controller.svelte";

	type Props = { map: SystemMapHandle | undefined };
	let { map }: Props = $props();

	const view = useSystemMapViewController();
	let categoriesOpen = $state(false);
	let kindsOpen = $state(false);
</script>

<header class="border-border flex shrink-0 flex-wrap items-center justify-between gap-2 border-b px-3 py-2">
	<div class="flex min-w-0 flex-wrap items-center gap-2">
		<Popover.Root bind:open={categoriesOpen}>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" size="sm" aria-expanded={categoriesOpen}>
						Categories{#if view.filters.categories.length > 0}
							({view.filters.categories.length}){/if}
					</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content align="start" class="max-h-80 w-64 overflow-y-auto p-2">
				<p class="px-2 pb-2 text-xs font-medium">Filter by category</p>
				<div class="flex flex-col gap-1" role="group" aria-label="Categories">
					{#each systemMapCategoryOptions as category (category)}
						<div class="flex items-center gap-2 rounded-sm px-2 py-1">
							<Checkbox.Root
								checked={view.isCategorySelected(category)}
								onCheckedChange={(checked) => view.categoryChanged(category, checked)}
								aria-label={`Include ${getMapCategoryDisplay(category).categoryLabel}`}
							/>
							<span class="text-sm">{getMapCategoryDisplay(category).categoryLabel}</span>
						</div>
					{/each}
				</div>
			</Popover.Content>
		</Popover.Root>

		<Popover.Root bind:open={kindsOpen}>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" size="sm" aria-expanded={kindsOpen}>
						Kinds{#if view.filters.kinds.length > 0}
							({view.filters.kinds.length}){/if}
					</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content align="start" class="max-h-80 w-72 overflow-y-auto p-2">
				<p class="px-2 pb-2 text-xs font-medium">Filter by exact kind</p>
				{#if view.availableKinds.length === 0}
					<p class="text-muted-foreground px-2 py-1 text-sm">No kinds are in the loaded records.</p>
				{:else}
					<div class="flex flex-col gap-1" role="group" aria-label="Kinds">
						{#each view.availableKinds as kind (kind)}
							<div class="flex min-w-0 items-center gap-2 rounded-sm px-2 py-1">
								<Checkbox.Root
									checked={view.isKindSelected(kind)}
									onCheckedChange={(checked) => view.kindChanged(kind, checked)}
									aria-label={`Include kind ${kind}`}
								/>
								<span class="min-w-0 break-all text-sm">{kind}</span>
							</div>
						{/each}
					</div>
				{/if}
			</Popover.Content>
		</Popover.Root>

		{#if view.hasFilters}
			<Button variant="ghost" size="sm" onclick={view.clearFilters}>Clear filters</Button>
		{/if}
	</div>

	<div class="flex shrink-0 flex-wrap items-center gap-2">
		<Button variant="outline" size="sm" disabled={!map} onclick={() => map?.fit()}>Fit</Button>
		<Button variant="outline" size="sm" disabled={!map} onclick={() => map?.recenter()}>Recenter</Button>
		<Button
			variant={view.inspectorOpen ? "secondary" : "outline"}
			size="sm"
			aria-pressed={view.inspectorOpen}
			onclick={view.toggleInspector}
		>
			Inspector
		</Button>
		<Button variant="outline" size="sm" onclick={view.refresh}>Refresh</Button>
	</div>
</header>
