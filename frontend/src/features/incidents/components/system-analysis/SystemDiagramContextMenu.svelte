<script lang="ts" module>
	export type ContextMenuProps = {
		nodeId: string;
		top?: number;
		left?: number;
		right?: number;
		bottom?: number;
	}

</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";

	const { nodeId, top, left, right, bottom }: ContextMenuProps = $props();

	const nodes = useNodes();
	const edges = useEdges();

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

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	style="top: {top}px; left: {left}px; right: {right}px; bottom: {bottom}px;"
	class="absolute context-menu"
	onclick={() => {console.log("clicked")}}
>
	<p style="margin: 0.5em;">
		<span>{top} {left}</span>
		<small>node: {nodeId}</small>
	</p>
	<button onclick={duplicateNode}>duplicate</button>
	<button onclick={deleteNode}>delete</button>
</div>

<style>
	.context-menu {
		background: white;
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
