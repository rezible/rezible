<script lang="ts">
	import EventRow from "$components/common/events/EventRow.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import AlertEventsFilters from "./AlertEventsFilters.svelte";
	import { AlertEventsViewController } from "./alertEventsViewController.svelte";
	import type { Event } from "$lib/api";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const events = new AlertEventsViewController();
</script>

<div class="w-full h-full flex flex-col gap-2">
	<AlertEventsFilters bind:rosterId={events.rosterId} />

	<div class="flex-1 min-h-0 border p-1">
		<PaginatedQueryListBox {...events.paginatedEventsQuery}>
			<LoadingQueryWrapper query={events.paginatedEventsQuery.query}>
				{#snippet view(events: Event[])}
					{#each events as ev (ev.id)}
						<EventRow event={ev} />
					{:else}
						<div class="p-2">
							<span>No results</span>
						</div>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</PaginatedQueryListBox>
	</div>
</div>
