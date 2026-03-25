<script lang="ts">
	import { appShell } from "$features/app";
	import { type MeetingSession } from "$lib/api";
	import type { IdProps } from "$lib/utils.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SessionContent from "./SessionContent.svelte";
	import { initMeetingSessionViewController } from "./controller.svelte";

	const { id }: IdProps = $props();
	const view = initMeetingSessionViewController(() => id);
	const query = $derived(view.query);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions", href: "/meetings/sessions" },
		{ label: view.title, href: `/meetings/sessions/${id}` },
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(session: MeetingSession)}
		<SessionContent {session} />
	{/snippet}
</LoadingQueryWrapper>