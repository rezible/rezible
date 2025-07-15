<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button } from "svelte-ux";
	import { listMeetingSessionsOptions, type MeetingSchedule, type MeetingSession } from "$lib/api";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Header from "$src/components/header/Header.svelte";

	type Props = {
		schedule: MeetingSchedule;
	};
	const { schedule }: Props = $props();

	const queryOptions = () =>
		listMeetingSessionsOptions({
			query: { meetingScheduleId: schedule.id },
		});
	const query = createQuery(queryOptions);
</script>

<div class="border p-2">
	<Header title="Next Session" classes={{ title: "text-lg" }} />
	<div class="h-20">
		<!--MeetingSessionCard /-->
	</div>
</div>

<div class="border p-2 flex-1 min-h-0 overflow-y-auto">
	<Header title="Past Sessions" classes={{ title: "text-lg" }}>
		
	</Header>
	<LoadingQueryWrapper {query}>
		{#snippet view(sessions: MeetingSession[])}
			{#each sessions as session}
				<MeetingSessionCard {session} />
			{/each}
		{/snippet}
	</LoadingQueryWrapper>
</div>
