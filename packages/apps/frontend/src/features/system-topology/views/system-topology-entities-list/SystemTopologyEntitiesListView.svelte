<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listSystemTopologyEntitiesOptions, type ListSystemTopologyEntitiesData, type SystemTopologyEntity } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SystemTopologyEntityListItem from "./SystemTopologyEntityListItem.svelte";

	setPageBreadcrumbs(() => [{ label: "Topology" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListSystemTopologyEntitiesData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listSystemTopologyEntitiesOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

{#snippet entityListItem(c: SystemTopologyEntity)}
	<SystemTopologyEntityListItem entity={c} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(entities: SystemTopologyEntity[])}
				{#each entities as c (c.id)}
					{@render entityListItem(c)}
				{:else}
					<div class="grid flex-1 place-items-center rounded border border-dashed border-border p-8">
						<span class="text-sm text-muted-foreground">No topology entities found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
