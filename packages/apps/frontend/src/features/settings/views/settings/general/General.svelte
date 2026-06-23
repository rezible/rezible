<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { initGeneralSettingsController } from "./controller.svelte";

	const controller = initGeneralSettingsController();

	setPageBreadcrumbs(() => ([
		{ label: "Settings", path: "/settings" },
	]));
</script>

<div class="flex flex-col gap-2">
	<span>general</span>
</div>
