<script lang="ts">
	import EventRow from "$src/components/events/EventRow.svelte";
	import EventAnnotationDialog from "$src/components/events/annotation-dialog/EventAnnotationDialog.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	const view = useOncallShiftViewState();

	const shift = $derived(view.shift);
	const events = $derived(view.filteredEvents);
	const roster = $derived(shift?.attributes.roster);
</script>

{#if shift && events && roster}
	{#each view.filteredEvents as event}
		<EventRow {event} />
	{:else}
		<span class="w-full text-center py-8">No Events</span>
	{/each}
{/if}

<EventAnnotationDialog />