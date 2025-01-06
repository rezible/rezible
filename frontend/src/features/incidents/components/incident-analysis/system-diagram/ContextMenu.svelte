<script lang="ts" module>
	export type ContextMenuProps = {
		id: string;
		top?: number;
		left?: number;
		right?: number;
		bottom?: number;
	}

</script>

<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";

	type Props = ContextMenuProps & {
		onClick: () => void;
	};
	const { onClick, id, top, left, right, bottom }: Props = $props();

	const nodes = useNodes();
	const edges = useEdges();

	const duplicateNode = () => {
		const node = $nodes.find((node) => node.id === id);
		if (node) {
			$nodes.push({
				...node,
				// You should use a better id than this in production
				id: `${id}-copy${Math.random()}`,
				position: {
					x: node.position.x,
					y: node.position.y + 50,
				},
			});
		}
		$nodes = $nodes;
	}

	const deleteNode = () => {
		$nodes = $nodes.filter((node) => node.id !== id);
		$edges = $edges.filter(
			(edge) => edge.source !== id && edge.target !== id,
		);
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	style="top: {top}px; left: {left}px; right: {right}px; bottom: {bottom}px;"
	class="absolute context-menu"
	onclick={onClick}
>
	<p style="margin: 0.5em;">
		<span>{top} {left}</span>
		<small>node: {id}</small>
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
