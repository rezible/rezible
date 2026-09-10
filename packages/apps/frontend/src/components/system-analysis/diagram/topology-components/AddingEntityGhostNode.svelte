<script lang="ts">
	import { useSvelteFlow, ViewportPortal, type XYPosition } from "@xyflow/svelte";
	import { useSystemDiagram } from "../diagramController.svelte";

	const diagram = useSystemDiagram();

	const { screenToFlowPosition } = useSvelteFlow();

	const addingEntity = $derived(diagram.addingEntityGhost);
	const entityState = $derived(addingEntity?.attributes.latestState);

	let pos = $state<XYPosition>({ x: 0, y: 0 });
	const onPointerMove = (e: PointerEvent) => {
		if (!addingEntity) return;
		pos = screenToFlowPosition({ x: e.clientX, y: e.clientY });
	};
</script>

<svelte:body onpointermove={onPointerMove} />

<ViewportPortal target="front">
	{#if !!addingEntity}
		<div
			class="absolute border rounded-lg bg-card p-1 z-10 opacity-75"
			style="left: {pos.x}px; top: {pos.y}px"
		>
			<span>adding: {entityState?.displayName ?? "Unknown entity"}</span>
		</div>
	{/if}
</ViewportPortal>
