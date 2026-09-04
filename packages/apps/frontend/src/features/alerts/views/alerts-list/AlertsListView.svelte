<script lang="ts">
	import { resolve } from "$app/paths";
	import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
	import { listAlertsOptions, type Alert, type ListAlertsData } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";

	registerPageDescriptor(() => ({ title: "Alerts" }));

	let searchValue = $state<string>();
	const params = $derived<ListAlertsData["query"]>({
		search: searchValue,
	});
	const paginatedAlertsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listAlertsOptions({ query: { ...params, ...pagination } }),
		resetWhen: () => [searchValue],
	});
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

{#snippet alertListItem(a: Alert)}
	<a href={resolve(`/alerts/${a.id}`)}>
		<span>{a.attributes.title}</span>
		<!-- <ListItem title={a.attributes.title} subheading={a.attributes.description} /> -->
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedQueryListBox {...paginatedAlertsQuery}>
		<LoadingQueryWrapper query={paginatedAlertsQuery.query}>
			{#snippet view(alerts: Alert[])}
				{#each alerts as a (a.id)}
					{@render alertListItem(a)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Alerts Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
