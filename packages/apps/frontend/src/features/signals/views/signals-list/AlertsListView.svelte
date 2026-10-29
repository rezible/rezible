<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Button } from "$components/ui/button";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterOptions from "$components/forms/filter-options/FilterOptions.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initSignalsListViewController } from "./controller.svelte";
	import SignalRow from "./SignalRow.svelte";

	const controller = initSignalsListViewController();
	registerPageDescriptor(() => ({ title: "Signals" }));
</script>

{#snippet filters()}
	<SearchInput value={controller.filters.search} oninput={controller.setSearch} />
	<FilterOptions
		label="Attention"
		options={controller.attentionOptions}
		value={controller.filters.attention}
		onchange={controller.setAttention}
	/>
{/snippet}

<FilterPage
	{filters}
	title="Signals"
	count={controller.total}
	updating={controller.query.isFetching}
	activeFilters={controller.activeFilterCount}
	onclear={controller.resetFilters}
>
	<PaginatedQueryListBox {...controller.paginatedAlertsQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(definitions) => definitions.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading signals" />
			{/snippet}
			{#snippet empty()}
				<div
					class="flex flex-col items-start gap-2 rounded-lg border bg-card p-6 text-sm text-muted-foreground"
				>
					{#if controller.activeFilterCount > 0}
						<p>No signals match these filters.</p>
						<Button variant="link" size="inline" onclick={controller.resetFilters}>
							Clear filters
						</Button>
					{:else}
						<p>No signals yet.</p>
					{/if}
				</div>
			{/snippet}
			{#snippet view(definitions)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each definitions as definition (definition.id)}
						<SignalRow {definition} />
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
