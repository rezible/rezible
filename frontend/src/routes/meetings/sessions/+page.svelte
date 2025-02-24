<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	
	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
