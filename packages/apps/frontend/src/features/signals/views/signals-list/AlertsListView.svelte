<script lang="ts">
	import { resolve } from "$app/paths";
	import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
	import { listAlertsOptions, type AlertDefinition, type ListAlertsData } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import { attentionStatus } from "$features/signals/lib/attention";

	registerPageDescriptor(() => ({ title: "Signals" }));

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

{#snippet alertListItem(a: AlertDefinition)}
	{@const attention = attentionStatus(a.attributes.situationSignalAttention)}
	<a href={resolve(`/signals/${a.id}`)} class="flex flex-wrap items-center gap-2">
		<span>{a.attributes.title}</span>
		{#if attention}
			<StatusBadge status={attention} variant="inline" />
		{/if}
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedQueryListBox {...paginatedAlertsQuery}>
		<LoadingQueryWrapper query={paginatedAlertsQuery.query}>
			{#snippet view(alerts: AlertDefinition[])}
				{#each alerts as a (a.id)}
					{@render alertListItem(a)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-muted-foreground">No signals found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
