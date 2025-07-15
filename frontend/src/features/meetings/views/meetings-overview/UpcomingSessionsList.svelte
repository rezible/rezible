<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Month } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronDown, mdiFilter } from "@mdi/js";
	import { listMeetingSessionsOptions, type ListMeetingSessionsData, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import Header from "$components/header/Header.svelte";

	let queryParams = $state<ListMeetingSessionsData["query"]>({});
	const query = createQuery(() => listMeetingSessionsOptions({ query: queryParams }));

	let monthStart = $state<Date>();
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Header title="Upcoming" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			<Button icon={mdiFilter} iconOnly />
		{/snippet}
	</Header>

	<div class="grid grid-cols-4 h-full gap-2">
		<div class="h-full flex flex-col gap-2">
			<div class="pb-2 border">
				<Month bind:startOfMonth={monthStart} showOutsideDays />
			</div>
		</div>

		<div class="col-span-3 flex flex-col gap-2 overflow-y-auto p-2">
			<LoadingQueryWrapper {query}>
				{#snippet view(sessions: MeetingSession[])}
					{#each sessions as session}
						<MeetingSessionCard {session} />
					{/each}
					{#if !sessions || sessions.length === 0}
						<span>No upcoming sessions</span>
					{/if}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	</div>
</div>
