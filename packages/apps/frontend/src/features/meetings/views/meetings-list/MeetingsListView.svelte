<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import MeetingsPageActions from "./MeetingsListPageActions.svelte";
	import { initMeetingsListViewController } from "./controller.svelte";

	const controller = initMeetingsListViewController();

	registerPageDescriptor(() => ({
		title: "Meetings",
		pageActions: actions,
	}));
</script>

{#snippet actions()}
	<MeetingsPageActions />
{/snippet}

<FilterPage title="Meetings" count={controller.total} updating={controller.query.isFetching}>
	<LoadingQueryWrapper
		query={controller.query}
		feedback="quiet"
		isEmpty={(sessions) => sessions.length === 0}
	>
		{#snippet loading()}
			<ListSkeleton label="Loading meetings" />
		{/snippet}
		{#snippet empty()}
			<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No upcoming meetings.</p>
		{/snippet}
		{#snippet view(sessions)}
			<ul class="divide-y overflow-hidden rounded-lg border bg-card">
				{#each sessions as session (session.id)}
					{#snippet startsAt()}
						<Timestamp value={session.attributes.startsAt} />
					{/snippet}
					<ListRow
						title={session.attributes.title}
						href={resolve("/(hidden)/meetings/sessions/[id]", { id: session.id })}
						aside={startsAt}
					/>
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</FilterPage>
