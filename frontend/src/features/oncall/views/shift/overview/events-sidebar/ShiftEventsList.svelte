<script lang="ts">
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import { useShiftViewState } from "../../shiftViewState.svelte";

	const viewState = useShiftViewState();

	const shift = $derived(viewState.shift);
	const events = $derived(viewState.filteredEvents);
	const roster = $derived(shift?.attributes.roster);
</script>

{#if shift && events && roster}
	{#each viewState.filteredEvents as event}
		<EventRow {event} />
	{:else}
		<span class="w-full text-center py-8">No Events</span>
	{/each}
{/if}

<EventAnnotationDialog />