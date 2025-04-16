import { onMount } from "svelte";
import { Context, watch } from "runed";
import { writable, get } from "svelte/store";

import {
	SvelteFlow,
	useSvelteFlow,
	useStore as useSvelteFlowStore,
	type Node,
	type Edge,
	type XYPosition,
	type Connection,
} from "@xyflow/svelte";

import { updateSystemAnalysisComponentMutation, type SystemAnalysis, type SystemAnalysisComponent, type SystemAnalysisRelationship, type SystemComponent } from "$lib/api";

import { ContextMenuWidth, ContextMenuHeight, type ContextMenuProps } from "./ContextMenu.svelte";
import { createMutation } from "@tanstack/svelte-query";
import { useIncidentAnalysis } from "../analysisState.svelte";
import { useRelationshipDialog } from "./relationship-dialog/dialogState.svelte";
import { useComponentDialog } from "./component-dialog/dialogState.svelte";

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

type SvelteFlowEvents = SvelteFlow["$$events_def"];
type SvelteFlowContextMenuEvent =
	| SvelteFlowEvents["panecontextmenu"]
	| SvelteFlowEvents["nodecontextmenu"]
	| SvelteFlowEvents["edgecontextmenu"]
	| SvelteFlowEvents["selectioncontextmenu"];
type DiagramSelectionState = { node?: Node; edge?: Edge };

export class SystemDiagramState {
	analysis = useIncidentAnalysis();

	relationshipDialog = useRelationshipDialog();
	componentDialog = useComponentDialog();

	selected = $state<DiagramSelectionState>({});
	toolbarPosition = $state<XYPosition>({ x: 0, y: 0 });
	containerEl = $state<HTMLElement>();
	ctxMenuProps = $state<ContextMenuProps>();
	addingComponent = $state<SystemComponent>();

	nodes = writable<Node[]>([]);
	edges = writable<Edge[]>([]);

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
				this.nodes.set(translated.nodes);
				this.edges.set(translated.edges);
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
		return this.flowStore && !get(this.flowStore.elementsSelectable);
	}

	updateToolbarPosition() {
		if (!this.flow) return;
		const { node, edge } = this.selected;
		if (edge) {
			const { x, y, width, height } = this.flow.getNodesBounds([edge.source, edge.target]);
			this.toolbarPosition = {
				x: x + width / 2,
				y: y + height / 2,
			};
		} else if (node) {
			const { x, y, width, height } = this.flow.getNodesBounds([node]);
			this.toolbarPosition = {
				x: x + width / 2,
				y: y + height + 25,
			};
		} else {
			this.toolbarPosition = { x: 0, y: 0 };
		}
	};

	setSelected(state: DiagramSelectionState) {
		this.ctxMenuProps = undefined;
		this.selected = state;
		this.updateToolbarPosition();
	};

	handleNodeClicked(e: SvelteFlowEvents["nodeclick"]) {
		if (this.interactionLocked()) return;
		const { event, node } = e.detail;
		this.setSelected({ node });
	};

	handleNodeDragStart(e: SvelteFlowEvents["nodedragstart"]) {
		this.setSelected({ node: e.detail.targetNode ?? undefined });
	};

	handleNodeDrag(e: SvelteFlowEvents["nodedrag"]) {
		if (this.selected.node?.id === e.detail.targetNode?.id) {
			this.updateToolbarPosition();
		}
	};

	handleNodeDragStop(e: SvelteFlowEvents["nodedragstop"]) {
		const node = e.detail.targetNode;
		if (!node) return;
		const {analysisComponent} = node.data as SystemComponentNodeData;
		const attributes = {position: node.position};
		this.updateAnalysisComponentMut.mutate({path: {id: analysisComponent.id}, body: {attributes}});
	};

	setAddingComponent(c?: SystemComponent) {
		this.addingComponent = c;
	};

	handlePaneClicked(e: SvelteFlowEvents["paneclick"]) {
		this.setSelected({});

		if (this.addingComponent) {
			const event = e.detail.event;
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

	handleEdgeClicked(e: SvelteFlowEvents["edgeclick"]) {
		if (this.interactionLocked()) return;
		const { event, edge } = e.detail;
		this.setSelected({ edge });
	};

	handleContextMenuEvent(e: SvelteFlowContextMenuEvent) {
		if (this.interactionLocked()) return;
		if (!this.containerEl) return;

		const detail = e.detail;
		const event = detail.event;
		event.preventDefault();

		if (!("pageX" in event)) return;

		const { x, y, width, height } = this.containerEl.getBoundingClientRect();

		const posX = event.pageX - x;
		const posY = event.pageY - y;

		const boundLeft = Math.max(width - ContextMenuWidth, x);
		const boundTop = Math.max(height - ContextMenuHeight, y);

		const left = posX > boundLeft ? posX - ContextMenuWidth : posX;
		const top = posY;//posY > boundTop ? height - ContextMenuHeight : posY;

		this.ctxMenuProps = {
			nodeId: "node" in detail ? detail.node?.id : undefined,
			edgeId: "edge" in detail ? detail.edge?.id : undefined,
			top,
			left,
			x: posX,
			y: posY,
		};
	};

	closeContextMenu() {
		this.ctxMenuProps = undefined;
	}

	onEdgeConnect({source, target}: Connection) {
		const newEdges = get(this.edges).filter(e => (!(e.source === source && e.target === target)));
		this.edges.set(newEdges);
		this.relationshipDialog.setCreating(source, target);
	}
};

const diagramCtx = new Context<SystemDiagramState>("systemDiagramState");
export const setSystemDiagram = (s: SystemDiagramState) => diagramCtx.set(s);
export const useSystemDiagram = () => diagramCtx.get();
