<script lang="ts">
	import { BaseEdge, EdgeLabel, getBezierPath, Position, type EdgeProps } from "@xyflow/svelte";
	import type { BoxSide } from "$features/system/lib/system-map/geometry";
	import type { FlowEdge } from "../projection";

	type Props = EdgeProps<FlowEdge>;
	let props: Props = $props();
	const markerId = $props.id();

	const sidePositions: Record<BoxSide, Position> = {
		top: Position.Top,
		right: Position.Right,
		bottom: Position.Bottom,
		left: Position.Left,
	};

	const data = $derived(props.data!);
	const [path, labelX, labelY] = $derived(
		getBezierPath({
			sourceX: data.anchors.source.x,
			sourceY: data.anchors.source.y,
			targetX: data.anchors.target.x,
			targetY: data.anchors.target.y,
			sourcePosition: sidePositions[data.anchors.source.side],
			targetPosition: sidePositions[data.anchors.target.side],
		})
	);
</script>

<g
	class="connection"
	class:emphasized={props.selected || data.emphasized}
	class:dimmed={data.dimmed}
	class:placement={data.connection.placement}
>
	<title>{data.label}</title>
	<defs>
		<marker
			id={markerId}
			viewBox="0 0 10 10"
			refX="9"
			refY="5"
			markerWidth="5"
			markerHeight="5"
			orient="auto"
		>
			<path class="arrow" d="M 0 0 L 10 5 L 0 10 z" />
		</marker>
	</defs>
	<BaseEdge
		id={props.id}
		{path}
		markerEnd={`url(#${markerId})`}
		interactionWidth={20}
		class="connection-path"
	/>
</g>

{#if data.showLabel}
	<EdgeLabel x={labelX} y={labelY} class="connection-label" transparent>
		<!-- An opposed pair's labels shift apart along their separation so they sit side by side. -->
		<span
			class="chip bg-background text-foreground border-border block border whitespace-nowrap"
			class:opacity-25={data.dimmed}
			style:--separation-x={data.anchors.separation.x}
			style:--separation-y={data.anchors.separation.y}
		>
			{data.label}
		</span>
	</EdgeLabel>
{/if}

<style>
	/* Strokes are constant on screen at every zoom. */
	.connection {
		--connection-color: var(--muted-foreground);
		--connection-width: calc(1px / var(--map-zoom, 1));
		--connection-opacity: 0.5;
	}

	.emphasized {
		--connection-color: var(--foreground);
		--connection-width: calc(1.5px / var(--map-zoom, 1));
		--connection-opacity: 1;
	}
	.placement :global(.connection-path) {
		stroke-dasharray: calc(5px / var(--map-zoom, 1)) calc(4px / var(--map-zoom, 1));
	}
	.dimmed {
		opacity: 0.25;
	}
	.arrow {
		fill: var(--connection-color);
		fill-opacity: var(--connection-opacity);
	}

	.connection :global(.connection-path) {
		stroke: var(--connection-color);
		stroke-opacity: var(--connection-opacity);
		stroke-width: var(--connection-width);
		stroke-linecap: round;
		fill: none;
	}

	:global(.connection-label) {
		pointer-events: none !important;
	}

	/* Constant on screen at every zoom. */
	.chip {
		font-size: calc(11px / var(--map-zoom, 1));
		padding: 0.15em 0.5em;
		border-width: calc(1px / var(--map-zoom, 1));
		border-radius: 0.35em;
		translate: calc(var(--separation-x) * (50% + 0.2em)) calc(var(--separation-y) * (50% + 0.2em));
	}
</style>
