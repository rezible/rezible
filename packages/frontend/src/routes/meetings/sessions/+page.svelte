<script lang="ts">
	import { appShell } from "$features/app";
	
	appShell.setPageBreadcrumbs(() => [
		{ label: "Meetings", href: "/meetings" },
		{ label: "Sessions" },
	]);
</script>

<span>
	meeting sessions
</span>
