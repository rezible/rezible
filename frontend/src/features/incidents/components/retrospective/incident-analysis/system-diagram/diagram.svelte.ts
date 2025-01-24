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
} from "@xyflow/svelte";

import { ContextMenuWidth, ContextMenuHeight, type ContextMenuProps } from "./ContextMenu.svelte";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { incidentCtx } from '$features/incidents/lib/context.ts';
import { getSystemAnalysisOptions, type SystemAnalysis, type SystemAnalysisRelationship, type SystemAnalysisRelationshipAttributes, type SystemComponent } from "$lib/api";
import { analysis } from "../analysis.svelte";

/*
const convertRelationshipToEdge = ({id, attributes}: SystemComponentRelationship): Edge => {
	const {kind, details} = attributes;
	let source = "", target = "", label = "";
	if (kind === "control") {
		source = details.controller_id;
		target = details.controlled_id;
		label = details.control;
	} else if (kind === "feedback") {
		source = details.source_id;
		target = details.target_id;
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
	analysisId: string;
	component: SystemComponent;
	role: string;
}

const translateSystemAnalysis = (an: SystemAnalysis) => {
	let nodes: Node[] = [];
	let edges: Edge[] = [];

	an.attributes.components.forEach(({id, attributes}) => {
		const {component, role, position} = attributes;

		const data: SystemComponentNodeData = {
			analysisId: id,
			component,
			role,
		}

		nodes.push({
			id: component.id,
			type: "component",
			data,
			position,
		});
	});

	an.attributes.relationships.forEach(({id, attributes}) => {
		edges.push({
			id,
			type: "relationship",
			source: attributes.source_id,
			target: attributes.target_id,
			data: attributes,
		})
	});

	return {nodes, edges};
}

type SvelteFlowEvents = SvelteFlow["$$events_def"];
type SvelteFlowContextMenuEvent = SvelteFlowEvents["panecontextmenu"] | SvelteFlowEvents["nodecontextmenu"] | SvelteFlowEvents["edgecontextmenu"] | SvelteFlowEvents["selectioncontextmenu"];
type DiagramSelectionState = {node?: Node, edge?: Edge};
const createDiagramState = () => {
	let selected = $state<DiagramSelectionState>({});
	let toolbarPosition = $state<XYPosition>({x: 0, y: 0})
	let containerEl = $state<HTMLElement>();
	let ctxMenuProps = $state<ContextMenuProps>();

	const nodes = writable<Node[]>([]);
	const edges = writable<Edge[]>([]);

	const setup = (containerElFn: () => HTMLElement | undefined) => {
		onMount(() => {containerEl = containerElFn()});

		watch(() => analysis.data, data => {
			if (!data) return;
			const translated = translateSystemAnalysis($state.snapshot(data));
			nodes.set(translated.nodes);
			edges.set(translated.edges);
		});
	};

	let flow: ReturnType<typeof useSvelteFlow> | undefined;
	const onFlowInit = () => {flow = useSvelteFlow()}

	const updateToolbarPosition = () => {
		if (!flow) return;
		const {node, edge} = selected;
		if (edge) {
			const {x, y, width, height} = flow.getNodesBounds([edge.source, edge.target]);
			toolbarPosition = {
				x: x + (width / 2),
				y: y + (height / 2) + 40,
			};
		} else if (node) {
			const {x, y, width, height} = flow.getNodesBounds([node]);
			toolbarPosition = {
				x: x + (width / 2),
				y: y + height + 30,
			};
		} else {
			toolbarPosition = {x: 0, y: 0};
		}
	}

	const setSelected = (state: DiagramSelectionState) => {
		ctxMenuProps = undefined;
		selected = state;
		updateToolbarPosition();
	}

	const handleNodeClicked = (e: SvelteFlowEvents["nodeclick"]) => {
		const { event, node } = e.detail;
		setSelected({node});
	}

	const handleNodeDragStart = (e: SvelteFlowEvents["nodedragstart"]) => {
		setSelected({node: e.detail.targetNode ?? undefined});
	}

	const handleNodeDrag = (e: SvelteFlowEvents["nodedrag"]) => {
		if (selected.node?.id === e.detail.targetNode?.id) {
			updateToolbarPosition();
		}
	}

	const handlePaneClicked = (e: SvelteFlowEvents["paneclick"]) => {
		setSelected({});
	};

	const handleEdgeClicked = (e: SvelteFlowEvents["edgeclick"]) => {
		const { event, edge } = e.detail;
		setSelected({edge});
	}

	const handleContextMenuEvent = (e: SvelteFlowContextMenuEvent) => {
		if (!containerEl) return;

		const detail = e.detail;
		const event = detail.event;
		event.preventDefault();

		if (!("pageX" in event)) return;

		const {x, y, width, height} = containerEl.getBoundingClientRect();

		const posX = event.pageX - x;
		const posY = event.pageY - y;

		const boundLeft = width - ContextMenuWidth;
		const boundTop = height - ContextMenuHeight;

		ctxMenuProps = {
			nodeId: ("node" in detail) ? detail.node?.id : undefined,
			edgeId: ("edge" in detail) ? detail.edge?.id : undefined,
			top: posY > boundTop ? (height - ContextMenuHeight) : posY,
			left: posX > boundLeft ? (posX - ContextMenuWidth) : posX,
		}
	};

	const handleConnectEnd: OnConnectEnd = (event, connectionState) => {
		if (connectionState.isValid) return;
 
		const sourceNodeId = connectionState.fromNode?.id ?? '1';
		const { clientX, clientY } = 'changedTouches' in event ? event.changedTouches[0] : event;
	 
		console.log("dropped", sourceNodeId, clientX, clientY);
	}

	return {
		setup,
		onFlowInit,
		get nodes() { return nodes },
		get edges() { return edges },
		get selected() { return selected },
		get toolbarPosition() { return toolbarPosition },
		get ctxMenuProps() { return ctxMenuProps },
		handleContextMenuEvent,
		handleNodeClicked,
		handleNodeDragStart,
		handleNodeDrag,
		handleEdgeClicked,
		handlePaneClicked,
		handleConnectEnd,
	}
}

export const diagram = createDiagramState();