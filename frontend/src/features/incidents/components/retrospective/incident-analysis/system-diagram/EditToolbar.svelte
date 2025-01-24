<script lang="ts">
    import { ViewportPortal } from "@xyflow/svelte";
    import { diagram } from "./diagram.svelte";
    import { Button, ButtonGroup } from "svelte-ux";
    import { analysis } from "../analysis.svelte";
    import type { SystemAnalysisComponent, SystemAnalysisRelationship } from "$src/lib/api";

	const {node, edge} = $derived(diagram.selected);
	const component = $derived(node?.data ? node.data.component as SystemAnalysisComponent : undefined);
	const relationship = $derived(edge?.data ? edge.data.relationship as SystemAnalysisRelationship : undefined);
	const visible = $derived(!!node || !!edge);

	const containerTransformStyle = $derived(visible ? `transform: translate(-50%, -50%) translate(${diagram.toolbarPosition.x}px, ${diagram.toolbarPosition.y}px);` : "");
</script>

<ViewportPortal>
	{#if visible}
		<div class="pointer-events-auto absolute border rounded-lg bg-surface-100 p-1 z-10 svelte-flow__node-toolbar" style={containerTransformStyle}>
			<ButtonGroup variant="fill-light" color="accent" size="sm">
				{#if node}
					<Button on:click={() => analysis.setComponentDialogOpen(true, component)}>Edit Component</Button>
				{:else if edge}
					<Button on:click={() => analysis.setRelationshipDialogOpen(true, relationship)}>Edit Relationship</Button>
				{/if}
			</ButtonGroup>
		</div>
	{/if}
</ViewportPortal>