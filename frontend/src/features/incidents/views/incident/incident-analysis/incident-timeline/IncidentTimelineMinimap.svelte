<script lang="ts">
	import { SvelteMap } from "svelte/reactivity";
	import { ElementSize } from "runed";
	import type { TimelineState } from "./timelineState.svelte";

	type Props = {
		timelineState: TimelineState;
	}
	const { timelineState }: Props = $props();

	let eventItems = new SvelteMap<string, number>();
	timelineState.items.on("*", (ev, p) => {
		p.items.forEach(id => {
			const sId = id.toString();
			if (ev === "remove") {
				eventItems.delete(sId);
			} else {
			const item = timelineState.items.get(id);
				if (!item) return; 
				if (item.subgroup === "events") {
					eventItems.set(sId, new Date(item.start).valueOf());
				} else if (item.subgroup === "milestones") {

				}
			}
		});
	});

	const viewBounds = $derived(timelineState.viewBounds);
	const viewBoundsLength = $derived((viewBounds.end - viewBounds.start) || 1);

	let containerEl = $state<HTMLElement>(null!);
	const containerSize = new ElementSize(() => containerEl);
	const containerWidth = $derived(containerSize.width);

	const viewWindow = $derived(timelineState.viewWindow);
	const viewWindowLength = $derived((viewWindow.end - viewWindow.start) || 1);

	const windowHighlightX = $derived(((viewWindow.start - viewBounds.start) / viewBoundsLength) * containerWidth || 0);
	const windowHighlightWidth = $derived(containerWidth * Math.min(1, viewWindowLength / viewBoundsLength));

	const onMinimapClicked = (e: MouseEvent) => {
		const clickPosPct = e.offsetX / containerWidth;
		const viewBoundsPoint = viewBounds.start + (viewBoundsLength * clickPosPct);
		timelineState.timeline?.moveTo(viewBoundsPoint);
	}

	const incidentStartPct = $derived((timelineState.incidentWindow.start - viewBounds.start) / viewBoundsLength);
	const incidentWidth = $derived(containerWidth * ((timelineState.incidentWindow.end - timelineState.incidentWindow.start) / viewBoundsLength));
	const incidentHighlightX = $derived(incidentStartPct * containerWidth);

	const eventRectWidth = 5;
	const maxEventXPos = $derived(containerWidth - eventRectWidth);
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="w-full h-full" role="presentation" bind:this={containerEl} onclick={onMinimapClicked}>
	<svg role="img" width="100%" height="100%" viewBox="0 0 100 100%">
		<rect x={incidentHighlightX} y="0"
			width="{incidentWidth}" height="100%" 
			shape-rendering="crispEdges"
			class="fill-secondary"
			fill-opacity="10%"
			style="stroke: transparent; stroke-width: 2;"></rect>

		{#each eventItems.entries() as [id, start] (id)}
			{@const xPos = Math.min(maxEventXPos, Math.round(((start - viewBounds.start) / viewBoundsLength) * containerWidth))}
			<rect 
				x={xPos} y="0"
				width={eventRectWidth} height="100%" 
				shape-rendering="crispEdges" 
				class="cursor-pointer fill-accent"
				fill-opacity="30%"
				style="stroke: transparent; stroke-width: 2;"></rect>
		{/each}
		
		<rect x={windowHighlightX} y="0"
			width={windowHighlightWidth} height="100%" 
			shape-rendering="crispEdges"
			class="fill-primary"
			fill-opacity="25%"
			style="stroke: transparent; stroke-width: 2;"></rect>
	</svg>
</div>