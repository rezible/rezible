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

const translateSystemComponents = (incidentComponents: IncidentSystemComponent[]) => {
	const positions = new Map<string, XYPosition>();

	console.log("cmp", $state.snapshot(incidentComponents));

	let nodes: Node[] = [];
	let edges: Edge[] = [];

	let nextPos = {x: 0, y: 0};
	const edgeMap = new Map<string, SystemComponentRelationship>();
	incidentComponents.forEach(ic => {
		const component = ic.attributes.component;
		const id = component.id;

		let position = positions.get(id);
		if (!position) {
			position = {x: nextPos.x, y: nextPos.y};
			nextPos = {x: nextPos.x + 100, y: nextPos.y + 100};
		}

		const node: Node = {
			id: id,
			type: component.attributes.kind,
			data: {
				label: component.attributes.name,
				role: ic.attributes.role,
			},
			position,
		}
		nodes.push(node);

		component.attributes.relationships.forEach(r => {
			edgeMap.set(r.id, r);
		});
	});

	edgeMap.forEach(r => {
		let source = "";
		let target = "";
		let label = "";
		if (r.attributes.kind === "control") {
			const details = r.attributes.details;
			source = details.controller_id;
			target = details.controlled_id;
			label = details.control;
		} else if (r.attributes.kind === "feedback") {
			const details = r.attributes.details;
			source = details.source_id;
			target = details.target_id;
			label = details.feedback;
		}
		const edge: Edge = {
			id: r.id,
			type: r.attributes.kind,
			source,
			target,
			label,
			markerEnd: {
				type: MarkerType.ArrowClosed
			}
		};
		edges.push(edge);
	});

	return {nodes, edges};
}

type SvelteFlowEvents = SvelteFlow["$$events_def"];
type SvelteFlowContextMenuEvent = SvelteFlowEvents["panecontextmenu"] | SvelteFlowEvents["nodecontextmenu"] | SvelteFlowEvents["edgecontextmenu"] | SvelteFlowEvents["selectioncontextmenu"];

const createDiagramState = () => {
	let containerEl = $state<HTMLElement>();
	let ctxMenuProps = $state<ContextMenuProps>();

	const nodes = writable<Node[]>([]);
	const edges = writable<Edge[]>([]);

	const onIncidentSystemComponentsUpdated = (components: IncidentSystemComponent[]) => {
		const translated = translateSystemComponents(components);
		nodes.set(translated.nodes);
		edges.set(translated.edges);
	}

	const componentSetup = (containerElFn: () => HTMLElement | undefined) => {
		onMount(() => {containerEl = containerElFn()});
		createQueries();
	}

	const createQueries = () => {
		const queryClient = useQueryClient();
		const incidentId = incidentCtx.get().id;

		const componentsQuery = createQuery(() => listIncidentSystemComponentsOptions({path: {id: incidentId}}), queryClient);
		watch(() => componentsQuery.data, body => {
			const components = body?.data ?? [];
			onIncidentSystemComponentsUpdated(components);
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
		componentSetup,
		nodes,
		edges,
		get ctxMenuProps() { return ctxMenuProps },
		handleContextMenuEvent,
		handleNodeClicked,
		handlePaneClicked,
	}
}

export const diagram = createDiagramState();