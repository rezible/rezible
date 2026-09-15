import { createInfiniteQuery, createMutation, type CreateInfiniteQueryResult } from "@tanstack/svelte-query";
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
	type ErrorModel,
	type UpdateSystemAnalysisEdgeAttributes,
	type UpdateSystemAnalysisNodeAttributes,
} from "$lib/api";
import { getNextPageParam } from "$lib/api/utils";
import { flattenPages, mapEntryAttachments } from "./lib";

const pageSize = 50;

const idPath = (id?: string) => ({ id: id ?? "" });

const makeInfiniteQueryFetchWatcher = (q: CreateInfiniteQueryResult<any, ErrorModel>) => {
	watch(() => (q.hasNextPage && !q.isFetching && !q.isError), shouldFetchMore => {
		if (shouldFetchMore) q.fetchNextPage();
	});
}

export class SystemAnalysisController {
	analysisId = $state<string>();

	constructor(idFn: () => string | undefined) {
		watch(idFn, (id) => {
			this.analysisId = id;
		});
		
		makeInfiniteQueryFetchWatcher(this.nodesQuery);
		makeInfiniteQueryFetchWatcher(this.edgesQuery);
		makeInfiniteQueryFetchWatcher(this.entriesQuery);

		// $effect(() => {
		// 	if (this.nodesQuery.hasNextPage && !this.nodesQuery.isFetching && !this.nodesQuery.isError)
		// 		this.nodesQuery.fetchNextPage();
		// 	if (this.edgesQuery.hasNextPage && !this.edgesQuery.isFetching && !this.edgesQuery.isError)
		// 		this.edgesQuery.fetchNextPage();
		// 	if (this.entriesQuery.hasNextPage && !this.entriesQuery.isFetching && !this.entriesQuery.isError)
		// 		this.entriesQuery.fetchNextPage();
		// });
	}

	private analysisIdPath = $derived(idPath(this.analysisId));

	nodesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisNodesInfiniteOptions({
			path: this.analysisIdPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	private nodesQueryPending = $derived(this.nodesQuery.isPending || this.nodesQuery.hasNextPage);
	private nodesQueryData = $derived(this.nodesQuery.data);
	private refreshNodes = () => this.nodesQuery.refetch();

	edgesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEdgesInfiniteOptions({
			path: this.analysisIdPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	private edgesQueryPending = $derived(this.edgesQuery.isPending || this.edgesQuery.hasNextPage);
	private edgesQueryData = $derived(this.edgesQuery.data);
	private refreshEdges = () => this.edgesQuery.refetch();

	entriesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEntriesInfiniteOptions({
			path: this.analysisIdPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	refreshEntries = () => this.entriesQuery.refetch();

	refreshAll = () => Promise.all([this.refreshNodes(), this.refreshEdges(), this.refreshEntries()]);

	analysisNodes = $derived(flattenPages(this.nodesQueryData?.pages));
	analysisEdges = $derived(flattenPages(this.edgesQueryData?.pages));
	entries = $derived(flattenPages(this.entriesQuery.data?.pages));
	attachments = $derived(mapEntryAttachments(this.analysisNodes, this.analysisEdges, this.entries));
	
	graphLoading = $derived(!!this.analysisId && (this.nodesQueryPending || this.edgesQueryPending));
	graphError = $derived(this.nodesQuery.error ?? this.edgesQuery.error);
	hasGraph = $derived(!!this.nodesQueryData && !!this.edgesQueryData);

	private addNodeMut = createMutation(() => ({
		...addSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	addNode(attributes: AddSystemAnalysisNodeAttributes) {
		return this.addNodeMut.mutateAsync({ 
			path: this.analysisIdPath, 
			body: { attributes },
		});
	}

	private updateNodeMut = createMutation(() => ({
		...updateSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	updateNode(id: string, attributes: UpdateSystemAnalysisNodeAttributes) {
		return this.updateNodeMut.mutate({ 
			path: { id }, 
			body: { attributes },
		});
	}

	private removeNodeMut = createMutation(() => ({
		...deleteSystemAnalysisNodeMutation(),
		onSuccess: this.refreshNodes,
	}));
	removeNode(id: string) {
		return this.removeNodeMut.mutate({ 
			path: { id },
		});
	}

	private addEdgeMut = createMutation(() => ({
		...addSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	addEdge(attributes: AddSystemAnalysisEdgeAttributes) {
		return this.addEdgeMut.mutate({ 
			path: this.analysisIdPath,
			body: { attributes },
		});
	}

	private updateEdgeMut = createMutation(() => ({
		...updateSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	updateEdge(id: string, attributes: UpdateSystemAnalysisEdgeAttributes) {
		return this.updateEdgeMut.mutate({ 
			path: { id },
			body: { attributes },
		});
	}

	private removeEdgeMut = createMutation(() => ({
		...deleteSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshEdges,
	}));
	removeEdge(id: string) {
		return this.removeEdgeMut.mutate({ 
			path: { id },
		});
	}
}

const ctx = new Context<SystemAnalysisController>("SystemAnalysisController");
export const initSystemAnalysisController = (idFn: () => string | undefined) => ctx.set(new SystemAnalysisController(idFn));
export const useSystemAnalysisController = () => ctx.get();
