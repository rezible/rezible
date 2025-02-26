<script>
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import GeneralSettingsView from "$features/settings/views/general/GeneralSettingsView.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);
</script>

<GeneralSettingsView />
