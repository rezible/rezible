<script>
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import GeneralSettingsView from "$features/settings/view/general/GeneralSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);
</script>

<GeneralSettingsView />
