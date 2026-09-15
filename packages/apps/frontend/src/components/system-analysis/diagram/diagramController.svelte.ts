import { Context, watch, type Getter } from "runed";
import { tick } from "svelte";
import { SvelteMap } from "svelte/reactivity";

import {
	useSvelteFlow,
	useStore as useSvelteFlowStore,
	type Node,
	type Edge,
	type XYPosition,
	type Connection,
} from "@xyflow/svelte";

import { type SystemAnalysisNode, type SystemAnalysisEdge, type KnowledgeGraphEntity } from "$lib/api";

import { useSystemAnalysisController } from "../controller.svelte";

export type SystemTopologyNodeData = {
	analysisNode: SystemAnalysisNode;
	attachmentCount: number;
};

export type SystemRelationshipEdgeData = {
	edge: SystemAnalysisEdge;
	attachmentCount: number;
};

export type DiagramContextMenuState = {
	nodeId?: string;
	edgeId?: string;
	containerRect: DOMRect;
	clickPos: XYPosition;
};

type DiagramSelection = { node?: Node; edge?: Edge };
export type GraphSelection = { nodeId?: string; edgeId?: string };
export type GraphHighlights = { nodeIds: string[]; edgeIds: string[] };

export type GraphInteraction = {
	selection?: GraphSelection;
	highlights?: GraphHighlights;
	onReady?: (focus: (subjects: GraphHighlights) => Promise<void>) => void;
	select: (selection: GraphSelection, trigger?: HTMLElement) => void;
};

const translateSystemAnalysis = (
	analysisNodes: SystemAnalysisNode[],
	analysisEdges: SystemAnalysisEdge[],
	nodeAttachmentCount: (id: string) => number,
	edgeAttachmentCount: (id: string) => number
) => {
	let nodes: Node[] = [];
	const nodeIdsByEntityId = new SvelteMap<string, string>();
	analysisNodes.forEach((analysisNode) => {
		const { position, knowledgeEntity } = analysisNode.attributes;
		nodeIdsByEntityId.set(knowledgeEntity.id, analysisNode.id);
		nodes.push({
			id: analysisNode.id,
			type: "component",
			position,
			data: {
				analysisNode,
				attachmentCount: nodeAttachmentCount(analysisNode.id),
			} as SystemTopologyNodeData,
		});
	});

	let edges: Edge[] = [];
	analysisEdges.forEach((sr) => {
		const { id, attributes } = sr;
		const relattr = attributes.knowledgeRelationship.attributes;
		const source = nodeIdsByEntityId.get(relattr.sourceEntityId);
		const target = nodeIdsByEntityId.get(relattr.targetEntityId);
		if (!source || !target) return;
		edges.push({
			id,
			type: "relationship",
			source,
			target,
			data: { edge: sr, attachmentCount: edgeAttachmentCount(id) } as SystemRelationshipEdgeData,
		});
	});

	return { nodes, edges };
};

const getSelected = <T extends {id: string}>(items: T[], selectedId: string | undefined) => {
	if (!selectedId) return;
	return items.find(item => (item.id === selectedId));
}

export class DiagramController {
	analysis = useSystemAnalysisController();

	private localSelection = $state<GraphSelection>({});
	private externalInteraction = $state.raw<GraphInteraction>();
	interaction = $derived(this.externalInteraction ?? {
		selection: this.localSelection,
		select: (s) => {this.localSelection = s},
	});
	selection = $derived(this.interaction.selection);

	contextMenu = $state.raw<DiagramContextMenuState>();

	nodes = $state.raw<Node[]>([]);
	edges = $state.raw<Edge[]>([]);

	selectedNode = $derived(getSelected(this.nodes, this.selection?.nodeId));
	selectedEdge = $derived(getSelected(this.edges, this.selection?.edgeId));
	private selected = $derived<DiagramSelection>({ node: this.selectedNode, edge: this.selectedEdge });

	highlights = $derived(this.interaction.highlights);
	selectedLivePosition = $state<XYPosition>();

	containerEl = $state.raw<HTMLElement>(null!);
	addingEntityGhost = $state.raw<KnowledgeGraphEntity>();

	constructor(interactionFn: Getter<GraphInteraction | undefined>) {
		watch(interactionFn, interaction => {
			this.externalInteraction = interaction;
		});

		watch(
			() =>
				[
					this.analysis.analysisNodes,
					this.analysis.analysisEdges,
					this.analysis.attachments,
				] as const,
			([nodes, edges]) => {
				this.onAnalysisGraphUpdate(nodes, edges);
			}
		);

		watch(() => this.highlights, () => this.updateHighlights());
		watch(() => this.selected, (sel) => {this.updateSelectedPosition(sel)});
	}

	onAnalysisGraphUpdate(nodes: SystemAnalysisNode[], edges: SystemAnalysisEdge[]) {
		const translated = translateSystemAnalysis(
			$state.snapshot(nodes),
			$state.snapshot(edges),
			(id) => this.analysis.attachments.byNodeId.get(id)?.length ?? 0,
			(id) => this.analysis.attachments.byEdgeId.get(id)?.length ?? 0
		);
		this.nodes = translated.nodes;
		this.edges = translated.edges;
		this.updateHighlights();
	}

	private updateHighlights() {
		if (!this.highlights) return;
		const nodeIds = new Set(this.highlights.nodeIds);
		const edgeIds = new Set(this.highlights.edgeIds);
		this.nodes = this.nodes.map((node) =>
			!!node.selected === nodeIds.has(node.id) ? node : { ...node, selected: nodeIds.has(node.id) }
		);
		this.edges = this.edges.map((edge) =>
			!!edge.selected === edgeIds.has(edge.id) ? edge : { ...edge, selected: edgeIds.has(edge.id) }
		);
	}

	private flow?: ReturnType<typeof useSvelteFlow>;

	focusSubjects = async (subjects: GraphHighlights) => {
		await tick();
		const ids = new Set(subjects.nodeIds);
		for (const edge of this.edges) {
			if (!subjects.edgeIds.includes(edge.id)) continue;
			ids.add(edge.source);
			ids.add(edge.target);
		}
		const nodes = this.nodes.filter((node) => ids.has(node.id));
		if (nodes.length) {
			await this.flow?.fitView({ nodes, padding: 0.3, maxZoom: 1.25, duration: 200 });
		}
	};

	getNodesBounds = $state.raw<ReturnType<typeof useSvelteFlow>["getNodesBounds"]>();
	flowStore = $state.raw<ReturnType<typeof useSvelteFlowStore>>();

	onFlowInit(flow: ReturnType<typeof useSvelteFlow>, store: ReturnType<typeof useSvelteFlowStore>) {
		this.flow = flow;
		this.getNodesBounds = flow.getNodesBounds;
		this.flowStore = store;
		this.interaction.onReady?.(this.focusSubjects);
	}

	interactionLocked() {
		return this.flowStore && !this.flowStore.elementsSelectable;
	}

	updateSelectedPosition({ node, edge }: DiagramSelection) {
		if (edge) {
			this.selectedLivePosition = this.getNodesBounds?.([edge.source, edge.target]);
		} else if (node) {
			this.selectedLivePosition = node.position;
		} else {
			this.selectedLivePosition = undefined;
		}
	}

	setSelected({ node, edge }: DiagramSelection, event?: MouseEvent | TouchEvent) {
		this.closeContextMenu();
		const trigger =
			event?.target instanceof Element
				? (event.target.closest<HTMLElement>(".svelte-flow__node, .svelte-flow__edge") ?? undefined)
				: undefined;
		this.interaction.select({nodeId: node?.id,edgeId: edge?.id}, trigger);
	}

	handleNodeDragStart({targetNode}: { targetNode?: Node | null }) {
		this.setSelected({ node: !!targetNode ? targetNode: undefined });
	}

	handleNodeDrag({targetNode: node}: { targetNode?: Node | null }) {
		if (!this.selection?.nodeId || !node) return;
		if (this.selection.nodeId === node.id) {
			this.updateSelectedPosition({ node });
		}
	}

	handleNodeDragStop({targetNode: node}: { targetNode?: Node | null }) {
		if (!node) return;
		const { analysisNode } = node.data as SystemTopologyNodeData;
		if (!analysisNode) return;

		this.analysis.updateNode(analysisNode.id, {position: node.position});
	}

	setAddingEntityGhost(e?: KnowledgeGraphEntity) {
		this.addingEntityGhost = e;
	}

	handlePaneClicked({ event }: { event: MouseEvent }) {
		this.setSelected({});

		if (this.addingEntityGhost) {
			event.preventDefault();

			if (!this.containerEl || !("pageX" in event)) return;

			const { x, y } = this.containerEl.getBoundingClientRect();

			const position = { x: event.pageX - x, y: event.pageY - y };
			const knowledgeEntityId = this.addingEntityGhost.id;
			this.analysis.addNode({ knowledgeEntityId, position });
			// TODO: check if success? show pending state?
			this.setAddingEntityGhost();
		}
	}

	handleNodeClicked({ node, event }: { node: Node; event: MouseEvent | TouchEvent }) {
		if (this.interactionLocked()) return;
		this.setSelected({ node }, event);
	}

	handleEdgeClicked({ edge, event }: { edge: Edge; event: MouseEvent | TouchEvent }) {
		if (this.interactionLocked()) return;
		this.setSelected({ edge }, event);
	}

	handleContextMenuEvent({event, node, edge, nodes}: { event: MouseEvent; node?: Node; edge?: Edge; nodes?: Node[] }) {
		if (this.interactionLocked()) return;
		if (!this.containerEl) return;

		event.preventDefault();

		if (!("pageX" in event && "pageY" in event)) return;

		this.contextMenu = {
			nodeId: node?.id,
			edgeId: edge?.id,
			clickPos: { x: event.pageX, y: event.pageY },
			containerRect: this.containerEl.getBoundingClientRect(),
		};
	}

	closeContextMenu() {
		this.contextMenu = undefined;
	}

	onEdgeConnect({ source, target }: Connection) {
		// undo auto-created edge, need to confirm via dialog
		this.edges = this.edges.filter((e) => !(e.source === source && e.target === target));
	}
}

const ctx = new Context<DiagramController>("SystemDiagramController");
export const initDiagramController = (interactionFn: Getter<GraphInteraction | undefined>) => ctx.set(new DiagramController(interactionFn));
export const useDiagramController = () => ctx.get();
