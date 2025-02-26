<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getMeetingScheduleOptions, type MeetingSchedule } from "$lib/api";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import ScheduledMeeting from "./ScheduledMeeting.svelte";

	type Props = {
		scheduleId: string;
	};
	const { scheduleId }: Props = $props();

	const query = createQuery(() => getMeetingScheduleOptions({ path: { id: scheduleId } }));
	const title = $derived(query.data?.data.attributes.name);

	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled", href: "/meetings/scheduled" },
		{ label: title, href: `/meetings/scheduled/${scheduleId}` },
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(schedule: MeetingSchedule)}
		<ScheduledMeeting {schedule} />
	{/snippet}
</LoadingQueryWrapper>