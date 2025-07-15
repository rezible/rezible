<script lang="ts">
	import { mdiChevronRight, mdiFilter, mdiMagnify } from "@mdi/js";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, ListItem, TextField, Collapse } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";
	import Header from "$src/components/header/Header.svelte";
	import Icon from "$src/components/icon/Icon.svelte";
	import SectionHeader from "$src/components/section-header/SectionHeader.svelte";

	type QueryParams = ListTeamsData["query"];
	let params = $state<QueryParams>({});
	const query = createQuery(() => listTeamsOptions({ query: params }));

	let userParams = $state<ListTeamsData["query"]>({});
	const userTeamsQuery = createQuery(() => listTeamsOptions({ query: userParams }));

	const maybeUpdateParams = (newParams: QueryParams) => {
		let { limit, offset, search } = params || {};
		offset = offset ?? 0;

		const newLimit = newParams?.limit ?? limit;
		const newOffset = newParams?.offset ?? offset;
		const newSearch = newParams?.search ?? search;

		if (limit !== newLimit || offset !== newOffset || search !== newSearch) {
			// console.log('updateParams', { limit: newLimit, offset: newOffset, search: newSearch });
		}
	};

	/*
	const pagination = paginationStore({ total: 0 });
	$effect(() => {
		const total = query.data?.pagination.total ?? 0;
		if ($pagination.total !== total) pagination.setTotal(total);
	});

	$effect(() => {
		const unsubscribe = pagination.subscribe(({ perPage, page }) => {
			if (page === 0) return;
			maybeUpdateParams({ limit: perPage, offset: perPage * (page - 1)});
		});
		return unsubscribe;
	})
	*/

	let showFilters = $state(false);
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

<SplitPage>
	{#snippet nav()}
		<SectionHeader title="Your Teams" />

		<div class="flex flex-col h-full">
			<LoadingQueryWrapper query={userTeamsQuery}>
				{#snippet view(userTeams: Team[])}
					{#each userTeams as team}
						{@render teamCard(team)}
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	{/snippet}

	<div class="">
		<SectionHeader title="All Teams">
			{#snippet filters()}
				<div class="p-3 border-t">
					<TextField
						dense
						rounded
						label="Search For Teams"
						labelPlacement="float"
						icon={mdiMagnify}
						debounceChange={500}
						on:change={({ detail }) =>
							maybeUpdateParams({
								search: detail.value ? String(detail.value) : undefined,
							})}
					/>
				</div>
			{/snippet}
		</SectionHeader>
	</div>

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
</SplitPage>
