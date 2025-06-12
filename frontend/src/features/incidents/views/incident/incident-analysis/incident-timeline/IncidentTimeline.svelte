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
	import { fromAbsolute } from "@internationalized/date";
	import type { TimelineItem } from "vis-timeline";
	import { useIncidentAnalysis } from "../analysisState.svelte";

	const analysis = useIncidentAnalysis();
	
	const timelineState = new TimelineState();
	setIncidentTimeline(timelineState);
	setEventDialog(new EventDialogState(timelineState));
	setMilestonesDialog(new MilestonesDialogState());

	let containerRef = $state<HTMLElement>(null!);
	watch(() => containerRef, ref => {timelineState.mountTimeline(ref)});

	const closeContextMenu = () => {
		analysis.contextMenu = {};
	};

	const onContextMenu = (e: MouseEvent | PointerEvent) => {
		e.preventDefault();

		const clickPos = {x: e.x, y: e.y};

		let item: TimelineItem | undefined = undefined;
		let wasTimeline = true;
		if (e.target && "parentNode" in e.target) {
			const ref = e.target as HTMLElement | null;
			wasTimeline = !!ref?.classList.value.includes("vis-");

			let el = ref;
			while (!!el && el !== containerRef) {
				if ("vis-item" in el) {
					item = el["vis-item"] as TimelineItem;
					timelineState.timeline?.setSelection(item.id);
					// TODO
				}
				el = el.parentElement;
			}
		}

		const containerRect = containerRef.getBoundingClientRect();

		const pct = (e.x - containerRect.x) / containerRect.width;

		const timeRange = (wasTimeline && timelineState.timeline) ? timelineState.viewWindow : timelineState.viewBounds;
		const timestampMs = timeRange.start + (timeRange.end - timeRange.start) * pct;

		const timestamp = fromAbsolute(timestampMs, timelineState.viewState.timezone);

		analysis.contextMenu = {
			timeline: {
				clickPos,
				timestamp,
				item,
				containerRect,
				close: closeContextMenu,
			}
		};
	}
</script>

<div
	id="timeline-minimap-container"
	class="w-full h-full overflow-hidden border relative"
	role="presentation" 
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

{#if !!analysis.contextMenu.timeline}
	<IncidentTimelineContextMenu {...analysis.contextMenu.timeline} />
{/if}

<EventDialog />
<MilestonesDialog />
