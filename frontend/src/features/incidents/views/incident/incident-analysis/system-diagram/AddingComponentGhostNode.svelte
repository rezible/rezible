<script lang="ts">
	import { useSvelteFlow, ViewportPortal, type XYPosition } from "@xyflow/svelte";
	import { useSystemDiagram } from "./diagramState.svelte";

	const diagram = useSystemDiagram();

	const { screenToFlowPosition } = useSvelteFlow();

	const addingComponent = $derived(diagram.addingComponent);

	let pos = $state<XYPosition>({x: 0, y: 0});
	const onPointerMove = (e: PointerEvent) => {
		if (!addingComponent) return;
		pos = screenToFlowPosition({x: e.clientX, y: e.clientY})
	}
</script>

<svelte:body onpointermove={onPointerMove} />

<ViewportPortal>
	{#if !!addingComponent}
		<div class="absolute border rounded-lg bg-surface-100 p-1 z-10 opacity-75" style="left: {pos.x}px; top: {pos.y}px">
			<span>adding: {addingComponent.attributes.name}</span>
		</div>
	{/if}
</ViewportPortal>
