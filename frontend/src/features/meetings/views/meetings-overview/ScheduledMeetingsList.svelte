<script lang="ts">
	import {
		listMeetingSchedulesOptions,
		type ListMeetingSchedulesData,
		type MeetingSchedule,
		type MeetingScheduleTiming,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, ListItem, Header, Icon } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { mdiChevronDown, mdiChevronRight, mdiFilter } from "@mdi/js";
	import ScheduledMeetingCard from "$features/meetings/components/scheduled-meeting-card/ScheduledMeetingCard.svelte";
	import { addMinutes } from "date-fns";

	let queryParams = $state<ListMeetingSchedulesData["query"]>({});
	const query = createQuery(() => listMeetingSchedulesOptions({ query: { ...queryParams } }));

	const getScheduleTimeDisplay = (m: MeetingScheduleTiming) => {
		// TODO: Implement this
		return addMinutes(Date.now(), 60).toLocaleString();
	};
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
				{#each schedules as sched}
					<a href="/meetings/scheduled/{sched.id}">
						<ListItem
							title={sched.attributes.name}
							subheading={getScheduleTimeDisplay(sched.attributes.timing)}
							href={`/meetings/scheduled/${sched.id}`}
							classes={{ root: "border first:border-t rounded elevation-0 group hover:bg-primary-100/10" }}
							noShadow
						>
							<div slot="actions" class="group-hover:text-primary">
								<Icon data={mdiChevronRight} />
							</div>
						</ListItem>
					</a>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
