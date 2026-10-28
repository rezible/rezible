<script lang="ts">
	import { Handle, Position } from "@xyflow/svelte";
	import { categoryIcon } from "$features/system/lib/system-map/category-icon";
	import { cn } from "$lib/utils";
	import PlacementChips from "../placement-chips/PlacementChips.svelte";
	import type { CardNodeProps } from "../projection";

	type Props = CardNodeProps;
	const { id, data, selected }: Props = $props();

	/** The landscape band hides cards, except a selected one drawn as its icon. */
	const compact = $derived(data.band === "landscape");
	const CategoryIcon = $derived(categoryIcon(data.card.entity.category));
</script>

<!-- Svelte Flow draws an edge only between handles; the edge component ignores their position. -->
<Handle type="target" position={Position.Top} class="pointer-events-none opacity-0" aria-hidden="true" />
{#if compact}
	<div
		class="map-card compact selected absolute top-1/2 left-1/2 flex items-center justify-center rounded-full"
	>
		<CategoryIcon class="icon size-5 shrink-0" aria-hidden="true" />
	</div>
{:else}
	<div
		class={cn(
			"map-card text-card-foreground relative flex h-full w-full flex-col items-center justify-center rounded-[32px] px-4 text-center",
			data.highlighted && "highlighted",
			data.footprint && "footprint",
			selected && "selected"
		)}
	>
		<CategoryIcon class="icon mb-1 size-5 shrink-0" aria-hidden="true" />
		<span class="line-clamp-2 text-base leading-5 font-medium break-words" title={data.label}>
			{data.label}
		</span>
		<span
			class="text-muted-foreground max-w-full truncate text-[13px] leading-4"
			title={data.card.entity.kind}
		>
			{data.card.entity.kind}
		</span>
		{#if data.placement.targets.length > 0}
			<div class="mt-1.5 max-w-full">
				<PlacementChips cardId={id} cardSelected={selected} placement={data.placement} />
			</div>
		{/if}
		{#if data.alsoInLabel}
			<span
				class="bg-muted-foreground absolute top-4 right-5 size-2 rounded-full"
				title={data.alsoInLabel}
				aria-label={data.alsoInLabel}
				role="img"
			></span>
		{/if}
	</div>
{/if}
<Handle type="source" position={Position.Top} class="pointer-events-none opacity-0" aria-hidden="true" />

<style>
	:global(.svelte-flow__node:focus-visible) > .map-card {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}

	.map-card {
		background: color-mix(in srgb, var(--chart-2) 7%, var(--card));
		border: 1px solid color-mix(in srgb, var(--chart-2) 35%, transparent);
	}

	.map-card:hover {
		border-color: color-mix(in srgb, var(--chart-2) 70%, transparent);
	}

	.map-card :global(.icon) {
		color: var(--chart-2);
	}

	.map-card.highlighted {
		border-color: var(--primary);
	}

	.map-card.footprint {
		border-color: var(--primary);
		border-style: dashed;
	}

	.map-card.selected {
		background: var(--selection);
		color: var(--selection-foreground);
		border: 2px solid var(--primary);
	}

	/* Constant on screen at the landscape band, where the card's own box is too small to see. */
	.map-card.compact {
		width: min(calc(36px / var(--map-zoom, 1)), 112px);
		height: min(calc(36px / var(--map-zoom, 1)), 112px);
		translate: -50% -50%;
		border-width: calc(2px / var(--map-zoom, 1));
	}

	.compact :global(.icon) {
		width: 55%;
		height: 55%;
	}
</style>
