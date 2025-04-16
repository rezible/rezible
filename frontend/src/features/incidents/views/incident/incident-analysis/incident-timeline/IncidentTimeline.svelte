<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import "./timeline-styles.css";

	import { TimelineState, setIncidentTimeline } from "./timelineState.svelte";
	import { EventDialogState, setEventDialog } from "./event-dialog/dialogState.svelte";
	
	import IncidentTimelineActionsBar from "./IncidentTimelineActionsBar.svelte";
	import EventDialog from "./event-dialog/EventDialog.svelte";
	import MilestonesDialog from "./milestones-dialog/MilestonesDialog.svelte";
	import { MilestonesDialogState, setMilestonesDialog } from "./milestones-dialog/dialogState.svelte";
	import { watch } from "runed";
	
	const timelineState = new TimelineState();
	setIncidentTimeline(timelineState);
	setEventDialog(new EventDialogState());
	setMilestonesDialog(new MilestonesDialogState());

	let containerRef = $state<HTMLElement>();
	watch(() => containerRef, ref => {
		if (!containerRef) return;
		timelineState.mountTimeline(containerRef);
	})
</script>

<div class="w-full h-full overflow-y-hidden" bind:this={containerRef}></div>

<IncidentTimelineActionsBar />

<EventDialog />
<MilestonesDialog />
