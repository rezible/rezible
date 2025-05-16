<script lang="ts" module>
	export type ContextMenuProps = {
		nodeId?: string;
		edgeId?: string;
		top: number;
		left: number;
		x: number;
		y: number;
	};

	export const ContextMenuWidth = 180;
	export const ContextMenuHeight = 250;
</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";
	import { Button, Header } from "svelte-ux";
	import { mdiPlusCircle, mdiTrashCan } from "@mdi/js";
	
	import { useSystemDiagram } from "./diagramState.svelte";

	const diagram = useSystemDiagram();

	const props = $derived(diagram.ctxMenuProps);
	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

	const addComponent = () => {
		if (!props) return;
		const {x, y} = $state.snapshot(props);
		diagram.componentDialog.setAdding({x, y});
		diagram.closeContextMenu();
	}
</script>

{#if !!props}
	<div
		style="top: {props.top}px; left: {props.left}px; width: {ContextMenuWidth}px; max-height: {ContextMenuHeight}px"
		class="absolute context-menu border bg-surface-200"
	>
		<Header title="Actions" class="px-2 py-1" />

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
