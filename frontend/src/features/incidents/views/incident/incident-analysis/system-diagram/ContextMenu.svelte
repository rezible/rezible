<script lang="ts" module>
	export type ContextMenuProps = {
		nodeId?: string;
		edgeId?: string;
		containerRect: DOMRect;
		clickPos: {x: number, y: number};
	};
</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";
	import { Button } from "svelte-ux";
	import { mdiPlusCircle, mdiTrashCan } from "@mdi/js";
	
	import { useSystemDiagram } from "./diagramState.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { ElementSize } from "runed";

	const diagram = useSystemDiagram();

	const props = $derived(diagram.ctxMenuProps);
	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

	let ref = $state<HTMLElement>(null!);
	const size = new ElementSize(() => ref);

	const pos = $derived.by(() => {
		if (!props) return;
		const {clickPos, containerRect} = props;

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
		if (!props) return;
		const {x, y} = $state.snapshot(props.clickPos);
		diagram.componentDialog.setAdding({x, y});
		diagram.closeContextMenu();
	}
</script>

{#if !!props && !!pos}
	<div
		style="left: {pos.left}px; top: {pos.top}px;"
		class="absolute context-menu border bg-surface-200 w-48 h-fit"
		bind:this={ref}
	>
		<Header title="Diagram Actions" classes={{root: "px-2 py-1"}} />

		{#if props.nodeId}
			{@render nodeMenu(props.nodeId)}
		{:else if props.edgeId}
			{@render edgeMenu(props.edgeId)}
		{:else}
			{@render paneMenu()}
		{/if}
	</div>
{/if}

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
