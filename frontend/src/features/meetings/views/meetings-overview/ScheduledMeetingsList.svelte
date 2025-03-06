<script lang="ts">
	import {
		listMeetingSchedulesOptions,
		type ListMeetingSchedulesData,
		type MeetingSchedule,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Collapse, Header, Icon } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { mdiChevronDown, mdiFilter } from "@mdi/js";
	import ScheduledMeetingCard from "$features/meetings/components/scheduled-meeting-card/ScheduledMeetingCard.svelte";

	let queryParams = $state<ListMeetingSchedulesData["query"]>({});
	const query = createQuery(() => listMeetingSchedulesOptions({ query: { ...queryParams } }));
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Header title="Scheduled" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<div slot="actions" class="">
			<Button icon={mdiFilter} iconOnly>
				<span class="flex gap-1 items-center">
					Filters
					<Icon data={mdiChevronDown} />
				</span>
			</Button>
		</div>
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
