<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Alerts" }]);
</script>

<span>alerts</span>