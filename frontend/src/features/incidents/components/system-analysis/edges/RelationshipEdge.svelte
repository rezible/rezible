<script lang="ts">
    import type { SystemAnalysisRelationshipAttributes } from '$lib/api';
	import { useEdges, type EdgeProps, getBezierPath, BaseEdge, type Edge, EdgeLabelRenderer } from '@xyflow/svelte';
   
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

	const isBidirectional = (e: Edge) => {
		return (e.source === target && e.target === source) || (e.target === source && e.source === target)
	}
	let isBidirectionalEdge = $state($edges.some(isBidirectional));
	edges.subscribe(e => {
		isBidirectionalEdge = e.some(isBidirectional);
	});
   
	const [path, labelX, labelY] = $derived.by(() => {
		const edgePathParams = {
			sourceX,
			sourceY,
			sourcePosition,
			targetX,
			targetY,
			targetPosition
		};

		// if (isBidirectionalEdge) return getSpecialPath(edgePathParams, sourceX < targetX ? 25 : -25);
		return getBezierPath(edgePathParams);
	})
   
	type SpecialPathProps = {
		sourceX: number; sourceY: number; targetX: number; targetY: number
	}
	const getSpecialPath = ({ sourceX, sourceY, targetX, targetY}: SpecialPathProps, offset: number): [string, number, number] => {
		const centerX = (sourceX + targetX) / 2;
		const centerY = (sourceY + targetY) / 2;
		const loopPath = `M ${sourceX} ${sourceY} Q ${centerX} ${centerY + offset} ${targetX} ${targetY}`;
		return [loopPath, 0, 0];
	}
</script>

<BaseEdge {path} {markerEnd} class="animated" />

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