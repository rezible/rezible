<script lang="ts">
	import { useSvelteFlow, ViewportPortal, type XYPosition } from "@xyflow/svelte";
	import { Button, ButtonGroup } from "svelte-ux";
	import { analysis } from "../analysis.svelte";

	const { screenToFlowPosition } = useSvelteFlow();

	const visible = $derived(!!analysis.addingComponent);

	let pos = $state<XYPosition>({x: 0, y: 0});
	const onPointerMove = (e: PointerEvent) => {
		if (!visible) return;
		pos = screenToFlowPosition({x: e.clientX, y: e.clientY})
	}
</script>

<svelte:body onpointermove={onPointerMove} />

<ViewportPortal>
	{#if !!visible && !!analysis.addingComponent}
		<div class="absolute border rounded-lg bg-surface-100 p-1 z-10 opacity-75" style="left: {pos.x}px; top: {pos.y}px">
			<span>adding: {analysis.addingComponent.attributes.name}</span>
		</div>
	{/if}
</ViewportPortal>
