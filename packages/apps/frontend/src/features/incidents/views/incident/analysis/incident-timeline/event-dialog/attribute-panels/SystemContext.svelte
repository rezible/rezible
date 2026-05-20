<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		listSystemAnalysisNodesOptions,
		type IncidentTimelineEventTopologyContext,
		type IncidentTimelineEventTopologyContextAttributes,
		type SystemAnalysisNode,
	} from "$lib/api";
	import { v4 as uuidv4 } from "uuid";
	import { SvelteMap } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiPlus } from "@mdi/js";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { useEventDialogAttributes } from "./attributes.svelte";
	import { useIncidentAnalysis } from "$features/incidents/views/incident/analysis/controller.svelte";

	const attributes = useEventDialogAttributes();

	const analysis = useIncidentAnalysis();
	const analysisId = $derived(analysis.analysisId);

	const analysisNodesQuery = createQuery(() => ({
		...listSystemAnalysisNodesOptions({ path: { id: analysisId } }),
		enabled: !!analysisId,
	}));
	const analysisNodes = $derived(analysisNodesQuery.data?.data ?? []);
	const analysisNodeMap = $derived(
		new SvelteMap(analysisNodes.map((node) => [node.attributes.snapshotEntity.id, node]))
	);

	let relationship = $state("affected");

	const getAttributes = (node: SystemAnalysisNode): IncidentTimelineEventTopologyContextAttributes => ({
		snapshotEntityId: $state.snapshot(node.attributes.snapshotEntity.id),
		relationship: $state.snapshot(relationship),
	});

	let selecting = $state(false);
	let selectedNode = $state<SystemAnalysisNode>();
	let editing = $state<IncidentTimelineEventTopologyContext>();
	const editNode = $derived(
		editing?.attributes.snapshotEntityId ? analysisNodeMap.get(editing.attributes.snapshotEntityId) : undefined
	);

	const setEditing = (cx: IncidentTimelineEventTopologyContext) => {
		editing = $state.snapshot(cx);
		relationship = $state.snapshot(cx.attributes.relationship);
	};

	const confirmDelete = (cx: IncidentTimelineEventTopologyContext) => {
		const node = cx.attributes.snapshotEntityId ? analysisNodeMap.get(cx.attributes.snapshotEntityId) : undefined;
		editing = undefined;
		if (!node || !confirm(`Are you sure you want to remove ${node.attributes.snapshotEntity.attributes.displayName}?`)) return;
		const idx = attributes.systemContext.findIndex((c) => c.id === cx.id);
		if (idx >= 0) attributes.systemContext.splice(idx, 1);
	};

	const resetState = () => {
		selecting = false;
		selectedNode = undefined;
		editing = undefined;
		relationship = "affected";
	}

	const onCancel = () => {
		if (selecting && selectedNode) {
			selectedNode = undefined;
		} else {
			resetState();
		}
	};

	const onConfirm = () => {
		if (selecting && selectedNode) {
			attributes.systemContext.push({
				id: uuidv4(),
				attributes: getAttributes(selectedNode),
			});
		} else if (editing && editNode) {
			const idx = attributes.systemContext.findIndex((c) => c.id === editing?.id);
			if (idx < 0) return;
			attributes.systemContext[idx].attributes = getAttributes(editNode);
		}
		resetState();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#snippet topologyContextEditor(node: SystemAnalysisNode)}
		{@const attrs = node.attributes.snapshotEntity.attributes}
		<span class="text-lg">{attrs.displayName}</span>

		<span>relationship select</span>
	{/snippet}

	{#snippet confirmButtons()}
		<div class="w-full flex justify-end">
			<ConfirmButtons
				closeText="Cancel"
				onClose={onCancel}
				confirmText={selecting ? "Add" : "Save"}
				{onConfirm}
				saveEnabled={!!selectedNode}
			/>
		</div>
	{/snippet}

	{#snippet topologyNodeSelector()}
		{#each analysisNodes as node (node.id)}
			{@const attr = node.attributes.snapshotEntity.attributes}
			<span>entity list item: {attr.displayName}</span>
		{/each}

		{#if analysisNodes.length === 0 && analysisNodesQuery.isFetched}
			<span>No topology nodes linked to this analysis</span>
		{/if}
	{/snippet}

	{#if selecting || editing}
		<div class="border rounded flex flex-col gap-2 p-2">
			{#if selecting}
				{#if selectedNode}
					{@render topologyContextEditor(selectedNode)}
				{:else}
					{@render topologyNodeSelector()}
				{/if}

				{@render confirmButtons()}
			{:else if editing}
				{#if editNode}
					{@render topologyContextEditor(editNode)}
				{/if}

				{@render confirmButtons()}
			{/if}
		</div>
	{:else}
		{#each attributes.systemContext as cx (cx.id)}
			{@const node = cx.attributes.snapshotEntityId ? analysisNodeMap.get(cx.attributes.snapshotEntityId) : undefined}
			<span>entity list item: {node?.attributes.snapshotEntity.attributes.displayName ?? "Unknown Entity"}</span>
		{/each}

		<Button
			color="primary"
			onclick={() => (selecting = true)}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Entity
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
