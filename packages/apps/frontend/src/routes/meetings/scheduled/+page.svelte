<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appShell.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled" }
	]);
</script>

<span>scheduled meetings</span>
