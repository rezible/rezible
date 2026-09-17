<script lang="ts">
	import { BaseEdge, type EdgeProps } from "@xyflow/svelte";

	import { cn } from "$lib/utils";
	import { getSystemMapConnectionPath } from "./geometry";
	import type { FlowEdge } from "../flow-model";

	type Props = EdgeProps<FlowEdge>;
	let props: Props = $props();

	const data = $derived(props.data);
	const connection = $derived(data?.connection);
	const laneOffset = $derived(data?.laneOffset ?? 0);
	const path = $derived(getSystemMapConnectionPath({
		sourceX: props.sourceX,
		sourceY: props.sourceY,
		sourcePosition: props.sourcePosition,
		targetX: props.targetX,
		targetY: props.targetY,
		targetPosition: props.targetPosition,
		laneOffset,
	}));
	const count = $derived(connection?.sourceRelationshipIds.length ?? 1);
	const isMembership = $derived(connection?.predicate === "contains");
	const isSummary = $derived(connection?.classification === "summary");
	const isHighlighted = $derived(Boolean(props.selected || data?.isHighlighted));
	const isDimmed = $derived(Boolean(data?.isDimmed && !isHighlighted));
	const lineWidth = $derived(
		isMembership
			? Math.min(2.4, 1 + Math.log2(count + 1) * 0.7)
			: Math.min(5, 1.5 + Math.log2(count + 1) * 1.2) + (isHighlighted ? 0.6 : 0)
	);
	const edgeStyle = $derived(
		[
			`stroke-width: ${lineWidth}px`,
			"stroke-linecap: round",
			"stroke-linejoin: round",
			isSummary ? "stroke-dasharray: 7 4" : "",
		].filter(Boolean).join(";")
	);
	const labelStyle = $derived(
		[
			"font-size: 11px",
			"font-weight: 600",
			`color: ${isHighlighted ? "var(--primary)" : isMembership ? "var(--muted-foreground)" : "var(--foreground)"}`,
			"background: var(--background)",
			"border: 1px solid var(--border)",
			"border-radius: 4px",
			"box-shadow: 0 1px 2px color-mix(in srgb, var(--foreground) 12%, transparent)",
			"line-height: 1.1",
			"max-width: 180px",
			"overflow-wrap: anywhere",
			"padding: 2px 5px",
			"text-align: center",
			"white-space: normal",
			`opacity: ${isDimmed ? 0.35 : 1}`,
		].join(";")
	);
	const label = $derived(
		connection ? `${connection.predicate.replaceAll("_", " ")} · ${count}` : undefined
	);
</script>

<title>{props.ariaLabel}</title>

<BaseEdge
	id={props.id}
	path={path[0]}
	markerEnd={props.markerEnd}
	interactionWidth={props.interactionWidth ?? 24}
	style={edgeStyle}
	{label}
	labelX={path[1]}
	labelY={path[2]}
	labelStyle={labelStyle}
	aria-label={props.ariaLabel}
	data-connection-id={props.id}
	class={cn(
		isHighlighted ? "stroke-primary" : isSummary || isMembership ? "stroke-muted-foreground" : "stroke-foreground",
		isDimmed && "opacity-25"
	)}
/>
