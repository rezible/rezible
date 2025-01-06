<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";

	let containerEl = $state<HTMLElement>();

	timeline.componentSetup(() => containerEl);
</script>

<div class="h-full" bind:this={containerEl}></div>
