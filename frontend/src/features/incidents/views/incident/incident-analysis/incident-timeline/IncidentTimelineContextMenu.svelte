<script lang="ts">
	import { ElementSize } from "runed";
	import { Button } from "svelte-ux";
	import { useEventDialog } from "./event-dialog/dialogState.svelte";
	import { ZonedDateTime } from "@internationalized/date";
	import type { TimelineItem } from "vis-timeline";
	import { useIncidentTimeline } from "./timelineState.svelte";
	import { mdiPencilCircle, mdiPlusCircle } from "@mdi/js";
	import Header from "$src/components/header/Header.svelte";

	type Props = {
		containerRect: DOMRect;
		item?: TimelineItem;
		clickPos: {x: number; y: number};
		timestamp: ZonedDateTime;
		close: () => void;
	};
	const {containerRect, item, clickPos, timestamp, close}: Props = $props();

	const timelineState = useIncidentTimeline();
	const eventDialog = useEventDialog();

	let ref = $state<HTMLElement>(null!);
	const size = new ElementSize(() => ref);

	const xOverflows = $derived((clickPos.x + size.width) > containerRect.right);
	const naiveX = $derived(Math.round(clickPos.x - containerRect.x));
	const posX = $derived(xOverflows ? (naiveX - size.width) : naiveX);

	const yOverflows = $derived((clickPos.y + size.height) > containerRect.bottom);
	const naiveY = $derived(Math.round(clickPos.y - containerRect.y));
	const posY = $derived(yOverflows ? (naiveY - size.height) : naiveY);

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

<div 
	id="timeline-ctx-container"
	class="w-48 border bg-surface-200 flex flex-col h-fit"
	style="position: absolute; left: {posX}px; top: {posY}px"
	role="presentation"
	onclick={onClicked}
	bind:this={ref}
>
	<Header title="Timeline Actions" classes={{root: "px-2 py-1"}} />

	{#if event}
		<Button variant="fill-light" icon={mdiPencilCircle} rounded={false} classes={{root: "w-full gap-2"}} on:click={onEditEventClick}>
			Edit Event
		</Button>
	{/if}
	
	<Button variant="fill-light" icon={mdiPlusCircle} rounded={false} classes={{root: "w-full gap-2"}} on:click={onAddEventClick}>
		Add New Event
	</Button>
</div>