<script lang="ts">
	import type { Snippet } from "svelte";
	import { SvelteFlowProvider } from "@xyflow/svelte";
	import { initDiagramController, type GraphInteraction } from "./diagramController.svelte";
	import SystemDiagramCanvas from "./SystemDiagramCanvas.svelte";
	import SelectionInspector from "./panels/SelectionInspector.svelte";

	type Props = { 
		interaction?: GraphInteraction; 
		inspector?: Snippet;
	};
	let { interaction, inspector }: Props = $props();
	
	const controller = initDiagramController(() => interaction);
</script>

<SvelteFlowProvider>
	<div class="flex h-full w-full min-h-0 gap-3 overflow-hidden">
		<div
			class="relative min-w-0 flex-1 overflow-hidden"
			role="presentation"
			bind:this={controller.containerEl}
			oncontextmenu={(e) => e.preventDefault()}
		>
			<SystemDiagramCanvas />
		</div>
		{#if inspector}
			{@render inspector()}
		{:else}
			<SelectionInspector />
		{/if}
	</div>
</SvelteFlowProvider>
