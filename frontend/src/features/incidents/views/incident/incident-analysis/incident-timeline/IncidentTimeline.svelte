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
	import type { ComponentProps } from "svelte";
	
	const timelineState = new TimelineState();
	setIncidentTimeline(timelineState);
	setEventDialog(new EventDialogState(timelineState));
	setMilestonesDialog(new MilestonesDialogState());

	let containerRef = $state<HTMLElement>(null!);
	watch(() => containerRef, ref => {timelineState.mountTimeline(ref)});

	const minimapContainerId = "timeline-minimap-container";

	const closeContextMenu = () => {ctxMenu = undefined};
	let ctxMenu = $state<ComponentProps<typeof IncidentTimelineContextMenu>>();
	const onContextMenu = (e: MouseEvent | PointerEvent) => {
		e.preventDefault();

		let wasTimeline = true;
		if (e.target && "parentNode" in e.target) {
			const ref = e.target as HTMLElement;
			wasTimeline = ref.classList.value.includes("vis-");

			const node = ref.parentNode as Record<string, any>;
			if ("vis-item" in node) {
				const timelineItem = node["vis-item"];
				console.log(timelineItem);
				// TODO
			}
		}

		const rec = containerRef.getBoundingClientRect();

		const pct = (e.x - rec.x) / rec.width;

		const timeRange = (wasTimeline && timelineState.timeline) ? timelineState.viewWindow : timelineState.viewBounds;
		const timestamp = timeRange.start + (timeRange.end - timeRange.start) * pct;

		ctxMenu = {
			clickPos: {x: e.x, y: e.y},
			timestamp,
			containerRect: rec,
			close: closeContextMenu,
		};
	}
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
		id={minimapContainerId}
		class="w-full border-t"
		style="height: 10%">
		<IncidentTimelineMinimap {timelineState} />
	</div>
</div>

<div class="absolute top-2 right-2 w-fit mx-auto">
	<IncidentTimelineActionsBar />
</div>

{#if ctxMenu}
	<IncidentTimelineContextMenu {...ctxMenu} />
{/if}

<EventDialog />
<MilestonesDialog />
