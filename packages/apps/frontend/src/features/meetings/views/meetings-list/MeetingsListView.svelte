<script lang="ts">
	import { type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import { useAppShell } from "$lib/appShell.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import MeetingsPageActions from "./MeetingsListPageActions.svelte";
	import { initMeetingsListViewController } from "./controller.svelte";

	const view = initMeetingsListViewController();
	const query = $derived(view.query);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
	appShell.setPageActions(MeetingsPageActions, true);
</script>

{#snippet filters()}
	<SearchInput bind:value={view.searchValue} />

	<div class="pb-2 border">
		<span>month</span>
		<!-- <Month bind:startOfMonth={viewState.monthStart} showOutsideDays /> -->
	</div>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(sessions: MeetingSession[])}
				{#each sessions as session}
					<MeetingSessionCard {session} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Upcoming Meetings Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
