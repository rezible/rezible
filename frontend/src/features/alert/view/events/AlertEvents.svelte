<script lang="ts">
	import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/events/annotation-dialog/dialogState.svelte";
	import EventRow from "$src/components/events/EventRow.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import AlertEventsFilters from "./AlertEventsFilters.svelte";
	import { Pagination } from "svelte-ux";
	import { AlertEventsState } from "./alertEventsState.svelte";
	import type { Event } from "$src/lib/api";

	const eventsState = new AlertEventsState();

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

<div class="w-full h-full flex flex-col gap-2">
	<AlertEventsFilters bind:rosterId={eventsState.rosterId} bind:dateRange={eventsState.dateRange} />

	<div class="flex-1 flex flex-col gap-1 border">
		<LoadingQueryWrapper query={eventsState.query}>
			{#snippet view(events: Event[])}
				{#each events as ev}
					<EventRow event={ev} />
				{:else}
					<div class="p-2">
						<span>No results</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>

	<Pagination {...eventsState.paginator.paginationProps} />
</div>
