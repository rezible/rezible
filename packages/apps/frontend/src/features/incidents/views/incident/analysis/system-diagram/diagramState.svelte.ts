import { Context, watch } from "runed";

import {
	useSvelteFlow,
	useStore as useSvelteFlowStore,
	type Node,
	type Edge,
	type XYPosition,
	type Connection,
} from "@xyflow/svelte";

import {
	type SystemAnalysis,
	type SystemAnalysisNode,
	type SystemAnalysisEdge,
	type SystemTopologyEntity,
} from "$lib/api";

import { useIncidentAnalysis } from "../controller.svelte";

export type SystemTopologyNodeData = {
	analysisNode: SystemAnalysisNode;
};

export type SystemRelationshipEdgeData = {
	edge: SystemAnalysisEdge;
};

const translateSystemAnalysis = (an: SystemAnalysis) => {
	let nodes: Node[] = [];
	an.attributes.nodes.forEach(analysisNode => {
		const { position, snapshotEntity } = analysisNode.attributes;
		nodes.push({
			id: snapshotEntity.id,
			type: "component",
			position,
			data: { analysisNode } as SystemTopologyNodeData,
		});
	});

	let edges: Edge[] = [];
	an.attributes.edges.forEach(sr => {
		const { id, attributes } = sr;
		const relattr = attributes.snapshotRelationship.attributes;
		edges.push({
			id,
			type: "relationship",
			source: relattr.sourceSnapshotEntityId,
			target: relattr.targetSnapshotEntityId,
			data: { edge: sr } as SystemRelationshipEdgeData,
		});
	});

	return { nodes, edges };
};

type DiagramSelectionState = { node?: Node; edge?: Edge };

export class SystemDiagramState {
	analysis = useIncidentAnalysis();

	selected = $state<DiagramSelectionState>({});
	selectedLivePosition = $state<XYPosition>();

	containerEl = $state.raw<HTMLElement>(null!);
	addingEntityGhost = $state.raw<SystemTopologyEntity>();

	constructor(containerElFn: () => HTMLElement) {
		watch(containerElFn, ref => { this.containerEl = ref });
		watch(() => this.analysis.analysisData, data => { this.onAnalysisDataUpdate(data) });
	}

	nodes = $state.raw<Node[]>([]);
	edges = $state.raw<Edge[]>([]);

	onAnalysisDataUpdate(data?: SystemAnalysis) {
		if (!data) return;
		const translated = translateSystemAnalysis($state.snapshot(data));
		this.nodes = translated.nodes;
		this.edges = translated.edges;
	}

	getNodesBounds = $state.raw<ReturnType<typeof useSvelteFlow>["getNodesBounds"]>();
	flowStore = $state.raw<ReturnType<typeof useSvelteFlowStore>>();

	onFlowInit() {
		const flow = useSvelteFlow();
		this.getNodesBounds = flow.getNodesBounds;
		this.flowStore = useSvelteFlowStore();
	};

	interactionLocked() {
		return this.flowStore && !this.flowStore.elementsSelectable;
	}

	updateSelectedPosition({ node, edge }: { node?: Node, edge?: Edge }) {
		if (edge) {
			this.selectedLivePosition = this.getNodesBounds?.([edge.source, edge.target]);
		} else if (node) {
			this.selectedLivePosition = node.position;
		} else {
			this.selectedLivePosition = undefined;
		}
	}

	setSelected(state: DiagramSelectionState) {
		this.analysis.contextMenu = {};
		this.selected = state;
		this.updateSelectedPosition(state);
	};

	handleNodeClicked(e: { node: Node, event: MouseEvent | TouchEvent }) {
		if (this.interactionLocked()) return;
		this.setSelected({ node: e.node });
	};

	handleNodeDragStart(e: { targetNode?: Node | null }) {
		const node = !!e.targetNode ? e.targetNode : undefined;
		this.setSelected({ node });
	};

	handleNodeDrag(e: { targetNode?: Node | null }) {
		if (this.selected.node?.id === e.targetNode?.id && e.targetNode) {
			this.updateSelectedPosition({ node: e.targetNode });
		}
	};

	handleNodeDragStop(e: { targetNode?: Node | null }) {
		if (!e.targetNode) return;
		const { analysisNode } = e.targetNode.data as SystemTopologyNodeData;
		if (!analysisNode) return;

		this.analysis.updateNode(analysisNode.id, {
			position: e.targetNode.position,
		});
	};

	setAddingEntityGhost(c?: SystemTopologyEntity) {
		this.addingEntityGhost = c;
	};

	handlePaneClicked({ event }: { event: MouseEvent }) {
		this.setSelected({});

		if (this.addingEntityGhost) {
			event.preventDefault();

			if (!this.containerEl || !("pageX" in event)) return;

			const { x, y } = this.containerEl.getBoundingClientRect();

			const position = { x: event.pageX - x, y: event.pageY - y };
			const knowledgeEntityId = $state.snapshot(this.addingEntityGhost.id);
			this.analysis.addNode({knowledgeEntityId, position, description: ""});
			// TODO: check if success? show pending state?
			this.setAddingEntityGhost();
		}
	};

	handleEdgeClicked({ edge }: { edge: Edge }) {
		if (this.interactionLocked()) return;
		this.setSelected({ edge });
	};

	handleContextMenuEvent(e: { event: MouseEvent, node?: Node, edge?: Edge, nodes?: Node[] }) {
		if (this.interactionLocked()) return;
		if (!this.containerEl) return;

		e.event.preventDefault();

		if (!("pageX" in e.event)) return;

		const containerRect = this.containerEl.getBoundingClientRect();

		this.analysis.contextMenu = {
			diagram: {
				nodeId: e.node?.id,
				edgeId: e.edge?.id,
				clickPos: { x: e.event.pageX, y: e.event.pageY },
				containerRect,
			}
		}
	};

	closeContextMenu() {
		this.analysis.contextMenu = {};
	}

	onEdgeConnect({ source, target }: Connection) {
		// undo auto-created edge, need to confirm via dialog
		this.edges = this.edges.filter(e => (!(e.source === source && e.target === target)));
	}
};

const diagramCtx = new Context<SystemDiagramState>("systemDiagramState");
export const setSystemDiagram = (s: SystemDiagramState) => diagramCtx.set(s);
export const useSystemDiagram = () => diagramCtx.get();
