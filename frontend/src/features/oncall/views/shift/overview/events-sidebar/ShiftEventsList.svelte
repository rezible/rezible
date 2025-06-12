<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import { AnnotationDialogState, setAnnotationDialogState } from "$components/oncall-events/annotation-dialog/dialogState.svelte";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import { useShiftViewState } from "../../shiftViewState.svelte";

	const viewState = useShiftViewState();

	const shift = $derived(viewState.shift);
	const events = $derived(viewState.filteredEvents);
	const roster = $derived(shift?.attributes.roster);

	const annoDialog = new AnnotationDialogState({
		onClosed: (updated?: OncallAnnotation) => {
			
		}
	});
	setAnnotationDialogState(annoDialog);

	const listEvents = $derived.by<[OncallEvent, OncallAnnotation[]][]>(() => {
		return viewState.filteredEvents.map(event => {
			const anno = viewState.eventAnnotationsMap.get(event.id);
			return [event, !!anno ? [anno] : []]
		})
	});
</script>

{#if shift && events && roster}
	{#each listEvents as [event, annotations]}
		<EventRow {event} {annotations} />
	{/each}
{/if}

<EventAnnotationDialog />