<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";

	setPageBreadcrumbs(() => [{ label: "Meetings", href: "/meetings" }, { label: "Scheduled" }]);
</script>

<span>scheduled meetings</span>
