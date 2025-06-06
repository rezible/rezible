<script lang="ts">
	import { useEdges, useNodes, useSvelteFlow } from "@xyflow/svelte";
	import { Button } from "svelte-ux";
	import { mdiPlusCircle, mdiTrashCan } from "@mdi/js";
	
	import { useSystemDiagram } from "./diagramState.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { ElementSize } from "runed";

	type Props = {
		nodeId?: string;
		edgeId?: string;
		containerRect: DOMRect;
		clickPos: {x: number, y: number};
	};
	const { nodeId, edgeId, containerRect, clickPos }: Props = $props();

	const { screenToFlowPosition } = useSvelteFlow();
	const diagram = useSystemDiagram();
	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

	let ref = $state<HTMLElement>(null!);
	const size = new ElementSize(() => ref);

	const pos = $derived.by(() => {
		const xOverflows = (clickPos.x + size.width) > containerRect.right;
		const naiveX = Math.round(clickPos.x - containerRect.x);

		const yOverflows = (clickPos.y + size.height) > containerRect.bottom;
		const naiveY = Math.round(clickPos.y - containerRect.y);
		
		return {
			left: xOverflows ? (naiveX - size.width) : naiveX,
			top: yOverflows ? (naiveY - size.height) : naiveY,
		};
	});

	const addComponent = () => {
		const flowPos = screenToFlowPosition(clickPos, {snapToGrid: true});
		diagram.componentDialog.setAdding(flowPos);
		diagram.closeContextMenu();
	}
</script>

<div
	style="left: {pos.left}px; top: {pos.top}px;"
	class="absolute context-menu border bg-surface-200 w-48 h-fit"
	bind:this={ref}
>
	<Header title="Diagram Actions" classes={{root: "px-2 py-1"}} />

	{#if nodeId}
		{@render nodeMenu(nodeId)}
	{:else if edgeId}
		{@render edgeMenu(edgeId)}
	{:else}
		{@render paneMenu()}
	{/if}
</div>

{#snippet nodeMenu(id: string)}
	<Button variant="fill-light" icon={mdiTrashCan} rounded={false} classes={{root: "w-full gap-2"}} on:click={() => {deleteNode(id)}}>
		Delete Component
	</Button>
{/snippet}

{#snippet edgeMenu(id: string)}
	edge
{/snippet}

{#snippet paneMenu()}
	<Button variant="fill-light" icon={mdiPlusCircle} rounded={false} classes={{root: "w-full gap-2"}} on:click={addComponent}>
		Add Component
	</Button>
{/snippet}

<style>
	.context-menu {
		border-style: solid;
		box-shadow: 10px 19px 20px rgba(0, 0, 0, 10%);

		z-index: 10;
	}
</style>
