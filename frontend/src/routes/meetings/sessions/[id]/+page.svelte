<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import { getMeetingSessionOptions, type MeetingSession } from "$lib/api";
	import PageContainer, { type Breadcrumb } from "$components/page-container/PageContainer.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionView from "$features/meetings/views/meeting-session/MeetingSessionView.svelte";

	const query = createQuery(() => getMeetingSessionOptions({ path: { id: page.params.id } }));

	const breadcrumbs = $derived<Breadcrumb[]>([
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions", href: "/meetings/sessions" },
		{ label: query.data?.data.attributes.title ?? "" },
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper {query}>
		{#snippet view(session: MeetingSession)}
			<MeetingSessionView {session} />
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>
