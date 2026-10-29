<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initOncallRostersListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Oncall Rosters" }));

	const controller = initOncallRostersListController();
</script>

{#snippet filters()}
	<SearchInput value={controller.search} oninput={controller.setSearch} />
{/snippet}

<FilterPage {filters} title="Rosters" count={controller.total} updating={controller.query.isFetching}>
	<PaginatedQueryListBox {...controller.paginatedRostersQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(rosters) => rosters.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading rosters" />
			{/snippet}
			{#snippet empty()}
				<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No rosters found.</p>
			{/snippet}
			{#snippet view(rosters)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each rosters as roster (roster.id)}
						<ListRow
							title={roster.attributes.name}
							href={resolve("/(hidden)/oncall/rosters/[slug]", {
								slug: roster.attributes.slug,
							})}
						/>
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
