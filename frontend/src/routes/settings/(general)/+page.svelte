<script>
	import { appShell } from "$features/app";
	import GeneralSettingsView from "$features/settings/view/general/GeneralSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);
</script>

<GeneralSettingsView />
