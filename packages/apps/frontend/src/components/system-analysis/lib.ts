import type { SystemAnalysisEdge, SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import type { SystemDiagramNode, SystemDiagramEdge } from "$components/system-diagram";
import type { ResponsePage } from "$lib/api/utils";

export function flattenPages<T extends { id: string }>(pages: ResponsePage<T>[] | undefined) {
	const items = new Map<string, T>();
	for (const page of pages ?? []) {
		for (const item of page.data) {
			items.set(item.id, item);
		}
	}
	return [...items.values()];
}

export type EntryAttachments = {
	byNodeId: Map<string, SystemAnalysisEntry[]>;
	byEdgeId: Map<string, SystemAnalysisEntry[]>;
	unattached: SystemAnalysisEntry[];
};

export function mapEntryAttachments(
	nodes: SystemAnalysisNode[],
	edges: SystemAnalysisEdge[],
	entries: SystemAnalysisEntry[]
): EntryAttachments {
	const nodeIdByEntityId = new Map(nodes.map((node) => [node.attributes.knowledgeEntity.id, node.id]));
	const edgeIdByRelationshipId = new Map(
		edges.map((edge) => [edge.attributes.knowledgeRelationship.id, edge.id])
	);
	const byNodeId = new Map<string, SystemAnalysisEntry[]>();
	const byEdgeId = new Map<string, SystemAnalysisEntry[]>();
	const unattached: SystemAnalysisEntry[] = [];

	for (const entry of entries) {
		const nodeIds = new Set<string>();
		const edgeIds = new Set<string>();
		for (const subject of entry.attributes.subjects) {
			const nodeId = subject.attributes.knowledgeEntityId
				? nodeIdByEntityId.get(subject.attributes.knowledgeEntityId)
				: undefined;
			const edgeId = subject.attributes.knowledgeRelationshipId
				? edgeIdByRelationshipId.get(subject.attributes.knowledgeRelationshipId)
				: undefined;
			if (nodeId) {
				nodeIds.add(nodeId);
			}
			if (edgeId) {
				edgeIds.add(edgeId);
			}
		}
		for (const id of nodeIds) {
			byNodeId.set(id, [...(byNodeId.get(id) ?? []), entry]);
		}
		for (const id of edgeIds) {
			byEdgeId.set(id, [...(byEdgeId.get(id) ?? []), entry]);
		}
		if (!nodeIds.size && !edgeIds.size) {
			unattached.push(entry);
		}
	}
	return { byNodeId, byEdgeId, unattached };
}

export function buildAnalysisGraph(
	analysisNodes: SystemAnalysisNode[],
	analysisEdges: SystemAnalysisEdge[],
	attachments: EntryAttachments
) {
	const nodeIds = new Map(analysisNodes.map((node) => [node.attributes.knowledgeEntity.id, node.id]));
	const nodes: SystemDiagramNode[] = analysisNodes.map((node) => ({
		id: node.id,
		type: "entity",
		position: { ...node.attributes.position },
		data: {
			entity: node.attributes.knowledgeEntity,
			attachmentCount: attachments.byNodeId.get(node.id)?.length ?? 0,
		},
	}));
	const edges: SystemDiagramEdge[] = [];
	for (const edge of analysisEdges) {
		const relationship = edge.attributes.knowledgeRelationship;
		const source = nodeIds.get(relationship.attributes.sourceEntityId);
		const target = nodeIds.get(relationship.attributes.targetEntityId);
		if (!source || !target) {
			continue;
		}
		edges.push({
			id: edge.id,
			type: "relationship",
			source,
			target,
			data: {
				relationship,
				attachmentCount: attachments.byEdgeId.get(edge.id)?.length ?? 0,
			},
		});
	}
	return { nodes, edges };
}
