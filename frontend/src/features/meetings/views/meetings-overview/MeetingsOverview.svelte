<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import MeetingsPageActions from "$features/meetings/components/meetings-page-actions/MeetingsPageActions.svelte";
	import ScheduledMeetingsList from "./ScheduledMeetingsList.svelte";
	import UpcomingSessionsList from "./UpcomingSessionsList.svelte";

	appShell.setPageActions(MeetingsPageActions, true);
</script>

<SplitPage>
	{#snippet nav()}
		<ScheduledMeetingsList />
	{/snippet}

	<UpcomingSessionsList />
</SplitPage>
