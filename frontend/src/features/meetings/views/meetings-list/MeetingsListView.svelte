<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Month } from "svelte-ux";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { listMeetingSessionsOptions, type ListMeetingSessionsData, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import MeetingsPageActions from "$features/meetings/components/meetings-page-actions/MeetingsPageActions.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";

	appShell.setPageActions(MeetingsPageActions, true);

	const pagination = createPaginationStore();
	let searchValue = $state<string>();

	let queryParams = $state<ListMeetingSessionsData["query"]>({});
	const query = createQuery(() => listMeetingSessionsOptions({ query: queryParams }));

	let monthStart = $state<Date>();
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />

	<div class="pb-2 border">
		<Month bind:startOfMonth={monthStart} showOutsideDays />
	</div>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper {query}>
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
