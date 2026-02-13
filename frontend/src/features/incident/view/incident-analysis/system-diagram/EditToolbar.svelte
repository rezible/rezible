<script lang="ts">
	import { useSvelteFlow, ViewportPortal } from "@xyflow/svelte";
	import { Button } from "$components/ui/button";
	import { useSystemDiagram, type SystemComponentNodeData, type SystemRelationshipEdgeData } from "./diagramState.svelte";
	import { IsMounted } from "runed";
	import { mdiPencil, mdiTrashCan } from "@mdi/js";
	import { useIncidentAnalysis } from "../analysisState.svelte";
	import { useComponentDialog } from "./component-dialog/dialogState.svelte";
	import { useRelationshipDialog } from "./relationship-dialog/dialogState.svelte";

	const analysis = useIncidentAnalysis();
	const diagram = useSystemDiagram();
	const componentDialog = useComponentDialog();
	const relationshipDialog = useRelationshipDialog();
	const { getNodesBounds } = useSvelteFlow();

	const { node, edge } = $derived(diagram.selected);

	const nodeData = $derived(node?.data as SystemComponentNodeData | undefined);
	const component = $derived(nodeData?.analysisComponent);

	const edgeData = $derived(edge?.data as SystemRelationshipEdgeData | undefined);
	const relationship = $derived(edgeData?.relationship);

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

	const openEditDialog = () => {
		if (!!component) componentDialog.setEditing(component);
		if (!!relationship) relationshipDialog.setEditing(relationship);
		diagram.setSelected({});
	};
	
	const confirmDelete = () => {
		if (component && confirm("Remove this component?")) {
			analysis.removeComponent(component.id);
		} else if (relationship && confirm("Remove this relationship?")) {
			analysis.removeRelationship(relationship.id);
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
				<Button onclick={openEditDialog}>edit</Button>
				<Button onclick={confirmDelete}>delete</Button>
			</div>
		{/if}
	</ViewportPortal>
{/if}