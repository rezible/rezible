<script lang="ts">
	import { type MeetingSchedule as MeetingScheduleType } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useMeetingScheduleViewState } from "$features/meeting-schedule";
	import MeetingSchedule from "./MeetingSchedule.svelte";

	const viewState = useMeetingScheduleViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled", href: "/meetings/scheduled" },
		{ label: viewState.title, href: `/meetings/scheduled/${viewState.scheduleId}` },
	]);
</script>

<LoadingQueryWrapper query={viewState.query}>
	{#snippet view(schedule: MeetingScheduleType)}
		<MeetingSchedule {schedule} />
	{/snippet}
</LoadingQueryWrapper>