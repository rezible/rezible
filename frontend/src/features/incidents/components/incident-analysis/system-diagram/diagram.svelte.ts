import { writable } from "svelte/store";

import {
	type Node,
	type Edge,
	MarkerType,
} from "@xyflow/svelte";

import type { ContextMenuProps } from "./ContextMenu.svelte";

type PaneClickEvent = MouseEvent | TouchEvent;
type NodeClickEventDetail = {node: Node, event: PaneClickEvent};

const createDiagramState = () => {
	let containerEl = $state<HTMLElement>();
	let ctxMenuProps = $state<ContextMenuProps>();

	const nodes = writable<Node[]>([
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
	]);

	const edges = writable<Edge[]>([
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
	]);

	const mount = (el: HTMLElement) => {
		containerEl = el;
	}

	const unmount = () => {

	}

	const handleNodeClicked = ({node, event}: NodeClickEventDetail) => {
		console.log("node clicked", node);
	}

	const handleNodeContextMenu = ({node, event}: NodeClickEventDetail) => {
		event.preventDefault();
		const clientX = "clientX" in event ? event.clientX : 0;
		const clientY = "clientY" in event ? event.clientY : 0;

		const containerHeight = containerEl?.clientHeight ?? 0;
		const containerWidth = containerEl?.clientWidth ?? 0;

		ctxMenuProps = {
			nodeId: node.id,
			top: clientY < containerHeight - 200 ? clientY : undefined,
			left: clientX < containerWidth - 200 ? clientX : undefined,
			right: clientX >= containerWidth - 200 ? containerWidth - clientX : undefined,
			bottom: clientY >= containerHeight - 200 ? containerHeight - clientY : undefined,
		};
	};

	const handlePaneClicked = (event: PaneClickEvent) => {
		ctxMenuProps = undefined;
	};

	return {
		mount,
		unmount,
		nodes,
		edges,
		get ctxMenuProps() { return ctxMenuProps },
		handleNodeClicked,
		handleNodeContextMenu,
		handlePaneClicked,
	}
}

export const diagram = createDiagramState();