<script lang="ts">
	import { type NodeProps, Handle, Position, useStore } from "@xyflow/svelte";

	import type { SystemTopologyNodeData } from "./diagramState.svelte";

	const { selected, data: arbitraryData }: NodeProps = $props();
	const data = $derived(arbitraryData as SystemTopologyNodeData);
	const entity = $derived(data.analysisNode.attributes.snapshotEntity);

	const { nodesConnectable } = useStore();
</script>

<div
	data-is-selected={selected}
	class="node border bg-surface-100 data-[is-selected=true]:bg-surface-200 rounded-lg p-3 group"
>
	<span>{entity.attributes.displayName}</span>
	{#if nodesConnectable}
		<Handle type="target" position={Position.Left} class="invisible group-hover:visible" style="width: 10px; height: 10px;" />
		<Handle type="source" position={Position.Right} class="invisible group-hover:visible" style="width: 10px; height: 10px;" />
	{/if}
</div>
