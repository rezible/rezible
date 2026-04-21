<script lang="ts">
	import { type MeetingSchedule as MeetingScheduleType } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { initMeetingScheduleViewController } from "./controller.svelte";
	import MeetingSchedule from "./MeetingSchedule.svelte";

	const { id }: IdProp = $props();

	const view = initMeetingScheduleViewController(() => id);
	const query = $derived(view.query);

	setPageBreadcrumbs(() => [
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