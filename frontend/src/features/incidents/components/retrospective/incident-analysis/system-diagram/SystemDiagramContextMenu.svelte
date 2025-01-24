<script lang="ts" module>
	export type ContextMenuProps = {
		nodeId?: string;
		edgeId?: string;
		top?: number;
		left?: number;
	}

	export const ContextMenuWidth = 180;
	export const ContextMenuHeight = 250;
</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";

	const {
		nodeId,
		edgeId,
		top,
		left
	}: ContextMenuProps = $props();

	const nodes = useNodes();
	const edges = useEdges();

	const active = $derived(!!top && !!left);

	const duplicateNode = () => {
		const node = $nodes.find((node) => node.id === nodeId);
		if (node) {
			$nodes.push({
				...node,
				// TODO: use better id
				id: `${nodeId}-copy${Math.random()}`,
				position: {
					x: node.position.x,
					y: node.position.y + 50,
				},
			});
		}
		$nodes = $nodes;
	}

	const deleteNode = () => {
		$nodes = $nodes.filter((node) => node.id !== nodeId);
		$edges = $edges.filter(
			(edge) => edge.source !== nodeId && edge.target !== nodeId,
		);
	}
</script>

{#if active}
<div
	style="top: {top}px; left: {left}px; width: {ContextMenuWidth}px; height: {ContextMenuHeight}px"
	class="absolute context-menu border bg-surface-200"
>
	<p style="margin: 0.5em;">
		{#if nodeId}
			<small>node: {nodeId}</small>
		{:else if edgeId}
			<small>edge: {edgeId}</small>
		{:else}
			<small>menu</small>
		{/if}
	</p>
	<button onclick={duplicateNode}>duplicate</button>
	<button onclick={deleteNode}>delete</button>
</div>
{/if}

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
