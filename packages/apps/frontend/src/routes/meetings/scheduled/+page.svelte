<script lang="ts">
	import { appShell } from "$features/app";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Scheduled" }
	]);
</script>

<span>scheduled meetings</span>
