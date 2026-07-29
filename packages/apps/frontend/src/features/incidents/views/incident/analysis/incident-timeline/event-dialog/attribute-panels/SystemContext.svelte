<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		listSystemAnalysisNodesOptions,
		type IncidentTimelineEventSystemContext,
		type IncidentTimelineEventSystemContextAttributes,
		type SystemAnalysisNode,
	} from "$lib/api";
	import { v4 as uuidv4 } from "uuid";
	import { SvelteMap } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import Icon from "$components/common/icon/Icon.svelte";
	import { mdiPlus } from "@mdi/js";
	import ConfirmButtons from "$components/forms/confirm-buttons/ConfirmButtons.svelte";
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
	const analysisNodeMap = $derived(new SvelteMap(analysisNodes.map((node) => [node.id, node])));

	let relationship = $state<IncidentTimelineEventSystemContextAttributes["relationship"]>("affected");

	const getAttributes = (node: SystemAnalysisNode): IncidentTimelineEventSystemContextAttributes => ({
		systemAnalysisNodeId: node.id,
		relationship: $state.snapshot(relationship),
	});

	let selecting = $state(false);
	let selectedNode = $state<SystemAnalysisNode>();
	let editing = $state<IncidentTimelineEventSystemContext>();
	const editNode = $derived(
		editing?.attributes.systemAnalysisNodeId
			? analysisNodeMap.get(editing.attributes.systemAnalysisNodeId)
			: undefined
	);

	const setEditing = (cx: IncidentTimelineEventSystemContext) => {
		editing = $state.snapshot(cx);
		relationship = $state.snapshot(cx.attributes.relationship);
	};

	const confirmDelete = (cx: IncidentTimelineEventSystemContext) => {
		const node = analysisNodeMap.get(cx.attributes.systemAnalysisNodeId);
		editing = undefined;
		const nodeName = node?.attributes.knowledgeEntity.attributes.latestState?.displayName ?? "this entity";
		if (!node || !confirm(`Are you sure you want to remove ${nodeName}?`))
			return;
		const idx = attributes.systemContext.findIndex((c) => c.id === cx.id);
		if (idx >= 0) attributes.systemContext.splice(idx, 1);
	};

	const resetState = () => {
		selecting = false;
		selectedNode = undefined;
		editing = undefined;
		relationship = "affected";
	};

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
	{#snippet systemContextEditor(node: SystemAnalysisNode)}
		{@const attrs = node.attributes.knowledgeEntity.attributes}
		<span class="text-lg">{attrs.latestState?.displayName ?? "Unknown entity"}</span>

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
			{@const attr = node.attributes.knowledgeEntity.attributes}
			<button type="button" class="text-left" onclick={() => (selectedNode = node)}
				>{attr.latestState?.displayName ?? "Unknown entity"}</button
			>
		{/each}

		{#if analysisNodes.length === 0 && analysisNodesQuery.isFetched}
			<span>No topology nodes linked to this analysis</span>
		{/if}
	{/snippet}

	{#if selecting || editing}
		<div class="border rounded flex flex-col gap-2 p-2">
			{#if selecting}
				{#if selectedNode}
					{@render systemContextEditor(selectedNode)}
				{:else}
					{@render topologyNodeSelector()}
				{/if}

				{@render confirmButtons()}
			{:else if editing}
				{#if editNode}
					{@render systemContextEditor(editNode)}
				{/if}

				{@render confirmButtons()}
			{/if}
		</div>
	{:else}
		{#each attributes.systemContext as cx (cx.id)}
			{@const node = analysisNodeMap.get(cx.attributes.systemAnalysisNodeId)}
			<button type="button" class="text-left" onclick={() => setEditing(cx)}>
				{node?.attributes.knowledgeEntity.attributes.latestState?.displayName ?? "Unknown Entity"}
			</button>
		{/each}

		<Button color="primary" onclick={() => (selecting = true)}>
			<span class="flex items-center gap-2 text-primary-content">
				Add Entity
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
