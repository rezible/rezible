<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import "./timeline-styles.css";

	import { watch } from "runed";
	import { TimelineState, setIncidentTimeline } from "./timelineState.svelte";
	import { EventDialogState, setEventDialog } from "./event-dialog/dialogState.svelte";
	
	import IncidentTimelineActionsBar from "./IncidentTimelineActionsBar.svelte";
	import EventDialog from "./event-dialog/EventDialog.svelte";
	import MilestonesDialog from "./milestones-dialog/MilestonesDialog.svelte";
	import { MilestonesDialogState, setMilestonesDialog } from "./milestones-dialog/dialogState.svelte";
	import IncidentTimelineMinimap from "./IncidentTimelineMinimap.svelte";
	
	const timelineState = new TimelineState();
	setIncidentTimeline(timelineState);
	setEventDialog(new EventDialogState(timelineState));
	setMilestonesDialog(new MilestonesDialogState());

	let containerRef = $state<HTMLElement>();
	watch(() => containerRef, ref => {
		if (ref) timelineState.mountTimeline(ref);
	})
</script>

<div class="w-full h-full overflow-y-hidden border">
	<div class="w-full" style="height: 90%" bind:this={containerRef}></div>
	<div class="w-full border-t" style="height: 10%">
		<IncidentTimelineMinimap {timelineState} />
	</div>
</div>

<IncidentTimelineActionsBar />

<EventDialog />
<MilestonesDialog />
