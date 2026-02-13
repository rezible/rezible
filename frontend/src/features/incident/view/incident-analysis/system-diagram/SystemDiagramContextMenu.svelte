<script lang="ts">
	import { useEdges, useNodes, useSvelteFlow } from "@xyflow/svelte";
	import { Button } from "$components/ui/button";
	import { mdiPlusCircle, mdiTrashCan } from "@mdi/js";
	
	import { useSystemDiagram } from "./diagramState.svelte";
	import { useComponentDialog } from "./component-dialog/dialogState.svelte";
	import AnalysisContextMenu from "$features/incident/components/analysis-context-menu/AnalysisContextMenu.svelte";

	type Props = {
		nodeId?: string;
		edgeId?: string;
		containerRect: DOMRect;
		clickPos: {x: number, y: number};
	};
	const { nodeId, edgeId, containerRect, clickPos }: Props = $props();

	const { screenToFlowPosition } = useSvelteFlow();
	const componentDialog = useComponentDialog();
	const diagram = useSystemDiagram();
	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = (nodeId: string) => {
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};

	const addComponent = () => {
		const flowPos = screenToFlowPosition(clickPos, {snapToGrid: true});
		componentDialog.setAdding(flowPos);
		diagram.closeContextMenu();
	}
</script>

<AnalysisContextMenu title="Diagram Actions" {containerRect} {clickPos}>
	{#if nodeId}
		<Button onclick={() => {deleteNode(nodeId)}}>
			Delete Component
		</Button>
	{:else if edgeId}
		<span>relationship</span>
	{:else}
		<Button onclick={addComponent}>
			Add Component
		</Button>
	{/if}
</AnalysisContextMenu>