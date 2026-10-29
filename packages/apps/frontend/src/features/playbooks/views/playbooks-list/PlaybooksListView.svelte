<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initPlaybooksListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Playbooks" }));

	const controller = initPlaybooksListController();
</script>

{#snippet filters()}
	<SearchInput value={controller.search} oninput={controller.setSearch} />
{/snippet}

<FilterPage {filters} title="Playbooks" count={controller.total} updating={controller.query.isFetching}>
	<PaginatedQueryListBox {...controller.paginatedPlaybooksQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(playbooks) => playbooks.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading playbooks" />
			{/snippet}
			{#snippet empty()}
				<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No playbooks found.</p>
			{/snippet}
			{#snippet view(playbooks)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each playbooks as playbook (playbook.id)}
						<ListRow
							title={playbook.attributes.title}
							href={resolve("/(hidden)/playbooks/[id]", { id: playbook.id })}
						/>
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
