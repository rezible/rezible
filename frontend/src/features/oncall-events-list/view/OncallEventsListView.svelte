<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import ListFilters from "./ListFilters.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import { AnnotationDialogState, setAnnotationDialogState } from "$components/oncall-events/annotation-dialog/dialogState.svelte";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import { createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listOncallEventsOptions } from "$lib/api";
	import { EventsListFiltersState } from "./filters.svelte";

	const filtersState = new EventsListFiltersState();

	const paginator = new QueryPaginatorState();
	const queryOptions = $derived(listOncallEventsOptions({ 
		query: {
			...filtersState.queryData,
			limit: paginator.limit,
			offset: paginator.offset,
			withAnnotations: true,
		}
	}));
	const query = createQuery(() => ({
		...queryOptions,
		enabled: filtersState.queryEnabled,
	}));
	const events = $derived(query.data?.data ?? []);

	const queryClient = useQueryClient();
	const invalidateQuery = () => {
		queryClient.invalidateQueries(queryOptions);
	}

	appShell.setPageBreadcrumbs(() => [{ label: "Events" }]);

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

{#snippet filters()}
	<ListFilters {filtersState} />
{/snippet}

<EventAnnotationDialog />

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		{#if query.isLoading}
			<LoadingIndicator />
		{:else}
			{#each events as event (event.id)}
				<EventRow {event} />
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-surface-content/80">No Events</span>
				</div>
			{/each}
		{/if}
	</PaginatedListBox>
</FilterPage>
