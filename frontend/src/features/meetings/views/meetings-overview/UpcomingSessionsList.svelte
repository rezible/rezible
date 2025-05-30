<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Icon, Month } from "svelte-ux";
	import type { DateRange } from "@layerstack/utils/dateRange";
	import { startOfWeek, endOfWeek } from "date-fns";
	import { mdiChevronDown, mdiFilter } from "@mdi/js";
	import { listMeetingSessionsOptions, type ListMeetingSessionsData, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";
	import Header from "$src/components/header/Header.svelte";

	let queryParams = $state<ListMeetingSessionsData["query"]>({});
	const query = createQuery(() => listMeetingSessionsOptions({ query: queryParams }));

	let selectedWeek = $state<DateRange>({ from: startOfWeek(Date.now()), to: endOfWeek(Date.now()) });

	const onMonthDateChange = (e: CustomEvent<Date>) => {
		const date = e.detail;
		selectedWeek = { from: startOfWeek(date), to: endOfWeek(date) };
	};
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Header title="Upcoming" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			<Button icon={mdiFilter} iconOnly>
				<span class="flex gap-1 items-center">
					Filters
					<Icon data={mdiChevronDown} />
				</span>
			</Button>
		{/snippet}
	</Header>

	<div class="grid grid-cols-2 h-full gap-2">
		<div class="h-full flex flex-col gap-2">
			<div class="pb-2 border">
				<Month bind:selected={selectedWeek} on:dateChange={onMonthDateChange} showOutsideDays />
			</div>
		</div>

		<div class="flex-1 flex flex-col gap-2 overflow-y-auto p-2">
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
