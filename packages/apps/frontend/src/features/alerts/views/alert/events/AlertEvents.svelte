<script lang="ts">
	import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/common/events/annotation-dialog/dialogState.svelte";
	import EventRow from "$src/components/common/events/EventRow.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import AlertEventsFilters from "./AlertEventsFilters.svelte";
	import { AlertEventsViewController } from "./alertEventsViewController.svelte";
	import type { Event } from "$lib/api";

	const events = new AlertEventsViewController();

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

<div class="w-full h-full flex flex-col gap-2">
	<AlertEventsFilters bind:rosterId={events.rosterId} />

	<div class="flex-1 flex flex-col gap-1 border">
		<LoadingQueryWrapper query={events.query}>
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

	<!-- <Pagination {...eventsState.paginator.paginationProps} /> -->
</div>
