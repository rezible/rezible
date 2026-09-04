<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import ListFilters from "./ListFilters.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import LoadingIndicator from "$src/components/layout/loading-indicator/LoadingIndicator.svelte";
	import { initEventsListController } from "./controller.svelte";
	import EventListRow from "./EventListRow.svelte";

	const controller = initEventsListController();

	registerPageDescriptor(() => ({ title: "Events" }));
</script>

<FilterPage>
	{#snippet filters()}
		<ListFilters />
	{/snippet}
	<PaginatedQueryListBox {...controller.paginatedEventsQuery}>
		{#if controller.isLoading}
			<LoadingIndicator />
		{:else}
			{#each controller.events as event (event.id)}
				<EventListRow {event} />
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-muted-foreground">No events</span>
				</div>
			{/each}
		{/if}
	</PaginatedQueryListBox>
</FilterPage>
