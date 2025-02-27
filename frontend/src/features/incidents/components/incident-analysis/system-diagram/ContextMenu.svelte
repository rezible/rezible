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
	import { diagram } from "./diagram.svelte";
	import { componentDialog } from "$features/incidents/components/incident-analysis/component-dialog/dialogState.svelte";

	const props = $derived(diagram.ctxMenuProps);
	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set($nodes.filter(({ id }) => id !== nodeId));
		edges.set($edges.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

	const addComponent = () => {
		if (!props) return;
		const {x, y} = $state.snapshot(props);
		componentDialog.setAdding({x, y});
		diagram.closeContextMenu();
	}
</script>

{#if !!props}
	<div
		style="top: {props.top}px; left: {props.left}px; width: {ContextMenuWidth}px; height: {ContextMenuHeight}px"
		class="absolute context-menu border bg-surface-200"
	>
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
	<p style="margin: 0.5em;">
		<small>node id: {id}</small>
	</p>
	<button onclick={() => {deleteNode(id)}}>delete node</button>
{/snippet}

{#snippet edgeMenu(id: string)}
	<p style="margin: 0.5em;">
		<small>edge id: {id}</small>
	</p>
{/snippet}

{#snippet paneMenu()}
	<button onclick={addComponent}>add component</button>
{/snippet}

<style>
	.context-menu {
		border-style: solid;
		box-shadow: 10px 19px 20px rgba(0, 0, 0, 10%);

		z-index: 10;
	}

	.context-menu button {
		border: none;
		display: block;
		padding: 0.5em;
		text-align: left;
		width: 100%;
	}

	.context-menu button:hover {
		background: white;
	}
</style>
