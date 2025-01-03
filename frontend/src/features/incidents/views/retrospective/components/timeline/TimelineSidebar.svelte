<script lang="ts">
	import TimelineMinimap from "./TimelineMinimap.svelte";

	type Props = { incidentId: string; };
	let { incidentId }: Props = $props();

	let containerEl = $state<HTMLElement>();
	let eventsEl = $state<HTMLElement>();
	const events = [
		{label: "foo", size: 80},
		{label: "bar", size: 110},
		{label: "baz", size: 40},
		{label: "qux", size: 180}
	];

	const onEventClicked = (id: string) => {

	}
</script>

<div class="h-full grid grid-cols-[40px_minmax(0,_1fr)] bg-secondary/10" bind:this={containerEl}>
	<!--div class="grid">
		{#if containerEl && eventsEl}
			<TimelineMinimap {containerEl} {eventsEl} {events} {onEventClicked} />
		{/if}
	</div-->

	<div class="h-full grid auto-rows-min overflow-y-auto border-e" bind:this={eventsEl}>
		{#each events as e, i}
			<div style="height: {e.size}px">{e.label}</div>
		{/each}
	</div>
</div>
