<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import { EventsListViewState, eventsListViewStateCtx } from "./viewState.svelte";
	import ListFilters from "./ListFilters.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { AnnotationDialogState, setAnnotationDialogState } from "$components/oncall-events/annotation-dialog/dialogState.svelte";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";

	const viewState = new EventsListViewState();
	eventsListViewStateCtx.set(viewState);

	appShell.setPageBreadcrumbs(() => [{ label: "Events" }]);

	setAnnotationDialogState(new AnnotationDialogState({}));
</script>

{#snippet filters()}
	<ListFilters />
{/snippet}

<EventAnnotationDialog />

<FilterPage {filters}>
	<PaginatedListBox pagination={viewState.paginator.pagination}>
		{#if viewState.loading}
			<LoadingIndicator />
		{:else}
			{#each viewState.events as event (event.id)}
				<EventRow {event} />
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-surface-content/80">No Events</span>
				</div>
			{/each}
		{/if}
	</PaginatedListBox>
</FilterPage>
