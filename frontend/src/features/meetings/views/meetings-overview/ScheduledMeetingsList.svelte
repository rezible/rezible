<script lang="ts">
	import {
		listMeetingSchedulesOptions,
		type ListMeetingSchedulesData,
		type MeetingSchedule,
		type MeetingScheduleTiming,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { ListItem } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { addMinutes } from "date-fns";
	import SectionHeader from "$components/section-header/SectionHeader.svelte";

	let queryParams = $state<ListMeetingSchedulesData["query"]>({});
	const query = createQuery(() => listMeetingSchedulesOptions({ query: { ...queryParams } }));

	const getScheduleTimeDisplay = (m: MeetingScheduleTiming) => {
		// TODO: Implement this
		return addMinutes(Date.now(), 60).toLocaleString();
	};
</script>

<div class="flex flex-col min-h-0 h-full">
	<SectionHeader title="Your Meetings" />

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
