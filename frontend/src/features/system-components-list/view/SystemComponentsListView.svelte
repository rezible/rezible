<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listSystemComponentsOptions, type ListSystemComponentsData, type SystemComponent } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

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
		<span>list components</span>
		<!-- <ListItem title={c.attributes.name} subheading={c.attributes.description} /> -->
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(components: SystemComponent[])}
				{#each components as c (c.id)}
					{@render componentListItem(c)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Components Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
