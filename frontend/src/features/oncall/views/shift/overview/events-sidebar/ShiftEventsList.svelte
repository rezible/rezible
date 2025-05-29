<script lang="ts">
	import type { OncallEvent } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import { useShiftViewState } from "../../shiftViewState.svelte";

	const viewState = useShiftViewState();

	const shift = $derived(viewState.shift);
	const events = $derived(viewState.filteredEvents);
	const shiftRoster = $derived(shift?.attributes.roster);

	let showFilters = $state(false);

	let annoDialogEvent = $state<OncallEvent>();
	const annoDialogEventAnnotation = $derived(annoDialogEvent && viewState.eventAnnotationsMap.get(annoDialogEvent.id));
	const annotatableRosterIds = $derived(shiftRoster ? [shiftRoster.id] : []);

	const onEditAnnotation = (event: OncallEvent) => {
		annoDialogEvent = event;

	}
</script>

{#if shift && events && shiftRoster}
	{#each events as event}
		{@const annotation = viewState.eventAnnotationsMap.get(event.id)}
		<EventRowItem 
			{event}
			{annotatableRosterIds}
			annotations={annotation ? [annotation] : []}
			editAnnotation={() => onEditAnnotation(event)}
		/>
	{/each}
{/if}

{#if shiftRoster}
	<EventAnnotationDialog 
		event={annoDialogEvent}
		current={annoDialogEventAnnotation}
		roster={shiftRoster}
		onClose={() => (annoDialogEvent = undefined)}
	/>
{/if}