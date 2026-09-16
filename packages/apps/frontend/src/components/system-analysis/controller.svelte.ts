import { createInfiniteQuery, createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import {
	deleteSystemAnalysisEdgeMutation,
	deleteSystemAnalysisNodeMutation,
	listSystemAnalysisEdgesInfiniteOptions,
	listSystemAnalysisEntriesInfiniteOptions,
	listSystemAnalysisNodesInfiniteOptions,
	updateSystemAnalysisNodeMutation,
	type ErrorModel,
} from "$lib/api";
import { getNextPageParam } from "$lib/api/utils";
import { flattenPages, mapEntryAttachments, buildAnalysisGraph } from "./lib";
import type { XYPosition } from "@xyflow/svelte";
import { SystemDiagramController } from "$components/system-diagram/controller.svelte.ts";
import type { DiagramContextMenu, GraphSelection, GraphHighlights } from "$components/system-diagram";

export type GraphInteraction = {
	selection?: GraphSelection;
	highlights?: GraphHighlights;
	select: (selection: GraphSelection, trigger?: HTMLElement) => void;
};

export type SystemAnalysisOptions = {
	interaction?: GraphInteraction;
	readOnly?: boolean;
};

const pageSize = 50;

export class SystemAnalysisController {
	analysisId = $state<string>();
	options = $state.raw<SystemAnalysisOptions>();
	readOnly = $derived(this.options?.readOnly ?? false);

	private interaction = $derived(this.options?.interaction);
	private localSelection = $state<GraphSelection>({});
	selection = $derived(this.interaction?.selection ?? this.localSelection);

	ctxMenu = $state<DiagramContextMenu>();

	deleting = $state(false);

	diagram = new SystemDiagramController({
		selection: () => this.selection,
		highlights: () => this.interaction?.highlights,
		select: (selection, trigger) => {
			this.onSelected(selection, trigger);
		},
		onNodeMove: (id, pos) => {
			this.onMoveNode(id, pos);
		},
		onContextMenu: (menu) => {
			if (!this.readOnly) this.ctxMenu = menu;
		},
	});

	constructor(idFn: Getter<string | undefined>, optionsFn: Getter<SystemAnalysisOptions> = () => ({})) {
		watch(idFn, (id) => {
			this.analysisId = id;
			this.localSelection = {};
			this.ctxMenu = undefined;
			this.mutationError = undefined;
			this.deleting = false;
			this.framedAnalysis = undefined;
			this.diagram.setGraph([], []);
			this.diagram.viewport = { x: 0, y: 0, zoom: 1 };
			void this.diagram.fit();
		});

		watch(optionsFn, (opts) => {
			this.options = opts;
		});

		this.watchForRefresh();

		watch(
			() => [this.analysisId, this.graphLoading, this.diagram.nodes.length] as const,
			() => this.frameAnalysis()
		);

		for (const query of [this.nodesQuery, this.edgesQuery, this.entriesQuery]) {
			watch(
				() => query.hasNextPage && !query.isFetching && !query.isError,
				(hasMore) => {
					if (hasMore) query.fetchNextPage();
				}
			);
		}
	}

	private watchForRefresh() {
		watch(
			() =>
				[
					this.analysisId,
					this.analysisNodes,
					this.analysisEdges,
					this.attachments,
					this.readOnly,
				] as const,
			() => {
				if (this.readOnly) {
					this.ctxMenu = undefined;
				}
				this.refreshGraph();
			}
		);
	}

	private framedAnalysis?: string;

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

	attachments = $derived(mapEntryAttachments(this.analysisNodes, this.analysisEdges, this.entries));

	graphLoading = $derived(
		!!this.analysisId &&
			(this.nodesQuery.isPending ||
				this.nodesQuery.hasNextPage ||
				this.edgesQuery.isPending ||
				this.edgesQuery.hasNextPage)
	);
	graphError = $derived(this.nodesQuery.error ?? this.edgesQuery.error);
	hasGraph = $derived(!!this.nodesQuery.data && !!this.edgesQuery.data);

	selectionInspector = $derived.by(() => {
		const node = this.diagram.selectedNode;
		const edge = this.diagram.selectedEdge;
		if (!node && !edge) {
			return undefined;
		}
		const entries = node
			? this.attachments.byNodeId.get(node.id)
			: this.attachments.byEdgeId.get(edge!.id);

		return {
			title: node ? "Node entries" : "Relationship entries",
			entries: (entries ?? []).map(({ id, attributes }) => ({
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

	mutationError = $state<ErrorModel>();

	private updateNodeMutation = createMutation(() => ({
		...updateSystemAnalysisNodeMutation(),
		onSuccess: () => this.nodesQuery.refetch(),
	}));

	private removeNodeMutation = createMutation(() => ({
		...deleteSystemAnalysisNodeMutation(),
		onSuccess: this.refreshAll,
	}));

	private removeEdgeMutation = createMutation(() => ({
		...deleteSystemAnalysisEdgeMutation(),
		onSuccess: this.refreshAll,
	}));

	private frameAnalysis() {
		const id = this.analysisId;
		if (!id || this.graphLoading || !this.diagram.nodes.length || this.framedAnalysis === id) {
			return;
		}
		this.framedAnalysis = id;
		void this.diagram.fit();
	}

	private refreshGraph() {
		const graph = buildAnalysisGraph(this.analysisNodes, this.analysisEdges, this.attachments);
		for (const node of graph.nodes) {
			node.draggable = !this.readOnly;
		}
		this.diagram.setGraph(graph.nodes, graph.edges);
	}

	container = $state<HTMLElement>();
	setContainer = (element: HTMLElement) => {
		this.container = element;
		return () => {
			this.container = undefined;
		};
	};

	onSelected = (selection: GraphSelection, trigger?: HTMLElement) => {
		this.ctxMenu = undefined;
		const interaction = this.interaction;
		if (interaction) {
			interaction.select(selection, trigger);
		} else {
			this.localSelection = selection;
		}
	};

	private onMoveNode = async (id: string, position: XYPosition) => {
		if (this.readOnly) {
			return;
		}

		const analysisId = this.analysisId;
		this.mutationError = undefined;
		try {
			await this.updateNodeMutation.mutateAsync({
				path: { id },
				body: { attributes: { position } },
			});
		} catch (error) {
			if (this.analysisId === analysisId) {
				this.mutationError = error as ErrorModel;
				this.refreshGraph();
			}
		}
	};

	remove = async (selection: GraphSelection) => {
		if (this.readOnly || this.deleting || (!selection.nodeId && !selection.edgeId)) {
			return;
		}

		const subject = selection.nodeId ? "node" : "relationship";
		if (!confirm(`Remove this ${subject} from the analysis?`)) {
			return;
		}

		const analysisId = this.analysisId;
		this.mutationError = undefined;
		this.deleting = true;
		try {
			if (selection.nodeId) {
				await this.removeNodeMutation.mutateAsync({ path: { id: selection.nodeId } });
			} else if (selection.edgeId) {
				await this.removeEdgeMutation.mutateAsync({ path: { id: selection.edgeId } });
			}
			if (this.analysisId !== analysisId) {
				return;
			}
			if (this.selection.nodeId === selection.nodeId && this.selection.edgeId === selection.edgeId) {
				this.onSelected({});
			}
			this.ctxMenu = undefined;
		} catch (error) {
			if (this.analysisId === analysisId) {
				this.mutationError = error as ErrorModel;
			}
		} finally {
			if (this.analysisId === analysisId) {
				this.deleting = false;
			}
		}
	};
}

const ctx = new Context<SystemAnalysisController>("SystemAnalysisController");
export const initSystemAnalysisController = (
	idFn: Getter<string | undefined>,
	optionsFn?: Getter<SystemAnalysisOptions>
) => ctx.set(new SystemAnalysisController(idFn, optionsFn));
export const useSystemAnalysisController = () => ctx.get();
