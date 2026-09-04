<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initGeneralSettingsController } from "./controller.svelte";

	const controller = initGeneralSettingsController();

	registerPageDescriptor(() => ({ title: "Settings" }));
</script>

<div class="flex flex-col gap-2">
	<span>general</span>
</div>
