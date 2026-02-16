<script lang="ts">
	import { type MeetingSchedule as MeetingScheduleType } from "$lib/api";
	import type { IdProps } from "$lib/utils.svelte";
	import { appShell } from "$features/app";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { initMeetingScheduleViewController } from "./controller.svelte";
	import MeetingSchedule from "./MeetingSchedule.svelte";

	const { id }: IdProps = $props();

	const view = initMeetingScheduleViewController(() => id);
	const query = $derived(view.query);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled", href: "/meetings/scheduled" },
		{ label: view.title, href: `/meetings/scheduled/${id}` },
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(schedule: MeetingScheduleType)}
		<MeetingSchedule {schedule} />
	{/snippet}
</LoadingQueryWrapper>