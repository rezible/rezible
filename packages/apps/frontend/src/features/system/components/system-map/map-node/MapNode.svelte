<script lang="ts">
	import { Handle, Position, type NodeProps } from "@xyflow/svelte";

	import { cn } from "$lib/utils";
	import { getMapCategoryDisplay } from "$features/system/lib/system-map/category";
	import type { FlowNode } from "../flow-model";
	import { GROUP_HEADER_LAYOUT, nodeMinSizeForCategory } from "../layout";
	import { nodePresentationForEntity } from "./presentation";

	type Props = NodeProps<FlowNode>;
	let { data, selected }: Props = $props();

	const entity = $derived(data.entity);
	const category = $derived(entity.category);
	const isGroup = $derived(data.appearance === "group");
	const nodePresentation = $derived(nodePresentationForEntity(entity));
	const nodeSize = $derived(nodeMinSizeForCategory(category, isGroup ? "group" : "compact"));
	const level = $derived(getMapCategoryDisplay(category).level ?? 1);
	const layerStyle = $derived(
		[
			"border-amber-500/50 bg-amber-500/5",
			"border-sky-500/50 bg-sky-500/5",
			"border-violet-500/50 bg-violet-500/5",
			"border-teal-500/50 bg-teal-500/5",
		][level]
	);
	const layerDotStyle = $derived(["bg-amber-500", "bg-sky-500", "bg-violet-500", "bg-teal-500"][level]);
	const nodeStyle = $derived(
		`min-width: ${nodeSize.width}px; min-height: ${nodeSize.height}px;${
			isGroup ? ` padding: ${GROUP_HEADER_LAYOUT.contentPadding}px;` : ""
		}`
	);
	const categoryRowStyle = `height: ${GROUP_HEADER_LAYOUT.categoryHeight}px; line-height: ${GROUP_HEADER_LAYOUT.categoryHeight}px`;
	const titleStyle = $derived(
		`height: ${GROUP_HEADER_LAYOUT.titleHeight}px; line-height: ${GROUP_HEADER_LAYOUT.titleLineHeight}px;${
			isGroup ? ` margin-top: ${GROUP_HEADER_LAYOUT.titleGap}px;` : ""
		}`
	);
</script>

<!-- Svelte Flow needs one handle of each type for its custom edge model; route paths use fixed world geometry. -->
<Handle type="target" position={Position.Left} class="pointer-events-none opacity-0" />

<div
	data-appearance={isGroup ? "group" : "compact"}
	data-selected={selected}
	data-highlighted={data.isHighlighted ?? false}
	data-connection-endpoint={data.isConnectionEndpoint ?? false}
	data-draggable={data.isDraggable ?? false}
	data-category={category}
	style={nodeStyle}
	class={cn(
		"map-node-shell text-card-foreground flex h-full w-full flex-col rounded-lg border shadow-sm data-[selected=true]:border-primary data-[highlighted=true]:ring-2 data-[highlighted=true]:ring-primary/50 data-[connection-endpoint=true]:ring-1 data-[connection-endpoint=true]:ring-primary/60 data-[connection-endpoint=true]:ring-offset-1",
		layerStyle,
		isGroup
			? "p-0"
			: "px-3 py-2 data-[draggable=true]:cursor-grab data-[draggable=true]:active:cursor-grabbing"
	)}
	role="button"
	tabindex={data.isInteractive === false ? -1 : 0}
	aria-hidden={data.isInteractive === false}
	aria-label={`${nodePresentation.categoryLabel}: ${nodePresentation.label}${isGroup ? ", group boundary" : ""}`}
>
	<div
		class="text-muted-foreground flex min-w-0 items-center justify-between gap-2 text-[10px] font-semibold uppercase tracking-wide"
		style={categoryRowStyle}
	>
		<span class="flex min-w-0 items-center gap-1.5 truncate">
			<span class={`size-1.5 shrink-0 rounded-full ${layerDotStyle}`}></span>
			<span class="truncate" title={nodePresentation.categoryLabel}>
				{nodePresentation.categoryLabel}
			</span>
		</span>
	</div>
	<div class={`flex min-w-0 items-center ${isGroup ? "" : "my-auto"}`} style={titleStyle}>
		<span class="line-clamp-2 min-w-0 break-words text-sm font-semibold" title={nodePresentation.label}>
			{nodePresentation.label}
		</span>
	</div>
</div>

<Handle type="source" position={Position.Right} class="pointer-events-none opacity-0" />
