import { Context, watch } from "runed";

import {
	useSvelteFlow,
	useStore as useSvelteFlowStore,
	type Node,
	type Edge,
	type XYPosition,
	type Connection,
	type Rect as BoundsRect
} from "@xyflow/svelte";

import { updateSystemAnalysisComponentMutation, type SystemAnalysis, type SystemAnalysisComponent, type SystemAnalysisRelationship, type SystemComponent } from "$lib/api";

import { createMutation } from "@tanstack/svelte-query";
import { useIncidentAnalysis } from "../analysisState.svelte";
import { useRelationshipDialog } from "./relationship-dialog/dialogState.svelte";
import { useComponentDialog } from "./component-dialog/dialogState.svelte";
import ContextMenu from "./ContextMenu.svelte";
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
	let edges: Edge[] = [];

	an.attributes.components.forEach(analysisComponent => {
		const { position, component } = analysisComponent.attributes;

		const data: SystemComponentNodeData = {analysisComponent};

		nodes.push({
			id: component.id,
			type: "component",
			data,
			position,
		});
	});

	an.attributes.relationships.forEach((relationship) => {
		const { id, attributes } = relationship;

		const data: SystemRelationshipEdgeData = {
			relationship,
		};

		edges.push({
			id,
			type: "relationship",
			source: attributes.sourceId,
			target: attributes.targetId,
			data,
		});
	});

	return { nodes, edges };
};

type DiagramSelectionState = { node?: Node; edge?: Edge };

export class SystemDiagramState {
	analysis = useIncidentAnalysis();

	relationshipDialog = useRelationshipDialog();
	componentDialog = useComponentDialog();

	selected = $state<DiagramSelectionState>({});
	selectedLivePosition = $state<XYPosition>();

	containerEl = $state<HTMLElement>();
	ctxMenuProps = $state.raw<ComponentProps<typeof ContextMenu>>();
	addingComponent = $state<SystemComponent>();

	nodes = $state.raw<Node[]>([]);
	edges = $state.raw<Edge[]>([]);

	updateAnalysisComponentMut = createMutation(() => updateSystemAnalysisComponentMutation());

	constructor(containerElFn: () => HTMLElement | undefined) {
		this.componentDialog.onAddComponent = this.setAddingComponent;

		watch(containerElFn, ref => {
			this.containerEl = ref;
		});

		watch(
			() => this.analysis.analysisData,
			(data) => {
				if (!data) return;
				const translated = translateSystemAnalysis($state.snapshot(data));
				this.nodes = translated.nodes;
				this.edges = translated.edges;
			}
		);
	}

	flow = $state<ReturnType<typeof useSvelteFlow>>();
	flowStore = $state<ReturnType<typeof useSvelteFlowStore>>();

	onFlowInit() {
		this.flow = useSvelteFlow();
		this.flowStore = useSvelteFlowStore();
	};

	interactionLocked() {
		return this.flowStore && !this.flowStore.elementsSelectable;
	}

	updateSelectedPosition({node, edge}: {node?: Node, edge?: Edge}) {
		if (this.flow && edge) {
			this.selectedLivePosition = this.flow.getNodesBounds([edge.source, edge.target]);
		} else if (node) {
			this.selectedLivePosition = node.position;
		} else {
			this.selectedLivePosition = undefined;
		}
	}

	setSelected(state: DiagramSelectionState) {
		this.ctxMenuProps = undefined;
		this.selected = state;
		this.updateSelectedPosition(state);
	};

	handleNodeClicked(e: {node: Node, event: MouseEvent | TouchEvent}) {
		if (this.interactionLocked()) return;
		this.setSelected({ node: e.node });
	};

	handleNodeDragStart(e: {targetNode?: Node | null}) {
		const node = !!e.targetNode ? e.targetNode : undefined;
		this.setSelected({ node });
	};

	handleNodeDrag(e: {targetNode?: Node | null}) {
		if (this.selected.node?.id === e.targetNode?.id && e.targetNode) {
			this.updateSelectedPosition({node: e.targetNode});
		}
	};

	handleNodeDragStop(e: {targetNode?: Node | null}) {
		if (!e.targetNode) return;
		const {analysisComponent} = e.targetNode.data as SystemComponentNodeData;
		const attributes = {position: e.targetNode.position};
		this.updateAnalysisComponentMut.mutate({path: {id: analysisComponent.id}, body: {attributes}});
	};

	setAddingComponent(c?: SystemComponent) {
		this.addingComponent = c;
	};

	handlePaneClicked({event}: {event: MouseEvent}) {
		this.setSelected({});

		if (this.addingComponent) {
			event.preventDefault();

			if (!this.containerEl || !("pageX" in event)) return;

			const { x, y } = this.containerEl.getBoundingClientRect();

			const pos = {x: event.pageX - x, y: event.pageY - y};
			const component = $state.snapshot(this.addingComponent);
			this.analysis?.addComponent(component, pos);
			// TODO: check if success? show pending state?
			this.setAddingComponent();
		}
	};

	handleEdgeClicked({edge}: {edge: Edge}) {
		if (this.interactionLocked()) return;
		this.setSelected({ edge });
	};

	handleContextMenuEvent(e: {event: MouseEvent, node?: Node, edge?: Edge, nodes?: Node[]}) {
		if (this.interactionLocked()) return;
		if (!this.containerEl) return;

		e.event.preventDefault();

		if (!("pageX" in e.event)) return;

		const containerRect = this.containerEl.getBoundingClientRect();

		this.ctxMenuProps = {
			nodeId: e.node?.id,
			edgeId: e.edge?.id,
			clickPos: {x: e.event.pageX, y: e.event.pageY},
			containerRect,
		};
	};

	closeContextMenu() {
		this.ctxMenuProps = undefined;
	}

	onEdgeConnect({source, target}: Connection) {
		this.edges = this.edges.filter(e => (!(e.source === source && e.target === target)));
		this.relationshipDialog.setCreating(source, target);
	}
};

const diagramCtx = new Context<SystemDiagramState>("systemDiagramState");
export const setSystemDiagram = (s: SystemDiagramState) => diagramCtx.set(s);
export const useSystemDiagram = () => diagramCtx.get();
