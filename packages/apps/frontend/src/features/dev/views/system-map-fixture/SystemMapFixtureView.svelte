<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { DeploymentView } from "$features/system/components/deployment-view";
	import { MapInspector } from "$features/system/components/map-inspector";
	import { SystemMap } from "$features/system/components/system-map";
	import { initSystemMapFixtureController } from "./controller.svelte";
	import FixtureControls from "./FixtureControls.svelte";

	const view = initSystemMapFixtureController();

	registerPageDescriptor(() => ({ title: "System Map" }));
</script>

<svelte:head>
	<title>System Map · Rezible</title>
</svelte:head>

<section class="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden p-3">
	<FixtureControls />

	<div class="border-border flex min-h-0 min-w-0 flex-1 overflow-hidden rounded-md border">
		<div class="min-h-0 min-w-0 flex-1" class:hidden={view.mapView === "deployment"}>
			<SystemMap
				bind:this={view.mapHandle}
				graph={view.graph}
				selection={view.selection}
				highlights={view.highlights}
				onSelectionChange={view.setSelection}
				onExploreDeployment={view.exploreDeployment}
			/>
		</div>
		{#if view.mapView === "deployment"}
			<div class="min-h-0 min-w-0 flex-1">
				<DeploymentView
					graph={view.graph}
					selection={view.selection}
					onSelectionChange={view.setSelection}
					focusTargetId={view.deploymentFocusTargetId}
					onShowInStructure={view.showInStructure}
				/>
			</div>
		{/if}
		<MapInspector graph={view.graph} selection={view.selection} onSelectionChange={view.setSelection} />
	</div>
</section>
