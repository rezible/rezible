<script lang="ts">
    import type { SystemAnalysisRelationshipAttributes } from '$lib/api';
	import { type EdgeProps, EdgeLabelRenderer, getSmoothStepPath } from '@xyflow/svelte';
   
	const props: EdgeProps = $props();
	const data = $derived(props.data as SystemAnalysisRelationshipAttributes);

	const centerX = $derived((props.sourceX + props.targetX) / 2);
	const centerY = $derived((props.sourceY + props.targetY) / 2);

	const offset = 5;
	const sourceX = $derived(props.sourceX - offset);
	const targetX = $derived(props.targetX + offset);

	const centerXOffset = $derived(props.sourceY > props.targetY ? offset : (offset * -1));

	const [pathOut] = $derived(getSmoothStepPath({
		sourcePosition: props.sourcePosition,
		targetPosition: props.targetPosition,
		targetX, sourceX,
		centerX: centerX + centerXOffset,
		sourceY: props.sourceY + offset,
		targetY: props.targetY + offset,
	}));
	const [pathIn] = $derived(getSmoothStepPath({
		sourcePosition: props.sourcePosition,
		targetPosition: props.targetPosition,
		targetX, sourceX,
		centerX: centerX - centerXOffset,
		targetY: props.targetY - offset,
		sourceY: props.sourceY - offset,
	}));

	const animated = $derived(props.selected);
	const animatedPathProps = {"stroke-width": "5", "stroke-dasharray": "10", "stroke-dashoffset": "1", "stroke-miterlimit": "10"}
	const pathStrokeProps = $derived(animated ? animatedPathProps : {});

	const interactionWidth = 40;

	const labelTransformStyle = $derived(`transform: translate(-50%, -50%) translate(${centerX}px, ${centerY}px);`);
</script>

{#snippet edgePath(d: string, dir: "out" | "in")}
	<path id={props.id} {d} fill="none" style="" class="svelte-flow__edge-path"
		marker-start={props.markerStart}
		marker-end={props.markerEnd}
		{...pathStrokeProps}>
		<animate attributeName="stroke-dashoffset" values={dir === "out" ? "0;100" : "100;0"} dur="5s" calcMode="linear" repeatCount="indefinite" />
	</path>

	<path {d} stroke-opacity={0} stroke-width={interactionWidth} fill="none" class="svelte-flow__edge-interaction" />
{/snippet}

{@render edgePath(pathOut, "out")}
{@render edgePath(pathIn, "in")}

<EdgeLabelRenderer>
	<div class="nodrag nopan relationship-label flex flex-col gap-2" style={labelTransformStyle}>
		<span class="">{data.description}relationship</span>
	</div>
</EdgeLabelRenderer>

<style lang="postcss">
	.relationship-label {
		@apply absolute p-2 rounded-lg border border-surface-200 text-sm;
	}
</style>