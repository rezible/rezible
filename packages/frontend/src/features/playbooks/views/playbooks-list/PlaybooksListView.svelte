<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listPlaybooksOptions, type ListPlaybooksData, type Playbook } from "$lib/api";
	import { appShell } from "$features/app";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

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
		<span>{pb.attributes.title}</span>
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(playbooks: Playbook[])}
				{#each playbooks as pb (pb.id)}
					{@render playbookListItem(pb)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Playbooks Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
