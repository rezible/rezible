<script lang="ts">
	import { mdiFilter } from "@mdi/js";
	import { Header, Button } from "svelte-ux";
	import type { OncallEvent } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import { shiftViewStateCtx } from "../context.svelte";

	const viewState = shiftViewStateCtx.get();

	const shift = $derived(viewState.shift);
	const events = $derived(viewState.filteredEvents);
	const shiftRoster = $derived(shift?.attributes.roster);

	let showFilters = $state(false);

	let annoEvent = $state<OncallEvent>();
	const annotationRosterIds = $derived(shiftRoster ? [shiftRoster.id] : []);
	const currentAnnotation = $derived(annoEvent?.attributes.annotations.find(a => a.attributes.roster.id === shiftRoster?.id));
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit pt-2 flex flex-col gap-2" class:pb-2={!showFilters}>
		<Header title="Shift Events" subheading="Showing All" class="px-2">
			<svelte:fragment slot="actions">
				<Button icon={mdiFilter} iconOnly on:click={() => (showFilters = !showFilters)} />
			</svelte:fragment>
		</Header>

		{#if showFilters}
			<div class="w-full h-12 border"></div>
		{/if}
	</div>
	<div class="flex-1 flex flex-col gap-1 px-0 overflow-y-auto">
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
	</div>
</div>

{#if shiftRoster}
	<EventAnnotationDialog 
		event={annoEvent}
		current={currentAnnotation}
		roster={shiftRoster}
		onClose={() => (annoEvent = undefined)}
	/>
{/if}