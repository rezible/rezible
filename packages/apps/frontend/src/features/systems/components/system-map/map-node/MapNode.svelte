<script lang="ts">
	import { Handle, Position, type NodeProps } from "@xyflow/svelte";

	import { Badge } from "$components/ui/badge";
	import type { FlowNode } from "../flow-model";
	import { COMPACT_NODE_SIZE, GROUP_NODE_MIN_SIZE } from "../layout";
	import { nodePresentationForEntity } from "./presentation";

	type Props = NodeProps<FlowNode>;
	let { data, selected }: Props = $props();

	const entity = $derived(data.entity);
	const category = $derived(entity.category);
	const isGroup = $derived(data.appearance === "group");
	const nodePresentation = $derived(nodePresentationForEntity(entity));
	const nodeSize = $derived(isGroup ? GROUP_NODE_MIN_SIZE : COMPACT_NODE_SIZE);
	const nodeStyle = $derived(`min-width: ${nodeSize.width}px; min-height: ${nodeSize.height}px`);
	const contextLabel = $derived(
		`${data.annotationCount} context annotation${data.annotationCount === 1 ? "" : "s"}`
	);
</script>

<Handle id="target-left" type="target" position={Position.Left} class="pointer-events-none opacity-0" />
<Handle id="target-right" type="target" position={Position.Right} class="pointer-events-none opacity-0" />
<Handle id="target-top" type="target" position={Position.Top} class="pointer-events-none opacity-0" />
<Handle id="target-bottom" type="target" position={Position.Bottom} class="pointer-events-none opacity-0" />

{#if isGroup}
	<div
		data-appearance="group"
		data-selected={selected}
		data-connection-endpoint={data.isConnectionEndpoint ?? false}
		data-category={category}
		style={nodeStyle}
		class="map-node-shell border-border bg-card/70 text-card-foreground flex h-full w-full flex-col gap-2 rounded-lg border p-3 shadow-sm data-[selected=true]:border-primary data-[connection-endpoint=true]:ring-1 data-[connection-endpoint=true]:ring-primary/60 data-[connection-endpoint=true]:ring-offset-1"
		role="button"
		tabindex="0"
		aria-label={`${nodePresentation.label}, group boundary`}
	>
		<div class="flex min-w-0 items-center justify-between gap-2">
			<span class="truncate text-sm font-semibold" title={nodePresentation.label}>
				{nodePresentation.label}
			</span>
			<Badge variant="outline">{nodePresentation.categoryLabel}</Badge>
		</div>

		{#if data.annotationCount > 0}
			<Badge variant="secondary" aria-label={contextLabel} data-context-count={data.annotationCount}>
				{data.annotationCount} context
			</Badge>
		{/if}
	</div>
{:else}
	<div
		data-appearance="compact"
		data-selected={selected}
		data-connection-endpoint={data.isConnectionEndpoint ?? false}
		data-category={category}
		style={nodeStyle}
		class="map-node-shell border-border bg-card text-card-foreground flex h-full w-full flex-col justify-center gap-1 rounded-md border px-3 py-2 shadow-sm data-[selected=true]:border-primary data-[selected=true]:bg-muted data-[connection-endpoint=true]:ring-1 data-[connection-endpoint=true]:ring-primary/60 data-[connection-endpoint=true]:ring-offset-1"
		role="button"
		tabindex="0"
		aria-label={nodePresentation.label}
	>
		<span
			class="line-clamp-2 break-words text-sm font-semibold leading-tight"
			title={nodePresentation.label}
		>
			{nodePresentation.label}
		</span>
		<div class="flex min-w-0 items-center gap-2">
			<Badge variant={nodePresentation.isActor ? "secondary" : "outline"}>
				{nodePresentation.categoryLabel}
			</Badge>
			{#if nodePresentation.kindLabel}
				<span class="text-muted-foreground min-w-0 truncate text-xs" title={nodePresentation.kindLabel}>
					{nodePresentation.kindLabel}
				</span>
			{/if}
			{#if data.annotationCount > 0}
				<Badge
					variant="secondary"
					aria-label={contextLabel}
					data-context-count={data.annotationCount}
				>
					{data.annotationCount}
				</Badge>
			{/if}
		</div>
	</div>
{/if}

<Handle id="source-left" type="source" position={Position.Left} class="pointer-events-none opacity-0" />
<Handle id="source-right" type="source" position={Position.Right} class="pointer-events-none opacity-0" />
<Handle id="source-top" type="source" position={Position.Top} class="pointer-events-none opacity-0" />
<Handle id="source-bottom" type="source" position={Position.Bottom} class="pointer-events-none opacity-0" />
