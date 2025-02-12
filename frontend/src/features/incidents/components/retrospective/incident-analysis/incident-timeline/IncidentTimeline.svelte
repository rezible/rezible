<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";
	import EventEditorDialog from "./event-editor/EventEditorDialog.svelte";
	import IncidentTimelineActionsBar from "./IncidentTimelineActionsBar.svelte";

	let containerEl = $state<HTMLElement>();
	timeline.setup(() => containerEl);
</script>

<div class="w-full h-full overflow-y-hidden" bind:this={containerEl}></div>

<IncidentTimelineActionsBar />

<EventEditorDialog />
