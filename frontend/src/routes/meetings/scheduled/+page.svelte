<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Meetings", href: "/meetings" }, { label: "Scheduled" }]);
</script>

<span>scheduled meetings</span>
