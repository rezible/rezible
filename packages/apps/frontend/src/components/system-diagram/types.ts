import type { Edge, Node, XYPosition } from "@xyflow/svelte";
import type { KnowledgeGraphEntity, KnowledgeGraphRelationship } from "$lib/api";

export type SystemDiagramNode = Node<{
	entity: KnowledgeGraphEntity;
	attachmentCount?: number;
	highlighted?: boolean;
}>;

export type SystemDiagramEdge = Edge<{
	relationship: KnowledgeGraphRelationship;
	attachmentCount?: number;
	highlighted?: boolean;
}>;

export type GraphSelection = {
	nodeId?: string;
	edgeId?: string;
};

export type GraphHighlights = {
	nodeIds: string[];
	edgeIds: string[];
};

export type DiagramContextMenu = {
	selection: GraphSelection;
	position: XYPosition;
};
