import { createInfiniteQuery, createMutation } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import {
	addSystemAnalysisEdgeMutation,
	addSystemAnalysisNodeMutation,
	deleteSystemAnalysisEdgeMutation,
	deleteSystemAnalysisNodeMutation,
	listSystemAnalysisEdgesInfiniteOptions,
	listSystemAnalysisEntriesInfiniteOptions,
	listSystemAnalysisNodesInfiniteOptions,
	updateSystemAnalysisEdgeMutation,
	updateSystemAnalysisNodeMutation,
	type AddSystemAnalysisEdgeAttributes,
	type AddSystemAnalysisNodeAttributes,
	type UpdateSystemAnalysisEdgeAttributes,
	type UpdateSystemAnalysisNodeAttributes,
} from "$lib/api";
import { flattenPages, mapEntryAttachments, nextPageOffset } from "./lib";

const pageSize = 50;

export class SystemAnalysisController {
	analysisId = $state.raw("");

	nodesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisNodesInfiniteOptions({
			path: { id: this.analysisId },
			query: { limit: pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 0,
		getNextPageParam: (_last, pages) => nextPageOffset(pages),
	}));
	edgesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEdgesInfiniteOptions({
			path: { id: this.analysisId },
			query: { limit: pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 0,
		getNextPageParam: (_last, pages) => nextPageOffset(pages),
	}));
	entriesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEntriesInfiniteOptions({
			path: { id: this.analysisId },
			query: { limit: pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 0,
		getNextPageParam: (_last, pages) => nextPageOffset(pages),
	}));

	analysisNodes = $derived(flattenPages(this.nodesQuery.data?.pages));
	analysisEdges = $derived(flattenPages(this.edgesQuery.data?.pages));
	entries = $derived(flattenPages(this.entriesQuery.data?.pages));
	attachments = $derived(mapEntryAttachments(this.analysisNodes, this.analysisEdges, this.entries));
	graphLoading = $derived(
		this.nodesQuery.isPending ||
			this.edgesQuery.isPending ||
			this.nodesQuery.hasNextPage ||
			this.edgesQuery.hasNextPage
	);
	graphError = $derived(this.nodesQuery.error ?? this.edgesQuery.error);

	constructor(getAnalysisId: () => string) {
		this.analysisId = getAnalysisId();
		watch(getAnalysisId, (id) => {
			this.analysisId = id;
		});
		$effect(() => {
			if (this.nodesQuery.hasNextPage && !this.nodesQuery.isFetchingNextPage)
				this.nodesQuery.fetchNextPage();
			if (this.edgesQuery.hasNextPage && !this.edgesQuery.isFetchingNextPage)
				this.edgesQuery.fetchNextPage();
			if (this.entriesQuery.hasNextPage && !this.entriesQuery.isFetchingNextPage)
				this.entriesQuery.fetchNextPage();
		});
	}

	refreshEntries = () => this.entriesQuery.refetch();
	private refreshNodes = () => this.nodesQuery.refetch();
	private refreshEdges = () => this.edgesQuery.refetch();

	async refreshAll() {
		return Promise.all([
			this.refreshNodes(),
			this.refreshEdges(),
		])
	}

	private addNodeMutation = createMutation(() => ({
		...addSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	addNode(attributes: AddSystemAnalysisNodeAttributes) {
		return this.addNodeMutation.mutateAsync({ path: { id: this.analysisId }, body: { attributes } });
	}
	private updateNodeMutation = createMutation(() => ({
		...updateSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	updateNode(id: string, attributes: UpdateSystemAnalysisNodeAttributes) {
		return this.updateNodeMutation.mutate({ path: { id }, body: { attributes } });
	}
	private removeNodeMutation = createMutation(() => ({
		...deleteSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	removeNode(id: string) {
		return this.removeNodeMutation.mutate({ path: { id } });
	}
	private addEdgeMutation = createMutation(() => ({
		...addSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	addEdge(attributes: AddSystemAnalysisEdgeAttributes) {
		return this.addEdgeMutation.mutate({ path: { id: this.analysisId }, body: { attributes } });
	}
	private updateEdgeMutation = createMutation(() => ({
		...updateSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	updateEdge(id: string, attributes: UpdateSystemAnalysisEdgeAttributes) {
		return this.updateEdgeMutation.mutate({ path: { id }, body: { attributes } });
	}
	private removeEdgeMutation = createMutation(() => ({
		...deleteSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	removeEdge(id: string) {
		return this.removeEdgeMutation.mutate({ path: { id } });
	}
}

const context = new Context<SystemAnalysisController>("SystemAnalysisController");
export const initSystemAnalysisController = (getAnalysisId: () => string) =>
	context.set(new SystemAnalysisController(getAnalysisId));
export const useSystemAnalysisController = () => context.get();
