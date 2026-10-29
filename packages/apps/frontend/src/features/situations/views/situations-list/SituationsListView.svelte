<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initSituationsListController } from "./controller.svelte";
	import Filters from "$features/situations/components/situation-filters/SituationFilters.svelte";
	import Row from "$features/situations/components/situation-row/SituationRow.svelte";
	import { Button } from "$components/ui/button";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const controller = initSituationsListController();

	registerPageDescriptor(() => ({ title: "Situations" }));
</script>

{#snippet filters()}
	<Filters
		filters={controller.filters}
		onchange={controller.setFilters}
		statusOptions={controller.statusOptions}
	/>
{/snippet}

<FilterPage
	{filters}
	title={`${controller.activeStatus.label} situations`}
	count={controller.total}
	updating={controller.query.isFetching}
	activeFilters={controller.activeFilterCount}
	onclear={controller.clearSearch}
>
	<PaginatedQueryListBox {...controller.paginatedSituationsQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(situations) => situations.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading situations" />
			{/snippet}
			{#snippet empty()}
				<div
					class="flex flex-col items-start gap-2 rounded-lg border bg-card p-6 text-sm text-muted-foreground"
				>
					{#if controller.activeFilterCount > 0}
						<p>No {controller.activeStatus.label.toLowerCase()} situations match this search.</p>
						<Button variant="link" size="inline" onclick={controller.clearSearch}>
							Clear search
						</Button>
					{:else}
						<p>{controller.activeStatus.emptyMessage}</p>
					{/if}
				</div>
			{/snippet}
			{#snippet view(situations)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each situations as situation (situation.id)}
						<Row {situation} now={controller.now} />
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
