<script lang="ts">
	import { type MeetingSchedule as MeetingScheduleType } from "$lib/api";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initMeetingScheduleViewController } from "./controller.svelte";
	import MeetingSchedule from "./MeetingSchedule.svelte";

	const { id }: IdProp = $props();

	const view = initMeetingScheduleViewController(() => id);
	const query = $derived(view.query);

	registerPageDescriptor(() => ({
		title: view.title ?? "Meeting schedule",
		parents: [
			{ label: "Meetings", path: resolve("/meetings") },
			{ label: "Scheduled", path: resolve("/meetings/scheduled") },
		],
	}));
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(schedule: MeetingScheduleType)}
		<MeetingSchedule {schedule} />
	{/snippet}
</LoadingQueryWrapper>
