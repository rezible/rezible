import type { MapConnection } from "$features/system/lib/system-map/presentation";
import { relationshipPredicateLabel } from "$features/system/lib/system-map/labels";
import type { FlowEdge } from "../flow-model";

export type ConnectionSelection = {
	relationshipIds: ReadonlySet<string>;
	endpointIds: ReadonlySet<string>;
};

export type ConnectionPresentationState = {
	selected: boolean;
	hovered: boolean;
	highlighted: boolean;
	dimmed: boolean;
	endpointEmphasized: boolean;
	labelVisible: boolean;
};

export type ConnectionPresentationOptions = {
	selection?: ConnectionSelection;
	highlightedRelationshipIds?: ReadonlySet<string>;
	highlightedEndpointIds?: ReadonlySet<string>;
	hoveredConnectionId?: string;
	showAllConnectionLabels?: boolean;
};

/** Keeps route crossfades weighted while lifting the whole hovered connection to full visibility. */
export const connectionRouteOpacities = (
	routeOpacities: readonly number[],
	hovered: boolean,
	dimmed: boolean
): number[] => {
	const totalOpacity = routeOpacities.reduce((total, opacity) => total + opacity, 0);
	if (hovered && totalOpacity <= 0) return routeOpacities.map((_, index) => (index === 0 ? 1 : 0));
	const hoverScale = hovered ? 1 / totalOpacity : 1;
	const dimScale = dimmed && !hovered ? 0.25 : 1;
	return routeOpacities.map((opacity) => opacity * hoverScale * dimScale);
};

export const connectionLabelOpacity = (opacity: number, hovered: boolean, dimmed: boolean): number =>
	hovered ? 1 : opacity * (dimmed ? 0.35 : 1);

export const connectionEndpointOpacity = (opacity: number, hovered: boolean): number =>
	hovered ? 1 : opacity;

type ConnectionEdge = Pick<FlowEdge, "id" | "source" | "target" | "data">;

/** Formats a readable edge label; summary counts derive only from canonical contributor IDs. */
export const connectionLabel = (connection: MapConnection): string => {
	const predicate = relationshipPredicateLabel(connection.predicate);
	return connection.classification === "summary"
		? `${predicate} · ${connection.sourceRelationshipIds.length}`
		: predicate;
};

/** Derives emphasis from canonical relationship and entity IDs, including regrouped summaries. */
export const connectionPresentationState = (
	edge: ConnectionEdge,
	{
		selection,
		highlightedRelationshipIds = new Set(),
		highlightedEndpointIds = new Set(),
		hoveredConnectionId,
		showAllConnectionLabels = false,
	}: ConnectionPresentationOptions = {}
): ConnectionPresentationState => {
	const relationshipIds = edge.data?.connection.sourceRelationshipIds ?? [];
	const selected = relationshipIds.some((relationshipId) => selection?.relationshipIds.has(relationshipId));
	const hovered = hoveredConnectionId === edge.id;
	const hostRelationshipHighlighted = relationshipIds.some((relationshipId) =>
		highlightedRelationshipIds.has(relationshipId)
	);
	const connectedToSelection =
		selection?.endpointIds.has(edge.source) || selection?.endpointIds.has(edge.target) || false;
	const connectedToHighlight =
		highlightedEndpointIds.has(edge.source) || highlightedEndpointIds.has(edge.target);
	const highlighted =
		selected || hovered || hostRelationshipHighlighted || connectedToSelection || connectedToHighlight;
	const hasPresentationContext = Boolean(
		selection?.relationshipIds.size ||
		selection?.endpointIds.size ||
		highlightedRelationshipIds.size ||
		highlightedEndpointIds.size ||
		hoveredConnectionId
	);

	return {
		selected,
		hovered,
		highlighted,
		dimmed: hasPresentationContext && !highlighted,
		endpointEmphasized: selected || hovered || hostRelationshipHighlighted,
		labelVisible: showAllConnectionLabels || selected || hovered || hostRelationshipHighlighted,
	};
};

/** Returns displayed endpoints for selected or hovered source relationships. */
export const connectionEndpointIds = (
	edges: readonly ConnectionEdge[],
	options: ConnectionPresentationOptions
): ReadonlySet<string> => {
	const endpointIds = new Set<string>();
	for (const edge of edges) {
		if (!connectionPresentationState(edge, options).endpointEmphasized) continue;
		endpointIds.add(edge.source);
		endpointIds.add(edge.target);
	}
	return endpointIds;
};
