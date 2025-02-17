<script lang="ts">
	import {
		listMeetingSchedulesOptions,
		type ListMeetingSchedulesData,
		type MeetingSchedule,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Collapse, Header } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { mdiFilter } from "@mdi/js";
	import ScheduledMeetingCard from "$features/meetings/components/scheduled-meeting-card/ScheduledMeetingCard.svelte";

	let queryParams = $state<ListMeetingSchedulesData["query"]>({});
	const query = createQuery(() => listMeetingSchedulesOptions({ query: { ...queryParams } }));
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Collapse>
		<div slot="trigger" class="flex-1">
			<Button icon={mdiFilter} iconOnly />
			Filters
		</div>
		<div class="px-3 pb-3 border-t">
		  Lorem ipsum dolor sit amet consectetur adipisicing elit. Reiciendis quod
		  culpa et, dolores omnis, ipsum in perspiciatis porro ut nihil molestiae
		  molestias tenetur delectus velit! Inventore laborum rerum at id?
		</div>
	</Collapse>

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
