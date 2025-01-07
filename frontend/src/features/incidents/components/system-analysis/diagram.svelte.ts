import { onMount } from "svelte";
import { watch } from "runed";
import { writable } from "svelte/store";

import {
	type Node,
	type Edge,
	MarkerType,
} from "@xyflow/svelte";

import { listIncidentSystemComponentsOptions, type IncidentSystemComponent, type SystemComponent } from "$lib/api";
import { ContextMenuWidth, ContextMenuHeight, type ContextMenuProps } from "./SystemDiagramContextMenu.svelte";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { incidentCtx } from '$features/incidents/lib/context.ts';

const translateSystemComponents = (components: IncidentSystemComponent[]) => {
	const nodes: Node[] = [
		{
			id: "service-1",
			data: { label: "API Service" },
			position: { x: 0, y: 0 },
		},
		{
			id: "control-1",
			data: { label: "Rate Limiter" },
			position: { x: 100, y: 100 },
		},
	];

	const edges: Edge[] = [
		{
			id: "e1",
			source: "service-1",
			target: "control-1",
			type: "control",
			label: "Controls",
			markerEnd: {
				type: MarkerType.ArrowClosed,
			},
		},
	];

	return {nodes, edges};
}

type PaneClickEvent = MouseEvent | TouchEvent;
type NodeClickEventDetail = {node: Node, event: PaneClickEvent};

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

	const createQueries = () => {
		const queryClient = useQueryClient();
		const incidentId = incidentCtx.get().id;

		const milestonesQuery = createQuery(() => listIncidentSystemComponentsOptions({path: {id: incidentId}}), queryClient);
		watch(() => milestonesQuery, r => onIncidentSystemComponentsUpdated(r.data?.data ?? []));
	}

	const handleNodeClicked = ({node, event}: NodeClickEventDetail) => {
		console.log("node clicked", node);
	}

	const handleContextMenu = (event: PaneClickEvent, node?: Node) => {
		if (!containerEl) return;
		if (!("clientX" in event)) {
			return;
		}
		event.preventDefault();

		const {x, y, width, height} = containerEl.getBoundingClientRect();

		const posX = event.pageX - x;
		const posY = event.pageY - y;

		const boundLeft = width - ContextMenuWidth;
		const boundTop = height - ContextMenuHeight;
		
		ctxMenuProps = {
			nodeId: node?.id,
			top: posY > boundTop ? (height - ContextMenuHeight) : posY,
			left: posX > boundLeft ? (posX - ContextMenuWidth) : posX,
		}
	};

	const handlePaneClicked = (event: PaneClickEvent) => {
		ctxMenuProps = undefined;
	};

	const componentSetup = (containerElFn: () => HTMLElement | undefined) => {
		onMount(() => {containerEl = containerElFn()});
		createQueries();
	}

	return {
		componentSetup,
		nodes,
		edges,
		get ctxMenuProps() { return ctxMenuProps },
		handleContextMenu,
		handleNodeClicked,
		handlePaneClicked,
	}
}

export const diagram = createDiagramState();