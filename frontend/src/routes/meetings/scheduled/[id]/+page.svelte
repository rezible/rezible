<script lang="ts">
	import { page } from "$app/state";
	import { createQuery } from "@tanstack/svelte-query";
	import { getMeetingScheduleOptions, type MeetingSchedule } from "$lib/api";
	import PageContainer, {
		type Breadcrumb,
	} from "$components/page-container/PageContainer.svelte";
	import MeetingScheduleView from "$features/meetings/views/meeting-schedule/MeetingScheduleView.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	const meetingId = $derived(page.params.id);
	const queryOptions = () =>
		getMeetingScheduleOptions({ path: { id: meetingId } });
	const query = createQuery(queryOptions);

	const breadcrumbs = $derived<Breadcrumb[]>([
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled", href: "/meetings/scheduled" },
		{ label: query.data?.data.attributes.name ?? "" },
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper {query}>
		{#snippet view(schedule: MeetingSchedule)}
			<MeetingScheduleView {schedule} />
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>
