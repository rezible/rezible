import { Context, watch } from "runed";

import {
	useSvelteFlow,
	useStore as useSvelteFlowStore,
	type Node,
	type Edge,
	type XYPosition,
	type Connection,
} from "@xyflow/svelte";

import { type SystemAnalysis, type SystemAnalysisComponent, type SystemAnalysisRelationship, type SystemComponent } from "$lib/api";

import { useIncidentAnalysis } from "../analysisState.svelte";
import ContextMenu from "./SystemDiagramContextMenu.svelte";
import type { ComponentProps } from "svelte";

/*
const convertRelationshipToEdge = ({id, attributes}: SystemComponentRelationship): Edge => {
	const {kind, details} = attributes;
	let source = "", target = "", label = "";
	if (kind === "control") {
		source = details.controllerId;
		target = details.controlledId;
		label = details.control;
	} else if (kind === "feedback") {
		source = details.sourceId;
		target = details.targetId;
		label = details.feedback;
	}
	return {
		type: attributes.kind,
		id,
		source,
		target,
		label,
		markerEnd: {
			type: MarkerType.ArrowClosed
		}
	};
};

const translateIncidentComponents = (incidentComponents: IncidentSystemComponent[]) => {
	const positions = new Map<string, XYPosition>();

	let nodes: Node[] = [];
	let edges: Edge[] = [];

	let nextPos = {x: 0, y: 0};
	const relationships = new Map<string, SystemComponentRelationship>();
	incidentComponents.forEach(ic => {
		const component = ic.attributes.component;
		const id = component.id;

		let position = positions.get(id);
		if (!position) {
			position = {x: nextPos.x, y: nextPos.y};
			nextPos = {x: nextPos.x + 100, y: nextPos.y + 100};
		}

		nodes.push({
			id: id,
			type: component.attributes.kind,
			data: {
				label: component.attributes.name,
				role: ic.attributes.role,
			},
			position,
		});

		component.attributes.relationships.forEach(r => relationships.set(r.id, r));
	});

	relationships.forEach(rel => edges.push(convertRelationshipToEdge(rel)));

	return {nodes, edges};
}
*/

export type SystemComponentNodeData = {
	analysisComponent: SystemAnalysisComponent;
};

export type SystemRelationshipEdgeData = {
	relationship: SystemAnalysisRelationship;
};

const translateSystemAnalysis = (an: SystemAnalysis) => {
	let nodes: Node[] = [];
	an.attributes.components.forEach(analysisComponent => {
		const { position, component } = analysisComponent.attributes;
		nodes.push({
			id: component.id,
			type: "component",
			position,
			data: { analysisComponent } as SystemComponentNodeData,
		});
	});

	let edges: Edge[] = [];
	an.attributes.relationships.forEach(sr => {
		const { id, attributes } = sr;
		const relattr = attributes.relationship.attributes;
		edges.push({
			id,
			type: "relationship",
			source: relattr.sourceId,
			target: relattr.targetId,
			data: { relationship: sr } as SystemRelationshipEdgeData,
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
	addingComponentGhost = $state.raw<SystemComponent>();

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

	flowStore = $state.raw<ReturnType<typeof useSvelteFlowStore>>();

	onFlowInit() {
		this.flowStore = useSvelteFlowStore();
	};

	interactionLocked() {
		return this.flowStore && !this.flowStore.elementsSelectable;
	}

	updateSelectedPosition({ node, edge }: { node?: Node, edge?: Edge }) {
		const { getNodesBounds } = useSvelteFlow();
		if (edge) {
			this.selectedLivePosition = getNodesBounds([edge.source, edge.target]);
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
		const { analysisComponent } = e.targetNode.data as SystemComponentNodeData;
		if (!analysisComponent) return;

		this.analysis.updateComponent(analysisComponent.id, {
			position: e.targetNode.position,
		});		
	};

	setAddingComponentGhost(c?: SystemComponent) {
		this.addingComponentGhost = c;
	};

	handlePaneClicked({ event }: { event: MouseEvent }) {
		this.setSelected({});

		if (this.addingComponentGhost) {
			event.preventDefault();

			if (!this.containerEl || !("pageX" in event)) return;

			const { x, y } = this.containerEl.getBoundingClientRect();

			const position = { x: event.pageX - x, y: event.pageY - y };
			const componentId = $state.snapshot(this.addingComponentGhost.id);
			this.analysis.addComponent({componentId, position});
			// TODO: check if success? show pending state?
			this.setAddingComponentGhost();
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
