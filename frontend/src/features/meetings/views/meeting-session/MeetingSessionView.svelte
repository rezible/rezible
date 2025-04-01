<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getMeetingSessionOptions, type MeetingSession } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SessionContent from "./SessionContent.svelte";

	type Props = { sessionId: string; };
	const { sessionId }: Props = $props();

	const query = createQuery(() => getMeetingSessionOptions({ path: { id: sessionId } }));
	const sessionTitle = $derived(query.data?.data.attributes.title);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions", href: "/meetings/sessions" },
		{ label: sessionTitle, href: `/meetings/sessions/${sessionId}` },
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(session: MeetingSession)}
		<SessionContent {session} />
	{/snippet}
</LoadingQueryWrapper>