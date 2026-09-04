<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SessionContent from "./SessionContent.svelte";
	import { initMeetingSessionViewController } from "./controller.svelte";

	const { id }: IdProp = $props();
	const view = initMeetingSessionViewController(() => id);
	const query = $derived(view.query);

	registerPageDescriptor(() => ({
		title: view.title ?? "Meeting session",
		parents: [
			{ label: "Meetings", path: resolve("/meetings") },
			{ label: "Sessions", path: resolve("/meetings/sessions") },
		],
	}));
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(session: MeetingSession)}
		<SessionContent {session} />
	{/snippet}
</LoadingQueryWrapper>
