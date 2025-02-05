<script lang="ts">
	import { NodeToolbar, type NodeProps, Handle, Position, useNodes } from "@xyflow/svelte";
	import type { SystemComponentNodeData } from "../diagram.svelte";

	const { selected, data: arbitraryData }: NodeProps = $props();
	const data = $derived(arbitraryData as SystemComponentNodeData);
</script>

<div
	data-is-selected={selected}
	class="node border bg-surface-100 data-[is-selected=true]:bg-surface-200 rounded-lg p-3 group"
>
	<span>{data.component.attributes.component.attributes.name}</span>
	<Handle type="target" position={Position.Left} class="invisible group-hover:visible" />
	<Handle type="source" position={Position.Right} class="invisible group-hover:visible" />
</div>
