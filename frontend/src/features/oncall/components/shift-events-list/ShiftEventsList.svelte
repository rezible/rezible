<script lang="ts">
	import { mdiFilter } from "@mdi/js";
	import { Header, Button } from "svelte-ux";
	import type { OncallEvent, OncallShift } from "$lib/api";
	import EventAnnotationDialog from "$components/event-annotation-dialog/EventAnnotationDialog.svelte";
	import EventRowItem from "$components/oncall-events-table/EventRowItem.svelte";

	type Props = {
		shift: OncallShift;
		events: OncallEvent[];
	};
	const { shift, events }: Props = $props();

	let showFilters = $state(false);

	const rosterIds = $derived([shift.attributes.roster.id]);

	let annotationEvent = $state<OncallEvent>();
	const currentAnnotation = $derived(annotationEvent?.attributes.annotations.find(a => a.attributes.rosterId === shift.attributes.roster.id));
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
		{#each events as event}
			<EventRowItem 
				{event}
				annotationRosterIds={rosterIds}
				annotation={event.attributes.annotations?.at(0)}
				editAnnotation={() => (annotationEvent = event)}
			/>
		{/each}
	</div>
</div>

<EventAnnotationDialog event={annotationEvent} current={currentAnnotation} onClose={() => (annotationEvent = undefined)} />