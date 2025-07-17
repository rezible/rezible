<script lang="ts">
	import MeetingsListView from "$features/meetings/views/meetings-list/MeetingsListView.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
</script>

<MeetingsListView />
