<script lang="ts">
	import { useAlertViewState } from "$features/alert";
	import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/oncall-events/annotation-dialog/dialogState.svelte";
	import EventRow from "$src/components/oncall-events/EventRow.svelte";
	import PaginatedListBox from "$src/components/paginated-listbox/PaginatedListBox.svelte";
	import { listOncallEventsOptions, type ListOncallEventsData, type OncallEvent } from "$src/lib/api";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import FilterPage from "$src/components/filter-page/FilterPage.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";

	const viewState = useAlertViewState();

	const paginator = new QueryPaginatorState();

	const queryParams = $derived<ListOncallEventsData["query"]>({
		alertId: viewState.alertId,
		...paginator.queryParams,
	});
	const query = createQuery(() => listOncallEventsOptions({ query: queryParams }));
	paginator.watchQuery(query);

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

{#snippet filters()}
	<span>roster</span>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(events: OncallEvent[])}
				{#each events as ev}
					<EventRow event={ev} />
				{:else}
					<span>No results</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>

