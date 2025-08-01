<script lang="ts">
	import { Month } from "svelte-ux";
	import { type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meeting-session/components/meeting-session-card/MeetingSessionCard.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import MeetingsPageActions from "./MeetingsListPageActions.svelte";
	import { useMeetingsListViewState } from "../lib/viewState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
	appShell.setPageActions(MeetingsPageActions, true);

	const viewState = useMeetingsListViewState();
</script>

{#snippet filters()}
	<SearchInput bind:value={viewState.searchValue} />

	<div class="pb-2 border">
		<Month bind:startOfMonth={viewState.monthStart} showOutsideDays />
	</div>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={viewState.pagination}>
		<LoadingQueryWrapper query={viewState.query}>
			{#snippet view(sessions: MeetingSession[])}
				{#each sessions as session}
					<MeetingSessionCard {session} />
				{:else}
					<span>No upcoming meetings</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
