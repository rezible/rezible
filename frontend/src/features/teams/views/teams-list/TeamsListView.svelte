<script lang="ts">
	import { mdiChevronRight, mdiFilter, mdiMagnify } from "@mdi/js";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, ListItem, TextField, Header, Collapse } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import UserTeamsList from "./UserTeamsList.svelte";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";

	type QueryParams = ListTeamsData["query"];
	let params = $state<QueryParams>({});
	const query = createQuery(() => listTeamsOptions({ query: params }));

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
</script>

<SplitPage nav={userTeamsNav}>
	<div class="">
		<Header title="All Teams" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />
		<Collapse>
			<div slot="trigger" class="flex-1">
				<Button icon={mdiFilter} iconOnly />
				Filters
			</div>
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
		</Collapse>
	</div>

	<LoadingQueryWrapper {query}>
		{#snippet view(teams: Team[])}
			<div class="min-h-0 flex flex-col gap-2 overflow-y-auto flex-0">
				{#each teams as team}
					<a href="/teams/{team.attributes.slug}">
						<ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
							<svelte:fragment slot="avatar">
								<Avatar kind="team" size={32} id={team.id} />
							</svelte:fragment>
							<div slot="actions">
								<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
							</div>
						</ListItem>
					</a>
				{/each}
			</div>

			<!--Pagination
				{pagination}
				hideSinglePage
				perPageOptions={[10, 25, 50]}
				show={['perPage', 'pagination', 'prevPage', 'nextPage']}
				classes={{ perPage: 'flex-1 text-right', pagination: 'px-8' }}
			/-->
		{/snippet}
	</LoadingQueryWrapper>
</SplitPage>

{#snippet userTeamsNav()}
	<Header title="Your Teams" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />

	<UserTeamsList />
{/snippet}
