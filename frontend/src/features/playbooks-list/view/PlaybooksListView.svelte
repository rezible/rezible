<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listPlaybooksOptions, type ListPlaybooksData, type Playbook } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { ListItem } from "svelte-ux";

	appShell.setPageBreadcrumbs(() => [{ label: "Playbooks" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListPlaybooksData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listPlaybooksOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
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
				{#each playbooks as pb (pb.id)}
					{@render playbookListItem(pb)}
				{:else}
					<span>No Playbooks Found</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
