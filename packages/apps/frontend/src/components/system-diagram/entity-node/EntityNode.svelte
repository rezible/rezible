<script lang="ts">
	import { Handle, Position, type NodeProps } from "@xyflow/svelte";
	import { Badge } from "$components/ui/badge";
	import type { SystemDiagramNode } from "../types";

	type Props = NodeProps<SystemDiagramNode>;
	let { data, selected }: Props = $props();
	
	const entity = $derived(data.entity);
	const attrs = $derived(entity.attributes);
	const state = $derived(attrs.latestState);
	const aliasRef = $derived(attrs.aliases[0]?.attributes.resourceRef.resourceRef);
	const label = $derived(state?.displayName || aliasRef || attrs.kind);
</script>

<Handle type="target" position={Position.Left} />

<div
	data-selected={selected || data.highlighted}
	class="border-border bg-card text-card-foreground data-[selected=true]:border-primary data-[selected=true]:bg-muted flex w-56 flex-col gap-1 border px-3 py-2 shadow-sm"
>
	<div class="flex items-center justify-between gap-2">
		<Badge variant="outline">
			{attrs.kind.replaceAll("_", " ")}
		</Badge>
		{#if data.attachmentCount}
			<Badge aria-label={`${data.attachmentCount} attached entries`}>
				{data.attachmentCount}
			</Badge>
		{/if}
	</div>
	
	<span class="truncate text-sm font-medium">
		{label}
	</span>

	{#if state?.description}
		<span class="text-muted-foreground line-clamp-2 text-xs">
			{state.description}
		</span>
	{/if}
</div>

<Handle type="source" position={Position.Right} />
