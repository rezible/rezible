<script lang="ts">
	import type { SystemAnalysisComponent, SystemAnalysisRelationship } from "$lib/api";
	import { useSvelteFlow, ViewportPortal } from "@xyflow/svelte";
	import { Button, ButtonGroup } from "svelte-ux";
	import { useSystemDiagram } from "./diagramState.svelte";
	import { IsMounted } from "runed";

	const { getNodesBounds } = useSvelteFlow();
	const diagram = useSystemDiagram();

	const { node, edge } = $derived(diagram.selected);
	const component = $derived(node?.data?.component as SystemAnalysisComponent | undefined);
	const relationship = $derived(edge?.data?.relationship as SystemAnalysisRelationship | undefined);

	const rect = $derived.by(() => {
		if (!diagram.flow) return;
		if (node) {
			return getNodesBounds([node]);
		} else if (edge) {
			return getNodesBounds([edge.source, edge.target]);
		}
	});

	const transform = $derived.by(() => {
		if (!diagram.selectedLivePosition || !rect) return;
		const { x, y } = diagram.selectedLivePosition;
		const posX = (x + rect.width / 2);
		const offset = 10;
		const posY = !!node ? (y + rect.height + offset) : (y + rect.height / 2) - offset;
		return `translate(${posX}px, ${posY}px) translate(-50%, 0%)`
	});

	const openEditComponentDialog = () => {
		if (component) diagram.componentDialog.setEditing(component);
	};

	const openEditRelationshipDialog = () => {
		if (relationship) diagram.relationshipDialog.setEditing(relationship);
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
{/if}