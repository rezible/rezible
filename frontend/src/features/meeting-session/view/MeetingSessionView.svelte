<script lang="ts">
	import { type MeetingSession } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import SessionContent from "./SessionContent.svelte";
	import { useMeetingSessionViewState } from "$features/meeting-session";

	const viewState = useMeetingSessionViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions", href: "/meetings/sessions" },
		{ label: viewState.title, href: `/meetings/sessions/${viewState.sessionId}` },
	]);
</script>

<LoadingQueryWrapper query={viewState.query}>
	{#snippet view(session: MeetingSession)}
		<SessionContent {session} />
	{/snippet}
</LoadingQueryWrapper>