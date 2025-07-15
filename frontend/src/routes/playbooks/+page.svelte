<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import PlaybooksList from "$src/features/playbooks/views/playbooks-list/PlaybooksList.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Playbooks" }]);
</script>

<PlaybooksList />