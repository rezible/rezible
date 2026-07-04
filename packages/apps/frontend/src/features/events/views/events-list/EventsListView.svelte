<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import ListFilters from "./ListFilters.svelte";
	import PaginatedListBox from "$src/components/layout/paginated-listbox/PaginatedListBox.svelte";
	import LoadingIndicator from "$src/components/layout/loading-indicator/LoadingIndicator.svelte";
	import { initEventsListController } from "./controller.svelte";
	import EventListRow from "./EventListRow.svelte";

	const controller = initEventsListController();

	setPageBreadcrumbs(() => [{ label: "Events" }]);
</script>

<FilterPage>
	{#snippet filters()}
		<ListFilters />
	{/snippet}
	<PaginatedListBox>
		{#if controller.query.isLoading}
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
	</PaginatedListBox>
</FilterPage>
