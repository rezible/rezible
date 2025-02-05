<script lang="ts">
	import {
		listMeetingSchedulesOptions,
		type ListMeetingSchedulesData,
		type MeetingSchedule,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Header } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { mdiFilter } from "@mdi/js";
	import ScheduledMeetingCard from "$features/meetings/components/scheduled-meeting-card/ScheduledMeetingCard.svelte";

	let queryParams = $state<ListMeetingSchedulesData["query"]>({});
	const query = createQuery(() => listMeetingSchedulesOptions({ query: { ...queryParams } }));
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Header title="Scheduled Meetings">
		<svelte:fragment slot="actions">
			<Button icon={mdiFilter}>filters</Button>
		</svelte:fragment>
	</Header>

	<div class="flex-1 flex flex-col gap-2 overflow-y-auto">
		<LoadingQueryWrapper {query}>
			{#snippet view(schedules: MeetingSchedule[])}
				{#each schedules as schedule}
					<ScheduledMeetingCard {schedule} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
