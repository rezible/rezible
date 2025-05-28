<script lang="ts">
	import { mdiFilter } from "@mdi/js";
	import { Header, Button } from "svelte-ux";
	import type { OncallEvent } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import { useShiftViewState } from "../../shiftViewState.svelte";

	const viewState = useShiftViewState();

	const shift = $derived(viewState.shift);
	const events = $derived(viewState.filteredEvents);
	const shiftRoster = $derived(shift?.attributes.roster);

	let showFilters = $state(false);

	let annoEvent = $state<OncallEvent>();
	const annotationRosterIds = $derived(shiftRoster ? [shiftRoster.id] : []);
	const currentAnnotation = $derived(annoEvent?.attributes.annotations.find(a => a.attributes.roster.id === shiftRoster?.id));
</script>

{#if shift && events && shiftRoster}
	{#each events as event}
		<EventRowItem 
			{event}
			{annotationRosterIds}
			annotation={event.attributes.annotations?.at(0)}
			editAnnotation={() => (annoEvent = event)}
		/>
	{/each}
{/if}

{#if shiftRoster}
	<EventAnnotationDialog 
		event={annoEvent}
		current={currentAnnotation}
		roster={shiftRoster}
		onClose={() => (annoEvent = undefined)}
	/>
{/if}