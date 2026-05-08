<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";
	import { Button } from "$components/ui/button";
	import AnalysisContextMenu from "../ContextMenu.svelte";

	type Props = {
		nodeId?: string;
		edgeId?: string;
		containerRect: DOMRect;
		clickPos: {x: number, y: number};
	};
	const { nodeId, edgeId, containerRect, clickPos }: Props = $props();

	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

</script>

<AnalysisContextMenu title="Diagram Actions" {containerRect} {clickPos}>
	{#if nodeId}
		<Button onclick={() => {deleteNode(nodeId)}}>
			Delete Entity
		</Button>
	{:else if edgeId}
		<span>relationship</span>
	{:else}
		<span>diagram</span>
	{/if}
</AnalysisContextMenu>
