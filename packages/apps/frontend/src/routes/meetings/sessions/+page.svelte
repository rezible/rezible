<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appShell.svelte";
	
	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
