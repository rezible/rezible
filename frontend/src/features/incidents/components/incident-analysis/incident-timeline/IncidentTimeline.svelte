<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import "./timeline-styles.css";
	import { timeline } from "./timeline.svelte";
	import IncidentTimelineActionsBar from "./IncidentTimelineActionsBar.svelte";
	import EventDialog from "./event-dialog/EventDialog.svelte";
	import MilestonesDialog from "./milestones-dialog/MilestonesDialog.svelte";

	let containerEl = $state<HTMLElement>();
	timeline.setup(() => containerEl);
</script>

<div class="w-full h-full overflow-y-hidden" bind:this={containerEl}></div>

<IncidentTimelineActionsBar />

<EventDialog />
<MilestonesDialog />
