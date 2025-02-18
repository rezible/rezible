import { onMount } from "svelte";
import { watch } from "runed";
import { writable, get } from "svelte/store";

import {
	type Node,
	type Edge,
	MarkerType,
	SvelteFlow,
	type XYPosition,
	useSvelteFlow,
	type OnConnectEnd,
	type OnConnect,
	type Connection,
	useStore,
} from "@xyflow/svelte";

import { ContextMenuWidth, ContextMenuHeight, type ContextMenuProps } from "./ContextMenu.svelte";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { incidentCtx } from "$features/incidents/lib/context.ts";
import {
	getSystemAnalysisOptions,
	type SystemAnalysis,
	type SystemAnalysisComponent,
	type SystemAnalysisRelationship,
	type SystemAnalysisRelationshipAttributes,
	type SystemComponent,
} from "$lib/api";
import { analysis } from "../analysis.svelte";
import { relationshipDialog } from "../relationship-dialog/dialogState.svelte";

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
	component: SystemAnalysisComponent;
};

export type SystemRelationshipEdgeData = {
	relationship: SystemAnalysisRelationship;
};

const translateSystemAnalysis = (an: SystemAnalysis) => {
	let nodes: Node[] = [];
	let edges: Edge[] = [];

	an.attributes.components.forEach((component) => {
		const { position } = component.attributes;

		const data: SystemComponentNodeData = {
			component,
		};

		nodes.push({
			id: component.attributes.component.id,
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

const createDiagramState = () => {
	let selected = $state<DiagramSelectionState>({});
	let toolbarPosition = $state<XYPosition>({ x: 0, y: 0 });
	let containerEl = $state<HTMLElement>();
	let ctxMenuProps = $state<ContextMenuProps>();

	const nodes = writable<Node[]>([]);
	const edges = writable<Edge[]>([]);

	const setup = (containerElFn: () => HTMLElement | undefined) => {
		onMount(() => {
			containerEl = containerElFn();
		});

		watch(
			() => analysis.data,
			(data) => {
				if (!data) return;
				const translated = translateSystemAnalysis($state.snapshot(data));
				nodes.set(translated.nodes);
				edges.set(translated.edges);
			}
		);
	};

	let flow = $state<ReturnType<typeof useSvelteFlow>>();
	let flowStore = $state<ReturnType<typeof useStore>>();
	const onFlowInit = () => {
		flow = useSvelteFlow();
		flowStore = useStore();
	};

	const interactionLocked = () => {
		return flowStore && !get(flowStore.elementsSelectable);
	}

	const updateToolbarPosition = () => {
		if (!flow) return;
		const { node, edge } = selected;
		if (edge) {
			const { x, y, width, height } = flow.getNodesBounds([edge.source, edge.target]);
			toolbarPosition = {
				x: x + width / 2,
				y: y + height / 2,
			};
		} else if (node) {
			const { x, y, width, height } = flow.getNodesBounds([node]);
			toolbarPosition = {
				x: x + width / 2,
				y: y + height + 25,
			};
		} else {
			toolbarPosition = { x: 0, y: 0 };
		}
	};

	const setSelected = (state: DiagramSelectionState) => {
		ctxMenuProps = undefined;
		selected = state;
		updateToolbarPosition();
	};

	const handleNodeClicked = (e: SvelteFlowEvents["nodeclick"]) => {
		if (interactionLocked()) return;
		const { event, node } = e.detail;
		setSelected({ node });
	};

	const handleNodeDragStart = (e: SvelteFlowEvents["nodedragstart"]) => {
		setSelected({ node: e.detail.targetNode ?? undefined });
	};

	const handleNodeDrag = (e: SvelteFlowEvents["nodedrag"]) => {
		if (selected.node?.id === e.detail.targetNode?.id) {
			updateToolbarPosition();
		}
	};

	const handlePaneClicked = (e: SvelteFlowEvents["paneclick"]) => {
		setSelected({});

		if (analysis.addingComponent) {
			const event = e.detail.event;
			event.preventDefault();

			if (!containerEl || !("pageX" in event)) return;

			const { x, y } = containerEl.getBoundingClientRect();

			const posX = event.pageX - x;
			const posY = event.pageY - y;

			alert(`add component: [${posX}, ${posY}]`);

			// createSystemAnalysisComponent mutation, then invalidate 

			analysis.setAddingComponent();
		}
	};

	const handleEdgeClicked = (e: SvelteFlowEvents["edgeclick"]) => {
		if (interactionLocked()) return;
		const { event, edge } = e.detail;
		setSelected({ edge });
	};

	const handleContextMenuEvent = (e: SvelteFlowContextMenuEvent) => {
		if (interactionLocked()) return;
		if (!containerEl) return;

		const detail = e.detail;
		const event = detail.event;
		event.preventDefault();

		if (!("pageX" in event)) return;

		const { x, y, width, height } = containerEl.getBoundingClientRect();

		const posX = event.pageX - x;
		const posY = event.pageY - y;

		const boundLeft = width - ContextMenuWidth;
		const boundTop = height - ContextMenuHeight;

		ctxMenuProps = {
			nodeId: "node" in detail ? detail.node?.id : undefined,
			edgeId: "edge" in detail ? detail.edge?.id : undefined,
			top: posY > boundTop ? height - ContextMenuHeight : posY,
			left: posX > boundLeft ? posX - ContextMenuWidth : posX,
		};
	};

	const onEdgeConnect = ({source, target}: Connection) => {
		edges.set(get(edges).filter(e => (!(e.source === source && e.target === target))));
		relationshipDialog.setCreating(source, target);
	}

	return {
		setup,
		onFlowInit,
		get nodes() {
			return nodes;
		},
		get edges() {
			return edges;
		},
		get selected() {
			return selected;
		},
		get toolbarPosition() {
			return toolbarPosition;
		},
		get ctxMenuProps() {
			return ctxMenuProps;
		},
		handleContextMenuEvent,
		handleNodeClicked,
		handleNodeDragStart,
		handleNodeDrag,
		handleEdgeClicked,
		handlePaneClicked,
		onEdgeConnect,
	};
};

export const diagram = createDiagramState();
