<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { SystemMap } from "$features/systems/components/system-map";
	import { initSystemMapFixtureController } from "./controller.svelte";
	import FixtureControls from "./FixtureControls.svelte";

	const view = initSystemMapFixtureController();

	registerPageDescriptor(() => ({title: "System Map"}));
</script>

<svelte:head>
	<title>System Map · Rezible</title>
</svelte:head>

<section class="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden p-3">
	<FixtureControls />

	<div class="border-border min-h-0 min-w-0 flex-1 overflow-hidden rounded-md border">
		{#key view.scenario}
			<SystemMap 
				graph={view.graph} 
				displayOptions={view.displayOptions}
			/>
		{/key}
	</div>
</section>
