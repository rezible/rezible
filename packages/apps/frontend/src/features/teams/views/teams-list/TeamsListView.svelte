<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initTeamsListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Teams" }));

	const controller = initTeamsListController();
</script>

{#snippet filters()}
	<SearchInput value={controller.search} oninput={controller.setSearch} />
{/snippet}

<FilterPage {filters} title="Teams" count={controller.total} updating={controller.query.isFetching}>
	<PaginatedQueryListBox {...controller.paginatedTeamsQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(teams) => teams.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading teams" />
			{/snippet}
			{#snippet empty()}
				<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No teams found.</p>
			{/snippet}
			{#snippet view(teams)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each teams as team (team.id)}
						<ListRow
							title={team.attributes.name}
							href={resolve("/(hidden)/teams/[slug]/[[view=teamView]]", {
								slug: team.attributes.slug,
							})}
						/>
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
