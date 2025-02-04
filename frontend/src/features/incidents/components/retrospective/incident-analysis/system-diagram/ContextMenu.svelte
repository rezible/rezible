<script lang="ts" module>
	export type ContextMenuProps = {
		nodeId?: string;
		edgeId?: string;
		top?: number;
		left?: number;
	};

	export const ContextMenuWidth = 180;
	export const ContextMenuHeight = 250;
</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";

	const { nodeId, edgeId, top, left }: ContextMenuProps = $props();

	const nodes = useNodes();
	const edges = useEdges();

	const active = $derived(!!top && !!left);

	const deleteNode = () => {
		nodes.set($nodes.filter(({id}) => id !== nodeId));
		edges.set($edges.filter(({source, target}) => source !== nodeId && target !== nodeId));
	};
</script>

{#if active}
	<div
		style="top: {top}px; left: {left}px; width: {ContextMenuWidth}px; height: {ContextMenuHeight}px"
		class="absolute context-menu border bg-surface-200"
	>
		{#if nodeId}
			{@render nodeMenu(nodeId)}
		{:else if edgeId}
			{@render edgeMenu(edgeId)}
		{:else}
			{@render paneMenu()}
		{/if}
	</div>
{/if}

{#snippet nodeMenu(id: string)}
	<p style="margin: 0.5em;">
		<small>node id: {id}</small>
	</p>
	<button onclick={deleteNode}>delete node</button>
{/snippet}

{#snippet edgeMenu(id: string)}
	<p style="margin: 0.5em;">
		<small>edge id: {id}</small>
	</p>
{/snippet}

{#snippet paneMenu()}
	<p style="margin: 0.5em;">
		pane menu
	</p>
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
