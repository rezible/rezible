<script lang="ts">
    import { appShell } from "$features/app";
	import { initGeneralSettingsViewController } from "./generalSettingsViewController.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "General" }
	]);

	const view = initGeneralSettingsViewController();
</script>

<div class="flex flex-col gap-2">

</div>
