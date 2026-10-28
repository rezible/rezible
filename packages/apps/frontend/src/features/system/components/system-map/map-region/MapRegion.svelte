<script lang="ts">
	import { Handle, Position } from "@xyflow/svelte";
	import {
		REGION_LABEL_HEIGHT,
		REGION_PADDING,
		regionContentBox,
	} from "$features/system/lib/system-map/layout";
	import type { RegionNodeProps } from "../projection";

	type Props = RegionNodeProps;
	const { data, selected, width = 0, height = 0 }: Props = $props();

	const level = $derived(data.region.level);
	const content = $derived(regionContentBox({ x: 0, y: 0, width, height }));
</script>

<!-- Svelte Flow draws an edge only between handles; the edge component ignores their position. -->
<Handle type="target" position={Position.Top} class="pointer-events-none opacity-0" aria-hidden="true" />
<div class="map-region h-full w-full" title={data.label}>
	<svg class="shape-svg" {width} {height} viewBox="0 0 {width} {height}" aria-hidden="true">
		<path
			d={data.path}
			class="shape"
			class:landscape={level === "landscape"}
			class:system={level === "system"}
			class:highlighted={data.highlighted}
			class:footprint={data.footprint}
			class:selected
		/>
	</svg>
	<!-- The landscape band draws its labels over the region contents instead; see MapRegionLabels. -->
	{#if data.band !== "landscape"}
		<div
			class="label-band absolute flex min-w-0 items-end"
			style:left="{content.x + REGION_PADDING}px"
			style:top="{content.y}px"
			style:width="{content.width - 2 * REGION_PADDING}px"
			style:height="{REGION_PADDING + REGION_LABEL_HEIGHT}px"
		>
			<span
				class="name text-foreground truncate font-semibold uppercase"
				class:landscape={level === "landscape"}
			>
				{data.label}
			</span>
		</div>
	{/if}
</div>
<Handle type="source" position={Position.Top} class="pointer-events-none opacity-0" aria-hidden="true" />

<style>
	/* Only the outline takes the pointer, so the corners outside it fall through to the region behind. */
	:global(.svelte-flow__node-system-map-region) {
		pointer-events: none !important;
	}

	:global(.svelte-flow__node:focus-visible) > .map-region .shape {
		stroke: var(--ring);
		stroke-width: calc(2px / var(--map-zoom, 1));
	}

	.shape-svg {
		position: absolute;
		inset: 0;
		overflow: visible;
	}

	.shape {
		--tier: var(--chart-4);
		--surface: var(--card);
		pointer-events: visiblePainted;
		cursor: pointer;
		/*
		 * Opaque, so a system's hue reads on its own rather than mixed with its landscape's. Systems tint the
		 * raised card surface so they sit lighter than the landscape around them in both themes.
		 */
		fill: color-mix(in srgb, var(--tier) 10%, var(--surface));
		stroke: color-mix(in srgb, var(--tier) 35%, transparent);
		stroke-width: calc(1px / var(--map-zoom, 1));
	}

	.landscape {
		--tier: var(--chart-1);
		--surface: var(--background);
	}

	.shape:hover {
		stroke: color-mix(in srgb, var(--tier) 70%, transparent);
	}

	.shape.highlighted {
		stroke: var(--primary);
	}

	.shape.footprint {
		stroke: var(--primary);
		stroke-dasharray: calc(6px / var(--map-zoom, 1)) calc(4px / var(--map-zoom, 1));
	}

	.shape.selected {
		fill: var(--selection);
		stroke: var(--primary);
		stroke-width: calc(2px / var(--map-zoom, 1));
		stroke-dasharray: none;
	}

	.label-band {
		padding-bottom: 8px;
	}

	/* Constant on screen, capped so the name fits the label band at any zoom. */
	.name {
		line-height: 1.2;
		letter-spacing: 0.06em;
		font-size: min(calc(13px / var(--map-zoom, 1)), 26px);
	}

	.name.landscape {
		letter-spacing: 0.14em;
	}
</style>
