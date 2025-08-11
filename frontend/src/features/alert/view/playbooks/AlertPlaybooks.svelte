<script lang="ts">
	import { useAlertViewState } from "$features/alert";
	import PaginatedListBox from "$src/components/paginated-listbox/PaginatedListBox.svelte";
	import { listPlaybooksOptions, type ListPlaybooksData, type Playbook } from "$src/lib/api";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import FilterPage from "$src/components/filter-page/FilterPage.svelte";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import { ListItem } from "svelte-ux";

	const viewState = useAlertViewState();

	const paginator = new QueryPaginatorState();

	const queryParams = $derived<ListPlaybooksData["query"]>({
		alertId: viewState.alertId,
		...paginator.queryParams,
	});
	const query = createQuery(() => listPlaybooksOptions({ query: queryParams }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<span>roster</span>
{/snippet}

{#snippet playbookListItem(pb: Playbook)}
	<a href="/playbooks/{pb.id}">
		<ListItem title={pb.attributes.title} />
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(playbooks: Playbook[])}
				{#each playbooks as pb}
					{@render playbookListItem(pb)}
				{:else}
					<span>No results</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>

