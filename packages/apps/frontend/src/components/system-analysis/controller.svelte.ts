import { createInfiniteQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import {
	listSystemAnalysisEdgesInfiniteOptions,
	listSystemAnalysisEntriesInfiniteOptions,
	listSystemAnalysisNodesInfiniteOptions,
	type SystemAnalysisEdge,
	type SystemAnalysisNode,
} from "$lib/api";
import { getNextPageParam } from "$lib/api/utils";
import { reconcileSelection } from "$features/system/lib/system-map/selection";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { buildAnalysisGraph, flattenPages, mapEntryAttachments, type AnalysisGraph } from "./lib";

export type GraphInteraction = {
	selection?: MapSelection;
	highlights?: MapHighlights;
	select: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
};

export type SystemAnalysisOptions = {
	interaction?: GraphInteraction;
};

export type PublishedSystemAnalysis = AnalysisGraph & {
	analysisId: string;
};

const pageSize = 50;
const emptyNodeMap = new Map<string, SystemAnalysisNode>();
const emptyEdgeMap = new Map<string, SystemAnalysisEdge>();

export class SystemAnalysisController {
	analysisId = $state<string>();
	options = $state.raw<SystemAnalysisOptions>();

	private interaction = $derived(this.options?.interaction);
	private localSelection = $state<MapSelection>();
	selection = $derived(this.interaction ? this.interaction.selection : this.localSelection);
	highlights = $derived(this.interaction?.highlights);

	publishedAnalysis = $state.raw<PublishedSystemAnalysis>();
	graph = $derived(this.publishedAnalysis?.graph);
	positions = $derived(this.publishedAnalysis?.positions);
	hasGraph = $derived(this.publishedAnalysis?.analysisId === this.analysisId && !!this.analysisId);

	constructor(idFn: Getter<string | undefined>, optionsFn: Getter<SystemAnalysisOptions> = () => ({})) {
		watch(idFn, (id) => {
			this.analysisId = id;
			this.localSelection = undefined;
			this.publishedAnalysis = undefined;
		});

		watch(optionsFn, (options) => {
			this.options = options;
		});

		watch(
			() =>
				[
					this.analysisId,
					this.nodesQuery.data?.pages,
					this.nodesQuery.isSuccess,
					this.nodesQuery.isFetching,
					this.nodesQuery.hasNextPage,
					this.edgesQuery.data?.pages,
					this.edgesQuery.isSuccess,
					this.edgesQuery.isFetching,
					this.edgesQuery.hasNextPage,
				] as const,
			() => this.publishCompletedGraph()
		);

		watch(
			() => [this.publishedAnalysis, this.selection] as const,
			([published, selection]) => {
				if (!published || !selection) return;
				if (!reconcileSelection(published.graph, selection)) this.onSelected(undefined);
			}
		);

		for (const query of [this.nodesQuery, this.edgesQuery, this.entriesQuery]) {
			watch(
				() => query.hasNextPage && !query.isFetching && !query.isError,
				(hasMore) => {
					if (hasMore) void query.fetchNextPage();
				}
			);
		}
	}

	private analysisPath = $derived({ id: this.analysisId ?? "" });

	nodesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisNodesInfiniteOptions({
			path: this.analysisPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	analysisNodes = $derived(flattenPages(this.nodesQuery.data?.pages));

	edgesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEdgesInfiniteOptions({
			path: this.analysisPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	analysisEdges = $derived(flattenPages(this.edgesQuery.data?.pages));

	entriesQuery = createInfiniteQuery(() => ({
		...listSystemAnalysisEntriesInfiniteOptions({
			path: this.analysisPath,
			query: { pageSize },
		}),
		enabled: !!this.analysisId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	entries = $derived(flattenPages(this.entriesQuery.data?.pages));
	refreshEntries = () => this.entriesQuery.refetch();
	refreshAll = () =>
		Promise.all([this.nodesQuery.refetch(), this.edgesQuery.refetch(), this.entriesQuery.refetch()]);

	graphError = $derived(this.nodesQuery.error ?? this.edgesQuery.error);
	graphLoading = $derived(
		!!this.analysisId &&
			!this.hasGraph &&
			(this.nodesQuery.isPending ||
				this.nodesQuery.isFetching ||
				this.nodesQuery.hasNextPage ||
				this.edgesQuery.isPending ||
				this.edgesQuery.isFetching ||
				this.edgesQuery.hasNextPage)
	);

	attachments = $derived(
		mapEntryAttachments(
			this.publishedAnalysis?.nodeByEntityId ?? emptyNodeMap,
			this.publishedAnalysis?.edgeByRelationshipId ?? emptyEdgeMap,
			this.entries
		)
	);

	nodeByEntityId = $derived(this.publishedAnalysis?.nodeByEntityId ?? emptyNodeMap);
	edgeByRelationshipId = $derived(this.publishedAnalysis?.edgeByRelationshipId ?? emptyEdgeMap);

	selectionInspector = $derived.by(() => {
		const selection = this.selection;
		if (!selection) return undefined;

		const entriesById = new Map<string, (typeof this.entries)[number]>();
		if (selection.kind === "entity") {
			for (const entry of this.attachments.byEntityId.get(selection.entityId) ?? []) {
				entriesById.set(entry.id, entry);
			}
		} else {
			const relationshipIds =
				selection.kind === "relationship" ? [selection.relationshipId] : selection.relationshipIds;
			for (const relationshipId of relationshipIds) {
				for (const entry of this.attachments.byRelationshipId.get(relationshipId) ?? []) {
					entriesById.set(entry.id, entry);
				}
			}
		}

		return {
			title: selection.kind === "entity" ? "Entity entries" : "Relationship entries",
			entries: [...entriesById.values()].map(({ id, attributes }) => ({
				id,
				title: attributes.title,
				body: attributes.body,
				kind: attributes.kind,
				occurredAt: attributes.occurredAt
					? new Date(attributes.occurredAt).toLocaleString()
					: undefined,
			})),
		};
	});

	private publishCompletedGraph() {
		const analysisId = this.analysisId;
		if (
			!analysisId ||
			!this.nodesQuery.isSuccess ||
			this.nodesQuery.isFetching ||
			this.nodesQuery.hasNextPage ||
			!this.edgesQuery.isSuccess ||
			this.edgesQuery.isFetching ||
			this.edgesQuery.hasNextPage
		) {
			return;
		}

		const graph = buildAnalysisGraph(this.analysisNodes, this.analysisEdges);
		if (this.analysisId === analysisId) {
			this.publishedAnalysis = { analysisId, ...graph };
		}
	}

	onSelected = (selection: MapSelection | undefined, trigger?: HTMLElement) => {
		const interaction = this.interaction;
		if (interaction) {
			interaction.select(selection, trigger);
		} else {
			this.localSelection = selection;
		}
	};
}

const ctx = new Context<SystemAnalysisController>("SystemAnalysisController");
export const initSystemAnalysisController = (
	idFn: Getter<string | undefined>,
	optionsFn?: Getter<SystemAnalysisOptions>
) => ctx.set(new SystemAnalysisController(idFn, optionsFn));
export const useSystemAnalysisController = () => ctx.get();
