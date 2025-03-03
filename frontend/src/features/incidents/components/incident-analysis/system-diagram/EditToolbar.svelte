<script lang="ts">
	import { ViewportPortal } from "@xyflow/svelte";
	import { Button, ButtonGroup } from "svelte-ux";
	import type { SystemAnalysisComponent, SystemAnalysisRelationship } from "$lib/api";
	import { componentDialog } from "$features/incidents/components/incident-analysis/system-diagram/component-dialog/dialogState.svelte";
	import { relationshipDialog } from "$features/incidents/components/incident-analysis/system-diagram/relationship-dialog/dialogState.svelte";
	import { diagram } from "./diagram.svelte";

	const { node, edge } = $derived(diagram.selected);
	const component = $derived(node?.data ? (node.data.component as SystemAnalysisComponent) : undefined);
	const relationship = $derived(
		edge?.data ? (edge.data.relationship as SystemAnalysisRelationship) : undefined
	);
	const visible = $derived(!!node || !!edge);

	const containerTransformStyle = $derived(
		visible
			? `transform: translate(-50%, -50%) translate(${diagram.toolbarPosition.x}px, ${diagram.toolbarPosition.y}px);`
			: ""
	);

	const openEditComponentDialog = () => {
		if (component) componentDialog.setEditing(component);
	};

	const openEditRelationshipDialog = () => {
		if (relationship) relationshipDialog.setEditing(relationship);
	};
</script>

<ViewportPortal>
	{#if visible}
		<div
			class="pointer-events-auto absolute border rounded-lg bg-surface-100 p-1 z-10 svelte-flow__node-toolbar"
			style={containerTransformStyle}
		>
			<ButtonGroup variant="fill-light" color="accent" size="sm">
				{#if node}
					<Button on:click={openEditComponentDialog}>Edit Component</Button>
				{:else if edge}
					<Button on:click={openEditRelationshipDialog}>Edit Relationship</Button>
				{/if}
			</ButtonGroup>
		</div>
	{/if}
</ViewportPortal>
