<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	
	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
