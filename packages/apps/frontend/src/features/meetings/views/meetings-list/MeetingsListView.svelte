<script lang="ts">
	import { type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import { useAppShell } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import MeetingsPageActions from "./MeetingsListPageActions.svelte";
	import { initMeetingsListViewController } from "./controller.svelte";

	const view = initMeetingsListViewController();
	const query = $derived(view.query);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
	appShell.setPageActions(MeetingsPageActions, true);
</script>

{#snippet filters()}
	<div class="pb-2 border">
		<span>month</span>
		<!-- <Month bind:startOfMonth={viewState.monthStart} showOutsideDays /> -->
	</div>
{/snippet}

<FilterPage {filters}>
	<div class="flex min-h-0 flex-col gap-1 overflow-auto">
		<LoadingQueryWrapper {query}>
			{#snippet view(sessions: MeetingSession[])}
				{#each sessions as session (session.id)}
					<MeetingSessionCard {session} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Upcoming Meetings Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</FilterPage>
