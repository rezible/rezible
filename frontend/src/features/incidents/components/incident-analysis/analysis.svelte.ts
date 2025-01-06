import { writable } from "svelte/store";
import {
	type Node,
	type Edge,
	MarkerType,
} from "@xyflow/svelte";

const createSystemViewState = () => {
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

	return {
		nodes,
		edges,
	}
}
export const systemView = createSystemViewState();