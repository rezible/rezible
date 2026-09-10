<script lang="ts">
	import { useEdges, useNodes } from "@xyflow/svelte";
	import { Button } from "$components/ui/button";
	import ContextMenu from "$components/common/context-menu/ContextMenu.svelte";
	import { useSystemDiagram } from "../diagramController.svelte";

	const diagram = useSystemDiagram();

	const props = $derived(diagram.contextMenu);

	const nodes = useNodes();
	const edges = useEdges();

	const deleteNode = () => {
		const nodeId = props?.nodeId;
		if (!nodeId) return;
		nodes.set(nodes.current.filter(({ id }) => id !== nodeId));
		edges.set(edges.current.filter(({ source, target }) => source !== nodeId && target !== nodeId));
	};
</script>

{#if !!props}
	<ContextMenu title="Diagram Actions" containerRect={props.containerRect} clickPos={props.clickPos}>
		{#if props.nodeId}
			<Button onclick={deleteNode}>Delete Entity</Button>
		{:else if props.edgeId}
			<span>relationship</span>
		{:else}
			<span>diagram</span>
		{/if}
	</ContextMenu>
{/if}
