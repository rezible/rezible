<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import MeetingsListView from "$features/meetings-list/view/MeetingsListView.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Meetings" }]);
</script>

<MeetingsListView />
