<script lang="ts">
	import IncidentTimeline from "$src/features/incidents/components/incident-timeline/IncidentTimeline.svelte";
	import { cls } from "svelte-ux";
	import TimelineMinimap from "./TimelineMinimap.svelte";

	type Props = { incidentId: string; };
	let { incidentId }: Props = $props();

	let containerEl = $state<HTMLElement>();
	const events = [
		{label: "foo", size: 80},
		{label: "bar", size: 110},
		{label: "baz", size: 40},
		{label: "qux", size: 180}
	];

	const onEventClicked = (id: string) => {

	}
</script>

<div class="grid grid-cols-[40px_minmax(0,_1fr)] min-h-0 bg-secondary/10">
	{#if containerEl}
		<TimelineMinimap {containerEl} {events} {onEventClicked} />
	{/if}

	<div class="h-full grid overflow-y-auto border-e" bind:this={containerEl}>
		{#each events as e, i}
			<div style="height: {e.size}px">{e.label}</div>
		{/each}
	</div>
</div>
