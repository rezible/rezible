<script lang="ts">
	import { Handle, Position, type NodeProps } from "@xyflow/svelte";
	import { cn } from "$lib/utils";
	import type { FlowNode } from "../flow-graph-model";
	import { nodePresentationForEntity } from "./presentation";

	type Props = NodeProps<FlowNode>;
	const { data, selected }: Props = $props();

	const presentation = $derived(nodePresentationForEntity(data.entity));
</script>

<Handle
	id="input"
	type="target"
	position={Position.Left}
	class="pointer-events-none opacity-0"
	aria-hidden="true"
	tabindex={-1}
/>
<div
	data-category={data.entity.category}
	class={cn(
		"map-node bg-card text-card-foreground border-border flex h-full w-full flex-col justify-center gap-1 border px-3 py-2 shadow-sm",
		(selected || data.highlighted) && "border-primary ring-primary/30 ring-2"
	)}
>
	<span class="category truncate text-[10px] font-medium uppercase tracking-wide">
		{presentation.categoryLabel}
	</span>
	<span class="line-clamp-2 text-sm leading-tight font-medium" title={presentation.label}>
		{presentation.label}
	</span>
</div>

<Handle
	id="output"
	type="source"
	position={Position.Right}
	class="pointer-events-none opacity-0"
	aria-hidden="true"
	tabindex={-1}
/>

<style>
	.map-node {
		--category-color: var(--muted-foreground);
		border-left-color: var(--category-color);
	}

	.map-node[data-category="system_function"] {
		--category-color: var(--chart-3);
	}
	.map-node[data-category="system"] {
		--category-color: var(--chart-2);
	}
	.map-node[data-category="container"] {
		--category-color: var(--chart-1);
	}
	.map-node[data-category="infrastructure"] {
		--category-color: var(--chart-4);
	}
	.map-node[data-category="component"] {
		--category-color: var(--chart-5);
	}

	.category {
		color: var(--category-color);
	}
</style>
