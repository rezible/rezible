<script lang="ts">
	import type { SystemAnalysisNode } from "$lib/api";
	import { v4 as uuidv4 } from "uuid";
	import { SvelteMap } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import Icon from "$components/common/icon/Icon.svelte";
	import { mdiPlus } from "@mdi/js";
	import ConfirmButtons from "$components/forms/confirm-buttons/ConfirmButtons.svelte";
	import { useEventDialogAttributes } from "./attributes.svelte";
	
	import { useSystemAnalysisController } from "$components/system-analysis";
	import type { TimelineEntrySystemContext, TimelineEntrySystemContextAttributes } from "../../entry-model";

	const attributes = useEventDialogAttributes();

	const analysis = useSystemAnalysisController();
	const analysisNodes = $derived(analysis.analysisNodes);
	const analysisNodeMap = $derived(new SvelteMap(analysisNodes.map((node) => [node.id, node])));
	const knowledgeEntityNodeMap = $derived(
		new SvelteMap(analysisNodes.map((node) => [node.attributes.knowledgeEntity.id, node]))
	);

	let relationship = $state<TimelineEntrySystemContextAttributes["relationship"]>("affected");

	const getAttributes = (node: SystemAnalysisNode): TimelineEntrySystemContextAttributes => ({
		systemAnalysisNodeId: node.id,
		knowledgeEntityId: node.attributes.knowledgeEntity.id,
		relationship: $state.snapshot(relationship),
	});

	const getContextNode = (cx: TimelineEntrySystemContext) =>
		(cx.attributes.systemAnalysisNodeId
			? analysisNodeMap.get(cx.attributes.systemAnalysisNodeId)
			: undefined) ?? knowledgeEntityNodeMap.get(cx.attributes.knowledgeEntityId);

	let selecting = $state(false);
	let selectedNode = $state<SystemAnalysisNode>();
	let editing = $state<TimelineEntrySystemContext>();
	const editNode = $derived(editing ? getContextNode(editing) : undefined);

	const setEditing = (cx: TimelineEntrySystemContext) => {
		editing = $state.snapshot(cx);
		relationship = $state.snapshot(cx.attributes.relationship);
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
				saveEnabled={selecting ? !!selectedNode : !!editNode}
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

		{#if analysisNodes.length === 0 && analysis.nodesQuery.isFetched}
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
			{@const node = getContextNode(cx)}
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
