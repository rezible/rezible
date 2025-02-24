<script>
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import GeneralSettingsView from "$features/settings/views/general/GeneralSettingsView.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);
</script>

<GeneralSettingsView />
