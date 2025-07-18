<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listPlaybooksOptions, listSystemComponentsOptions, type ListPlaybooksData, type ListSystemComponentsData, type SystemComponent } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import { ListItem } from "svelte-ux";

	appShell.setPageBreadcrumbs(() => [{ label: "System Components" }]);

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
	<a href="/components/{c.id}">
		<ListItem title={c.attributes.name} subheading={c.attributes.description} />
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(components: SystemComponent[])}
				{#each components as c (c.id)}
					{@render componentListItem(c)}
				{:else}
					<span>No Components Found</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
