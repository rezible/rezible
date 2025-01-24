<script lang="ts">
	import { Button, Dialog } from "svelte-ux";
	import ConfirmButtons from '$components/confirm-buttons/ConfirmButtons.svelte';
    import { timeline } from "$features/incidents/components/retrospective/incident-analysis/incident-timeline/timeline.svelte";
    import EventEditor from "./EventEditor.svelte";

	type Props = {
		creatingEvent: boolean;
	}
	let { creatingEvent = $bindable() }: Props = $props();
</script>

<Dialog
	open={creatingEvent}
	on:close={() => {creatingEvent = false}}
	persistent
	portal
	classes={{ root:"p-8", dialog: 'flex flex-col w-full max-w-7xl h-full' }}
>
	<div slot="header" class="border-b p-2">
		<span class="text-xl">Create Event</span>
	</div>

	<div class="flex-1 min-h-0">
		{#if creatingEvent}
			<EventEditor event={timeline.editingEvent} />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			onClose={() => {creatingEvent = false}}
			onConfirm={() => {console.log("confirm")}}
			saveEnabled={false}
		/>
	</svelte:fragment>
</Dialog>