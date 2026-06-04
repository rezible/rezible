<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTeamsOptions, type Team } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedListBox from "$src/components/layout/paginated-listbox/PaginatedListBox.svelte";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

	setPageBreadcrumbs(() => [{ label: "Teams" }]);

	let searchValue = $state<string>();
	const paginator = new QueryPaginatorState();
	const params = $derived(listTeamsOptions({ query: {
		limit: paginator.limit,
		offset: paginator.offset,
		search: searchValue,
	}}));
	const query = createQuery(() => params);
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

{#snippet teamCard(team: Team)}
	<a href="/teams/{team.attributes.slug}">
		<span>team card</span>
		<!--ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
			<svelte:fragment slot="avatar">
				<Avatar kind="team" size={32} id={team.id} />
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
			</div>
		</ListItem-->
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(teams: Team[])}
				{#each teams as team (team.id)}
					{@render teamCard(team)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Teams Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
