<script lang="ts">
	import { ViewportPortal } from "@xyflow/svelte";
	import {
		REGION_LABEL_HEIGHT,
		REGION_PADDING,
		regionContentBox,
	} from "$features/system/lib/system-map/layout";
	import { cn } from "$lib/utils";
	import { REGION_LABEL_Z, type FlowNode, type RegionFlowNode } from "../projection";

	type Props = { nodes: FlowNode[] };
	const { nodes }: Props = $props();

	/**
	 * Landscape names sit top-left inside the outline, on the bottom of the label band so a name taller than
	 * the band grows into the margin above it rather than over the first system. System names are centred.
	 */
	const labels = $derived(
		nodes
			.filter((node): node is RegionFlowNode => node.type === "system-map-region")
			.map((region) => {
				const box = { ...region.position, width: region.width ?? 0, height: region.height ?? 0 };
				const level = region.data.region.level;
				let area = box;
				if (level === "landscape") {
					const content = regionContentBox(box);
					area = { ...content, height: REGION_PADDING + REGION_LABEL_HEIGHT };
				}
				return { id: region.id, label: region.data.label, level, area };
			})
	);
</script>

<!-- Drawn over the regions at the landscape band, where a label band is too small to read on screen. -->
<ViewportPortal target="front" class="pointer-events-none absolute inset-0" style="z-index: {REGION_LABEL_Z}">
	{#each labels as { id, label, level, area } (id)}
		<div
			class={cn(
				"region-labels absolute flex",
				level === "landscape" ? "items-end justify-start" : "items-center justify-center"
			)}
			style:left="{area.x}px"
			style:top="{area.y}px"
			style:width="{area.width}px"
			style:height="{area.height}px"
		>
			<span
				class="name text-foreground max-w-full truncate font-semibold uppercase"
				class:landscape={level === "landscape"}
			>
				{label}
			</span>
		</div>
	{/each}
</ViewportPortal>

<style>
	.region-labels {
		padding: calc(4px / var(--map-zoom, 1)) calc(12px / var(--map-zoom, 1));
	}

	.name {
		line-height: 1.3;
		letter-spacing: 0.06em;
		font-size: calc(13px / var(--map-zoom, 1));
	}

	.name.landscape {
		letter-spacing: 0.14em;
	}
</style>
