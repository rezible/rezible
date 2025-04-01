<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	
	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
