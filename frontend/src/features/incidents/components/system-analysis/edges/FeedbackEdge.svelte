<script lang="ts">
	import { useEdges, type EdgeProps, getBezierPath, BaseEdge, type Edge } from '@xyflow/svelte';
   
	const {
		target,
		source,
		sourceX,
		sourceY,
		sourcePosition,
		targetX,
		targetY,
		targetPosition,
		markerEnd,
	}: EdgeProps = $props();
   
	const edges = useEdges();
   
	const isBidirectional = (e: Edge) => {
		return (e.source === target && e.target === source) || (e.target === source && e.source === target)
	}
	const isBidirectionalEdge = $derived($edges.some(isBidirectional))
   
	const path = $derived.by(() => {
		const edgePathParams = {
			sourceX,
			sourceY,
			sourcePosition,
			targetX,
			targetY,
			targetPosition
		};

		if (isBidirectionalEdge) {
			return getSpecialPath(edgePathParams, sourceX < targetX ? 25 : -25);
		}
		const bezPath = getBezierPath(edgePathParams);
		return bezPath[0];
	})
   
	type SpecialPathProps = {
		sourceX: number; sourceY: number; targetX: number; targetY: number
	}
	const getSpecialPath = ({ sourceX, sourceY, targetX, targetY}: SpecialPathProps, offset: number) => {
	  const centerX = (sourceX + targetX) / 2;
	  const centerY = (sourceY + targetY) / 2;
	  return `M ${sourceX} ${sourceY} Q ${centerX} ${centerY + offset} ${targetX} ${targetY}`;
	}
</script>

<BaseEdge {path} {markerEnd} />