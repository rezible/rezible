<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import Filters from "$features/events/components/event-filters/EventFilters.svelte";
	import { Button } from "$components/ui/button";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initEventsListController } from "./controller.svelte";
	import EventListRow from "./EventListRow.svelte";

	const controller = initEventsListController();
	registerPageDescriptor(() => ({ title: "Events" }));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
	<div class="flex flex-wrap items-end justify-between gap-2">
		<Filters filters={controller.filters} onchange={controller.setFilters} />
		<Button variant="ghost" size="sm" onclick={controller.resetFilters}>Reset filters</Button>
	</div>
	<LoadingQueryWrapper query={controller.query} feedbackOnly />
	<PaginatedQueryListBox {...controller.paginatedEventsQuery}>
		{#if controller.query.data}
			{#each controller.query.data.data as event (event.id)}
				<EventListRow {event} />
			{:else}
				<p class="p-6 text-sm text-muted-foreground">No matching events.</p>
			{/each}
		{/if}
	</PaginatedQueryListBox>
</div>
