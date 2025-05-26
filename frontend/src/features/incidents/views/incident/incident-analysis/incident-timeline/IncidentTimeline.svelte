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
	import IncidentTimelineContextMenu from "./IncidentTimelineContextMenu.svelte";
	
	const timelineState = new TimelineState();
	setIncidentTimeline(timelineState);
	setEventDialog(new EventDialogState(timelineState));
	setMilestonesDialog(new MilestonesDialogState());

	let containerRef = $state<HTMLElement>(null!);
	watch(() => containerRef, ref => {timelineState.mountTimeline(ref)});

	type XYPosition = {x: number, y: number};
	type ContextMenuState = {pos: XYPosition}
	let ctxMenu = $state<ContextMenuState>();
	const onContextMenu = (e: MouseEvent | PointerEvent) => {
		e.preventDefault();

		if (e.target && "parentNode" in e.target) {
			const node = e.target.parentNode as Record<string, any>;
			if ("vis-item" in node) {
				const timelineItem = node["vis-item"];
				console.log(timelineItem);
				// TODO
			}
		}

		const rec = containerRef.getBoundingClientRect();
		const pos = {x: Math.round(e.x - rec.x), y: Math.round(e.y - rec.y)};
		ctxMenu = {pos};
	}

	const closeContextMenu = () => {ctxMenu = undefined};
	const onContextMenuClicked = (e: MouseEvent) => {e.stopPropagation()};
</script>

<div class="w-full h-full overflow-hidden border relative" role="presentation" 
	oncontextmenu={onContextMenu}
	onclickcapture={closeContextMenu}
>
	<div 
		class="w-full"
		style="height: 90%"
		bind:this={containerRef}></div>
	<div
		class="w-full border-t"
		style="height: 10%">
		<IncidentTimelineMinimap {timelineState} />
	</div>
</div>

<div class="absolute top-2 right-2 w-fit mx-auto">
	<IncidentTimelineActionsBar />
</div>

{#if ctxMenu}
	<div 
		id="timeline-ctx-container"
		class="w-fit"
		style="position: absolute; left: {ctxMenu.pos.x}px; top: {ctxMenu.pos.y}px"
		role="presentation"
		onclick={onContextMenuClicked}
	>
		<IncidentTimelineContextMenu />
	</div>
{/if}

<EventDialog />
<MilestonesDialog />
