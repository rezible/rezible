<script lang="ts">
	import { useSvelteFlow, ViewportPortal } from "@xyflow/svelte";
	import { Button } from "$components/ui/button";
	import { useSystemDiagram, type SystemTopologyNodeData, type SystemRelationshipEdgeData } from "./diagramState.svelte";
	import { IsMounted } from "runed";
	import { useIncidentAnalysis } from "../controller.svelte";

	const analysis = useIncidentAnalysis();
	const diagram = useSystemDiagram();
	const { getNodesBounds } = useSvelteFlow();

	const { node, edge } = $derived(diagram.selected);

	const nodeData = $derived(node?.data as SystemTopologyNodeData | undefined);
	const analysisNode = $derived(nodeData?.analysisNode);

	const edgeData = $derived(edge?.data as SystemRelationshipEdgeData | undefined);
	const analysisEdge = $derived(edgeData?.edge);

	const rect = $derived.by(() => {
		if (node) return getNodesBounds([node]);
		if (edge) return getNodesBounds([edge.source, edge.target]);
	});

	const transform = $derived.by(() => {
		if (!diagram.selectedLivePosition || !rect) return;
		const { x, y } = diagram.selectedLivePosition;
		const posX = (x + rect.width / 2);
		const offset = 10;
		const posY = !!node ? (y + rect.height + offset) : (y + rect.height / 2) - offset;
		return `translate(${posX}px, ${posY}px) translate(-50%, 0%)`
	});

	const confirmDelete = () => {
		if (analysisNode && confirm("Remove this node?")) {
			analysis.removeNode(analysisNode.id);
		} else if (analysisEdge && confirm("Remove this edge?")) {
			analysis.removeEdge(analysisEdge.id);
		}
		diagram.setSelected({});
	};

	const mounted = new IsMounted();
</script>

{#if mounted.current}
	<ViewportPortal target="front">
		{#if transform}
			<div
				class="pointer-events-auto absolute border rounded-lg bg-surface-100 p-1 z-[1001]"
				style:transform
			>
				<Button onclick={confirmDelete}>delete</Button>
			</div>
		{/if}
	</ViewportPortal>
{/if}
