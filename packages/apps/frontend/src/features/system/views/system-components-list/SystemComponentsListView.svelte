<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listSystemComponentsOptions, type ListSystemComponentsData, type SystemComponent } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SystemComponentListItem from "./SystemComponentListItem.svelte";

	setPageBreadcrumbs(() => [{ label: "System Components" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListSystemComponentsData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listSystemComponentsOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

{#snippet componentListItem(c: SystemComponent)}
	<SystemComponentListItem component={c} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(components: SystemComponent[])}
				{#each components as c (c.id)}
					{@render componentListItem(c)}
				{:else}
					<div class="grid flex-1 place-items-center rounded border border-dashed border-border p-8">
						<span class="text-sm text-muted-foreground">No components found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
