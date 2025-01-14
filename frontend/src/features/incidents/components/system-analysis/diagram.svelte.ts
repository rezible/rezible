import { onMount } from "svelte";
import { watch } from "runed";
import { writable } from "svelte/store";

import {
	type Node,
	type Edge,
	type SvelteFlow,
	MarkerType,
	type XYPosition,
} from "@xyflow/svelte";

import { listIncidentSystemComponentsOptions, type IncidentSystemComponent, type SystemComponent, type SystemComponentRelationship } from "$lib/api";
import { ContextMenuWidth, ContextMenuHeight, type ContextMenuProps } from "./SystemDiagramContextMenu.svelte";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { incidentCtx } from '$features/incidents/lib/context.ts';

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

type SvelteFlowEvents = SvelteFlow["$$events_def"];
type SvelteFlowContextMenuEvent = SvelteFlowEvents["panecontextmenu"] | SvelteFlowEvents["nodecontextmenu"] | SvelteFlowEvents["edgecontextmenu"] | SvelteFlowEvents["selectioncontextmenu"];

const createDiagramState = () => {
	let incidentId = $state<string>();
	let containerEl = $state<HTMLElement>();
	let ctxMenuProps = $state<ContextMenuProps>();

	const nodes = writable<Node[]>([]);
	const edges = writable<Edge[]>([]);

	const setup = (containerElFn: () => HTMLElement | undefined) => {
		incidentId = incidentCtx.get().id;
		onMount(() => {containerEl = containerElFn()});
		
		const queryClient = useQueryClient();

		const componentsQuery = createQuery(() => ({
			...listIncidentSystemComponentsOptions({path: {id: incidentId ?? ""}}),
			enabled: !!incidentId,
		}), queryClient);

		watch(() => componentsQuery.data, body => {
			const translated = translateIncidentComponents(body ? $state.snapshot(body.data) : []);
			nodes.set(translated.nodes);
			edges.set(translated.edges);
		});
	}

	const handleNodeClicked = (e: SvelteFlowEvents["nodeclick"]) => {
		const { event, node } = e.detail;
		console.log("node clicked", node);
	}

	const handlePaneClicked = (e: SvelteFlowEvents["paneclick"]) => {
		ctxMenuProps = undefined;
	};

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

	return {
		setup,
		get nodes() { return nodes },
		get edges() { return edges },
		get ctxMenuProps() { return ctxMenuProps },
		handleContextMenuEvent,
		handleNodeClicked,
		handlePaneClicked,
	}
}

export const diagram = createDiagramState();