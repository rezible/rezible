<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Meetings", href: "/meetings" }, { label: "Scheduled" }]);
</script>

<span>scheduled meetings</span>
