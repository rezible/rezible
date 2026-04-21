<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	
	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
