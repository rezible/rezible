<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listAlertsOptions, type Alert, type ListAlertsData } from "$lib/api";
	import { appShell } from "$features/app";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Alerts" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListAlertsData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listAlertsOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

{#snippet alertListItem(a: Alert)}
	<a href="/alerts/{a.id}">
		<span>{a.attributes.title}</span>
		<!-- <ListItem title={a.attributes.title} subheading={a.attributes.description} /> -->
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
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
	</PaginatedListBox>
</FilterPage>
