<script lang="ts">
	import { mdiChevronRight } from "@mdi/js";
	import { createQuery } from "@tanstack/svelte-query";
	import { ListItem } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$src/components/filter-page/FilterPage.svelte";
	import Icon from "$src/components/icon/Icon.svelte";
	import SearchInput from "$src/components/search-input/SearchInput.svelte";

	let searchValue = $state<string>();
	let params = $state<ListTeamsData["query"]>({});
	const query = createQuery(() => listTeamsOptions({ query: params }));

	// const maybeUpdateParams = (newParams: QueryParams) => {
	// 	let { limit, offset, search } = params || {};
	// 	offset = offset ?? 0;

	// 	const newLimit = newParams?.limit ?? limit;
	// 	const newOffset = newParams?.offset ?? offset;
	// 	const newSearch = newParams?.search ?? search;

	// 	if (limit !== newLimit || offset !== newOffset || search !== newSearch) {
	// 		// console.log('updateParams', { limit: newLimit, offset: newOffset, search: newSearch });
	// 	}
	// };
</script>

{#snippet teamCard(team: Team)}
	<a href="/teams/{team.attributes.slug}">
		<ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
			<svelte:fragment slot="avatar">
				<Avatar kind="team" size={32} id={team.id} />
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
			</div>
		</ListItem>
	</a>
{/snippet}

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<div class="min-h-0 flex flex-col gap-2 overflow-y-auto">
		<LoadingQueryWrapper {query}>
			{#snippet view(teams: Team[])}
				{#each teams as team}
					{@render teamCard(team)}
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
		<!--Pagination
			{pagination}
			hideSinglePage
			perPageOptions={[10, 25, 50]}
			show={['perPage', 'pagination', 'prevPage', 'nextPage']}
			classes={{ perPage: 'flex-1 text-right', pagination: 'px-8' }}
		/-->
	</div>
</FilterPage>
