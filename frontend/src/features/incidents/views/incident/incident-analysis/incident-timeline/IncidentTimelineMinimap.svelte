<script lang="ts">
	import { SvelteMap } from "svelte/reactivity";
	import { type TimelineState } from "./timelineState.svelte";

	type Props = {
		timelineState: TimelineState;
	}
	const { timelineState }: Props = $props();

	let itemStarts = new SvelteMap<string, number>();
	timelineState.items.on("*", (ev, p) => {
		p.items.forEach(id => {
			if (ev === "remove") {
				itemStarts.delete(ev);
				return;
			}
			const item = timelineState.items.get(id);
			if (!item || item.type !== "box") return; 
			itemStarts.set(id.toString(), new Date(item.start).valueOf());
		});
	});

	// TODO: this will be slow eventually
	const [rangeStart, rangeEnd] = $derived.by(() => {
		let min = timelineState.incidentWindow?.start.valueOf() || 0;
		let max = timelineState.incidentWindow?.end.valueOf() || 0;
		itemStarts.values().forEach(v => {
			const val = v.valueOf();
			if (!min || val < min) min = val;
			if (!max || val > max) max = val;
		});
		return [min, max];
	});
	const rangeLength = $derived(rangeEnd - rangeStart);

	const rectWidth = 5;

	const windowStart = $derived(timelineState.viewWindow?.start.valueOf() || rangeStart);
	const windowEnd = $derived(timelineState.viewWindow?.end.valueOf() || rangeEnd);
	const viewWindowX = $derived(((windowStart - rangeStart) / rangeLength) * 100);
	const viewWindowWidth = $derived(((windowEnd - windowStart) / rangeLength) * 100);

	let minimapEl = $state<SVGElement>(null!);

	const onMinimapClicked = (e: MouseEvent) => {
		const clickPosPct = e.offsetX / minimapEl.clientWidth; 
		const windowMidPoint = windowStart + ((windowEnd - windowStart) / 2);
		const rangePosPoint = rangeLength * clickPosPct;

		// TODO: center view window on click pos
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<svg role="img" width="100%" height="100%" viewBox="0 0 100 100%" onclick={onMinimapClicked} bind:this={minimapEl}>
	{#each itemStarts.entries() as [id, start] (id)}
		{@const pct = Math.min(100 - rectWidth, Math.round(((start - rangeStart) / rangeLength) * 100))}
		<rect 
			x="{pct}%" y="0"
			width="5px" height="100%" 
			shape-rendering="crispEdges" 
			class="cursor-pointer fill-accent"
			fill-opacity="30%"
			style="stroke: transparent; stroke-width: 2;"></rect>
	{/each}
	
	<rect x="{viewWindowX}%" y="0"
		width="{viewWindowWidth}%" height="100%" 
		shape-rendering="crispEdges"
		class="fill-primary"
		fill-opacity="25%"
		style="stroke: transparent; stroke-width: 2;"></rect>
</svg>