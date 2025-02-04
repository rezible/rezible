<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";
	import { Button } from "svelte-ux";
	import EventEditorDialog from "./event-editor/EventEditorDialog.svelte";

	let containerEl = $state<HTMLElement>();

	timeline.componentSetup(() => containerEl);

	let creatingEvent = $state(false);
</script>

<div class="h-full overflow-y-hidden" bind:this={containerEl}></div>

<EventEditorDialog bind:creatingEvent />
