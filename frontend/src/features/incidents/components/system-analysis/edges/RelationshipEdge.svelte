<script lang="ts">
    import type { SystemAnalysisRelationshipAttributes } from '$lib/api';
	import { useEdges, type EdgeProps, getBezierPath, BaseEdge, type Edge, EdgeLabelRenderer, type GetBezierPathParams } from '@xyflow/svelte';
   
	const {
		data: arbitraryData,
		target,
		source,
		sourceX,
		sourceY,
		labelStyle,
		sourcePosition,
		targetX,
		targetY,
		targetPosition,
		markerEnd,
	}: EdgeProps = $props();
	const data = $derived(arbitraryData as SystemAnalysisRelationshipAttributes);

	const edges = useEdges();

	const labelX = $derived((sourceX + targetX) / 2);
	const labelY = $derived((sourceY + targetY) / 2);
   
	const offset = 1;
	const [path1] = $derived(getBezierPath({
		sourcePosition,
		targetPosition,
		sourceX: sourceX - offset,
		sourceY: sourceY + offset,
		targetX: targetX + offset,
		targetY: targetY + offset,
	}));
	const [path2] = $derived(getBezierPath({
		sourcePosition,
		targetPosition,
		sourceX: sourceX - offset,
		sourceY: sourceY - offset,
		targetX: targetX + offset,
		targetY: targetY - offset,
	}));
</script>

<BaseEdge path={path1} {markerEnd} class="" style="" interactionWidth={40} />
<BaseEdge path={path2} {markerEnd} class="" interactionWidth={40} />

<EdgeLabelRenderer>
	<div class="nodrag nopan relationship-label" style="transform: translate(-50%, -50%) translate({labelX}px, {labelY}px)">
		{data.description}
	</div>
</EdgeLabelRenderer>

<style lang="postcss">
	.relationship-label {
		@apply absolute p-2 rounded-lg border border-surface-200 text-sm;
	}
</style>