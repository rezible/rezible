<script lang="ts">
	import MeetingsOverview from "$features/meetings/views/meetings-overview/MeetingsOverview.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
</script>

<MeetingsOverview />
