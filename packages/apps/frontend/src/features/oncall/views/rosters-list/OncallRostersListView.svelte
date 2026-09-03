<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import type { OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import RosterCard from "$features/oncall/components/roster-card/RosterCard.svelte";
	import { initOncallRostersListController } from "./controller.svelte";

	setPageBreadcrumbs(() => [{ label: "Oncall Rosters", path: "/oncall/rosters" }]);

	const controller = initOncallRostersListController();
</script>

{#snippet filters()}
	<SearchInput bind:value={controller.search} />
{/snippet}

<FilterPage {filters}>
	<PaginatedQueryListBox {...controller.paginatedRostersQuery}>
		<LoadingQueryWrapper query={controller.query}>
			{#snippet view(rosters: OncallRoster[])}
				{#each rosters as roster (roster.id)}
					<RosterCard {roster} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Rosters Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
