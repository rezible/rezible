<script lang="ts">
	import { useEventDialog } from "./event-dialog/controller.svelte";
	import { ZonedDateTime } from "@internationalized/date";
	import type { TimelineItem } from "vis-timeline";
	import { useIncidentTimelineController } from "./controller.svelte";
	import { mdiPencilCircle, mdiPlusCircle } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import AnalysisContextMenu from "$features/incidents/components/analysis-context-menu/AnalysisContextMenu.svelte";

	type Props = {
		containerRect: DOMRect;
		item?: TimelineItem;
		clickPos: {x: number; y: number};
		timestamp: ZonedDateTime;
		close: () => void;
	};
	const {item, containerRect, clickPos, timestamp, close}: Props = $props();

	const timelineController = useIncidentTimelineController();
	const eventDialog = useEventDialog();

	const onClicked = (e: MouseEvent) => {e.stopPropagation()};

	const onAddEventClick = () => {
		eventDialog.setCreating({timestamp: timestamp.toAbsoluteString()})
		close();
	}

	const itemId = $derived(item?.id.toString())
	const event = $derived(itemId && timelineController.events.events.find(e => (e.id === itemId)));

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
			<Button onclick={onEditEventClick}>
				Edit Event
			</Button>
		{/if}
		
		<Button onclick={onAddEventClick}>
			Add New Event
		</Button>
	</div>
</AnalysisContextMenu>