import ELK, { type ELK as ElkInstance, type ElkNode, type ElkExtendedEdge, type ELKConstructorArguments } from "elkjs/lib/elk-api.js";
import type { Point } from "$features/system/lib/system-map/geometry";
import { NODE_WIDTH, NODE_HEIGHT, type FlowNode, type FlowEdge } from "./flow-graph-model";

const elkArgs: ELKConstructorArguments = {
	algorithms: ["layered"],
	workerFactory: () => {
		return new Worker(new URL("./elk.worker.ts", import.meta.url), { type: "module" });
	},
}

const convertFlowNode = (node: FlowNode): ElkNode => 
	({ id: node.id, width: NODE_WIDTH, height: NODE_HEIGHT });

const convertFlowEdge = ({id, source, target}: FlowEdge): ElkExtendedEdge => 
	({ id, sources: [source], targets: [target] });

const doElkLayout = async (elk: ElkInstance, nodes: readonly FlowNode[], edges: readonly FlowEdge[]) => {
	return elk.layout({
		id: "system-map",
		layoutOptions: {
			"elk.algorithm": "layered",
			"elk.direction": "RIGHT",
			"elk.spacing.nodeNode": "32",
			"elk.layered.spacing.nodeNodeBetweenLayers": "100",
		},
		children: nodes.map(convertFlowNode),
		edges: edges.map(convertFlowEdge),
	});
}

/** Owns automatic placement and its worker. Svelte Flow owns edge paths and live dragging. */
export class SystemMapLayoutEngine {
	private elk?: ElkInstance;

	async layout(nodes: readonly FlowNode[], edges: readonly FlowEdge[]) {
		if (!this.elk) this.elk = new ELK(elkArgs);

		const result = await doElkLayout(this.elk, nodes, edges);
		const positions = new Map<string, Point>();
		for (const node of result.children ?? []) {
			const position = { x: node.x ?? 0, y: node.y ?? 0 };
			positions.set(node.id, position);
		}

		return positions;
	}

	dispose() {
		this.elk?.terminateWorker();
		this.elk = undefined;
	}
}
