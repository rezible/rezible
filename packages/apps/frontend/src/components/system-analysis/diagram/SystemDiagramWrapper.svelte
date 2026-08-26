<script lang="ts">
	import { SvelteFlowProvider } from "@xyflow/svelte";

	import { SystemDiagramState, setSystemDiagram } from "./diagramController.svelte";
	import SystemDiagram from "./SystemDiagram.svelte";

	let containerEl = $state<HTMLElement>(null!);
	const diagram = new SystemDiagramState(() => containerEl);
	setSystemDiagram(diagram);
</script>

<SvelteFlowProvider>
	<div
		class="h-full w-full overflow-hidden relative"
		role="presentation"
		bind:this={containerEl}
		oncontextmenu={(e) => e.preventDefault()}
	>
		<SystemDiagram />
	</div>
</SvelteFlowProvider>
