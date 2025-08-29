<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { appShell } from "$features/app-shell";

	const session = useAuthSessionState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Initial Setup", href: "/setup" },
	]);
</script>

<span>setup</span>