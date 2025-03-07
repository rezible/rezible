<script lang="ts">
	import { NodeToolbar, type NodeProps, Handle, Position, useStore } from "@xyflow/svelte";
	import type { SystemComponentNodeData } from "$features/incidents/components/incident-analysis/system-diagram/diagram.svelte";
	import { fromStore } from "svelte/store";

	const { selected, data: arbitraryData }: NodeProps = $props();
	const data = $derived(arbitraryData as SystemComponentNodeData);
	const component = $derived(data.analysisComponent.attributes.component);

	const { nodesConnectable } = useStore();
	const showHandles = $derived(fromStore(nodesConnectable).current);
</script>

<div
	data-is-selected={selected}
	class="node border bg-surface-100 data-[is-selected=true]:bg-surface-200 rounded-lg p-3 group"
>
	<span>{component.attributes.name}</span>
	{#if showHandles}
		<Handle type="target" position={Position.Left} class="invisible group-hover:visible" style="width: 10px; height: 10px;" />
		<Handle type="source" position={Position.Right} class="invisible group-hover:visible" style="width: 10px; height: 10px;" />
	{/if}
</div>
