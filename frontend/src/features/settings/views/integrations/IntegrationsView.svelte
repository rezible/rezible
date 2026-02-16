<script lang="ts">
    import { appShell } from "$features/app";
	import { initIntegrationsViewController } from "./integrationsViewController.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Integrations" }
	]);

	const view = initIntegrationsViewController();
</script>

<div class="flex flex-col gap-2">

</div>
