// TODO: move a lot of this

import type { SystemAnalysisEdge, SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import type { ResponsePage } from "$lib/api/utils";

export const flattenPages = <T extends { id: string }>(pages: ResponsePage<T>[] | undefined) => {
	const items = new Map<string, T>();
	for (const page of pages ?? []) for (const item of page.data) items.set(item.id, item);
	return [...items.values()];
};

const loadedPageCount = <T>(pages: ResponsePage<T>[] | undefined) => {
	if (!pages) return 0;
	return pages.reduce((count, {data}) => count + data.length, 0);
}

export const nextPageOffset = <T>(pages: ResponsePage<T>[]) => {
	const loaded = loadedPageCount(pages);
	return loaded < (pages.at(-1)?.pagination.total ?? 0) ? loaded : undefined;
};

export type EntryAttachments = {
	byNodeId: Map<string, SystemAnalysisEntry[]>;
	byEdgeId: Map<string, SystemAnalysisEntry[]>;
	unattached: SystemAnalysisEntry[];
};

export const mapEntryAttachments = (
	nodes: SystemAnalysisNode[],
	edges: SystemAnalysisEdge[],
	entries: SystemAnalysisEntry[]
): EntryAttachments => {
	const nodeByEntity = new Map(nodes.map((node) => [node.attributes.knowledgeEntity.id, node.id]));
	const edgeByRelationship = new Map(
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
				? nodeByEntity.get(subject.attributes.knowledgeEntityId)
				: undefined;
			const edgeId = subject.attributes.knowledgeRelationshipId
				? edgeByRelationship.get(subject.attributes.knowledgeRelationshipId)
				: undefined;
			if (nodeId) nodeIds.add(nodeId);
			if (edgeId) edgeIds.add(edgeId);
		}
		for (const id of nodeIds) byNodeId.set(id, [...(byNodeId.get(id) ?? []), entry]);
		for (const id of edgeIds) byEdgeId.set(id, [...(byEdgeId.get(id) ?? []), entry]);
		if (!nodeIds.size && !edgeIds.size) unattached.push(entry);
	}
	return { byNodeId, byEdgeId, unattached };
};
