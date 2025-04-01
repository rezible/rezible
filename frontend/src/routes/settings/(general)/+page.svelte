<script>
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import GeneralSettingsView from "$features/settings/views/general/GeneralSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);
</script>

<GeneralSettingsView />
