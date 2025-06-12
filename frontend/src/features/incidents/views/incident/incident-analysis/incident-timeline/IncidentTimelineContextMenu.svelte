<script lang="ts">
	import { ElementSize } from "runed";
	import { Button } from "svelte-ux";
	import { useEventDialog } from "./event-dialog/dialogState.svelte";
	import { ZonedDateTime } from "@internationalized/date";
	import type { TimelineItem } from "vis-timeline";
	import { useIncidentTimeline } from "./timelineState.svelte";
	import { mdiPencilCircle, mdiPlusCircle } from "@mdi/js";
	import AnalysisContextMenu from "$src/features/incidents/components/analysis-context-menu/AnalysisContextMenu.svelte";

	type Props = {
		containerRect: DOMRect;
		item?: TimelineItem;
		clickPos: {x: number; y: number};
		timestamp: ZonedDateTime;
		close: () => void;
	};
	const {item, containerRect, clickPos, timestamp, close}: Props = $props();

	const timelineState = useIncidentTimeline();
	const eventDialog = useEventDialog();

	const onClicked = (e: MouseEvent) => {e.stopPropagation()};

	const onAddEventClick = () => {
		eventDialog.setCreating({timestamp: timestamp.toAbsoluteString()})
		close();
	}

	const itemId = $derived(item?.id.toString())
	const event = $derived(itemId && timelineState.events.events.find(e => (e.id === itemId)));

	const onEditEventClick = () => {
		if (!event) return;
		eventDialog.setEditing(event);
		close();
	}
</script>

<AnalysisContextMenu title="Timeline Actions" {containerRect} {clickPos}>
	<div 
		id="timeline-ctx-container"
		onclick={onClicked}
		role="presentation"
	>
		{#if event}
			<Button variant="fill-light" icon={mdiPencilCircle} rounded={false} classes={{root: "w-full gap-2"}} on:click={onEditEventClick}>
				Edit Event
			</Button>
		{/if}
		
		<Button variant="fill-light" icon={mdiPlusCircle} rounded={false} classes={{root: "w-full gap-2"}} on:click={onAddEventClick}>
			Add New Event
		</Button>
	</div>
</AnalysisContextMenu>