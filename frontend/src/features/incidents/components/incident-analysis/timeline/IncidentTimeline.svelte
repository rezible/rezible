<script lang="ts">
	import { onMount } from "svelte";
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";

	let containerEl = $state<HTMLElement>();

	onMount(() => {
		if (!containerEl) return;
		timeline.mount(containerEl);
		return () => timeline.unmount();
	});
</script>

<div class="h-full" bind:this={containerEl}></div>
