<script lang="ts">
	import { BaseEdge, EdgeLabel, getBezierPath, type EdgeProps } from "@xyflow/svelte";
	import type { FlowEdge } from "../flow-graph-model";
	import { connectionLabel } from "./presentation";

	type Props = EdgeProps<FlowEdge>;
	let props: Props = $props();
	const markerId = $props.id();
	const connection = $derived(props.data!.connection);
	const label = $derived(connectionLabel(connection));
	const [path, labelX, labelY] = $derived(
		getBezierPath({
			sourceX: props.sourceX,
			sourceY: props.sourceY,
			targetX: props.targetX,
			targetY: props.targetY,
			sourcePosition: props.sourcePosition,
			targetPosition: props.targetPosition,
		})
	);
</script>

<g
	class="connection"
	class:emphasized={props.selected || props.data?.emphasized}
	class:dimmed={props.data?.dimmed}
	class:membership={connection.predicate === "contains"}
	class:aggregated={connection.relationshipIds.length > 1}
>
	<title>{label}</title>
	<defs>
		<marker
			id={markerId}
			viewBox="0 0 10 10"
			refX="9"
			refY="5"
			markerWidth="6"
			markerHeight="6"
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

{#if props.data?.showLabel}
	<EdgeLabel x={labelX} y={labelY} class="connection-label" transparent>
		<span class="bg-background text-foreground border-border rounded border px-1.5 py-0.5 text-[11px]">
			{label}
		</span>
	</EdgeLabel>
{/if}

<style>
	.connection {
		--connection-color: var(--muted-foreground);
		--connection-width: 1.5px;
	}

	.aggregated {
		--connection-width: 2.5px;
	}
	.emphasized {
		--connection-color: var(--primary);
		--connection-width: 3px;
	}
	.dimmed {
		opacity: 0.25;
	}
	.arrow {
		fill: var(--connection-color);
	}

	.connection :global(.connection-path) {
		stroke: var(--connection-color);
		stroke-width: var(--connection-width);
		stroke-linecap: round;
	}

	.membership :global(.connection-path) {
		stroke-dasharray: 5 4;
	}

	:global(.connection-label) {
		pointer-events: none !important;
	}
</style>
