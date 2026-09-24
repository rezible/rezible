<script lang="ts">
	import { BaseEdge, type EdgeProps } from "@xyflow/svelte";

	import { cn } from "$lib/utils";
	import { connectionRouteToSvgPath } from "./geometry";
	import { connectionLabel, connectionLabelOpacity, connectionRouteOpacities } from "./presentation";
	import type { ConnectionRouteVariant, FlowEdge } from "../flow-model";

	type Props = EdgeProps<FlowEdge>;
	let props: Props = $props();

	const data = $derived(props.data);
	const connection = $derived(data?.connection);
	const route = $derived(data!.route);
	const routeLayers = $derived<readonly ConnectionRouteVariant[]>(
		data?.routeVariants?.length ? data.routeVariants : [{ route, opacity: data?.opacity ?? 1 }]
	);
	const primaryRouteIndex = $derived(
		routeLayers.reduce(
			(bestIndex, candidate, index) =>
				candidate.opacity > routeLayers[bestIndex].opacity ? index : bestIndex,
			0
		)
	);
	const primaryRoute = $derived(routeLayers[primaryRouteIndex] ?? { route, opacity: data?.opacity ?? 1 });
	const count = $derived(connection?.sourceRelationshipIds.length ?? 1);
	const isMembership = $derived(connection?.predicate === "contains");
	const isSummary = $derived(connection?.classification === "summary");
	const isHighlighted = $derived(Boolean(props.selected || data?.isHighlighted));
	const isHovered = $derived(Boolean(data?.isHovered));
	const isDimmed = $derived(Boolean(data?.isDimmed && !isHighlighted));
	const isInteractive = $derived((props.interactionWidth ?? 24) > 0);
	const routeOpacities = $derived(
		connectionRouteOpacities(
			routeLayers.map((layer) => layer.opacity),
			isHovered,
			isDimmed
		)
	);
	const renderedRoutes = $derived(
		routeLayers.map((layer, index) => ({
			...layer,
			opacity: routeOpacities[index] ?? layer.opacity,
			path: connectionRouteToSvgPath(layer.route),
			isPrimary: index === primaryRouteIndex,
		}))
	);
	const lineWidth = $derived(
		isMembership
			? Math.min(2.4, 1 + Math.log2(count + 1) * 0.7)
			: Math.min(5, 1.5 + Math.log2(count + 1) * 1.2) + (isHighlighted ? 0.6 : 0)
	);
	const edgeStyle = (routeOpacity: number, primary: boolean) =>
		[
			`stroke-width: ${lineWidth}px`,
			"stroke-linecap: round",
			"stroke-linejoin: round",
			`opacity: ${routeOpacity}`,
			!primary || !isInteractive ? "pointer-events: none" : "",
			isSummary ? "stroke-dasharray: 7 4" : "",
		]
			.filter(Boolean)
			.join(";");
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
			`opacity: ${connectionLabelOpacity(primaryRoute.opacity, isHovered, isDimmed)}`,
		].join(";")
	);
	const label = $derived(
		connection && (props.selected || data?.isLabelVisible) ? connectionLabel(connection) : undefined
	);
</script>

<title>{props.ariaLabel}</title>

{#each renderedRoutes as routeLayer, index (`${props.id}-${index}`)}
	<BaseEdge
		id={routeLayer.isPrimary ? props.id : `${props.id}--route-${index}`}
		path={routeLayer.path}
		markerEnd={props.markerEnd}
		interactionWidth={routeLayer.isPrimary ? (props.interactionWidth ?? 24) : 0}
		style={edgeStyle(routeLayer.opacity, routeLayer.isPrimary)}
		label={routeLayer.isPrimary ? label : undefined}
		labelX={primaryRoute.route.labelPosition.x}
		labelY={primaryRoute.route.labelPosition.y}
		{labelStyle}
		aria-label={routeLayer.isPrimary ? props.ariaLabel : undefined}
		data-connection-id={props.id}
		class={cn(
			isHighlighted
				? "stroke-primary"
				: isSummary || isMembership
					? "stroke-muted-foreground"
					: "stroke-foreground"
		)}
	/>
{/each}
