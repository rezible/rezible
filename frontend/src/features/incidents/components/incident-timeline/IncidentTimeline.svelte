<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";
    import { Button } from "svelte-ux";
    import EventEditorDialog from "./event-editor/EventEditorDialog.svelte";

	let containerEl = $state<HTMLElement>();

	timeline.componentSetup(() => containerEl);

	let creatingEvent = $state(false);
</script>

<div class="flex flex-col gap-2 h-full overflow-y-hidden">
	<div class="min-h-0 flex-1" bind:this={containerEl}></div>
	<div class="h-20 flex flex-row justify-end">
		<Button color="secondary" variant="fill" on:click={() => {creatingEvent = true}}>
			Add Event
		</Button>
	</div>
</div>

<EventEditorDialog bind:creatingEvent />