<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Month } from "svelte-ux";
	import { listMeetingSessionsOptions, type ListMeetingSessionsData, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import SectionHeader from "$components/section-header/SectionHeader.svelte";

	let queryParams = $state<ListMeetingSessionsData["query"]>({});
	const query = createQuery(() => listMeetingSessionsOptions({ query: queryParams }));

	let monthStart = $state<Date>();
</script>

<div class="flex flex-col min-h-0 h-full">
	<SectionHeader title="Upcoming" />

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
					{:else}
						<span>No upcoming meetings</span>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	</div>
</div>
