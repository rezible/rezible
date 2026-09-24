import { onMount, tick } from "svelte";
import { watch } from "runed";
import type { NodeEvents } from "@xyflow/svelte";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { selectedEntityIds } from "$features/system/lib/system-map/selection";
import { SystemMapLayoutEngine } from "./layout-engine";
import { applyInteractionState } from "./interaction-state";
import { graphToFlow, validPosition, type FlowNode, type FlowEdge } from "./flow-graph-model";

type Options = {
	graph: () => GraphSubset;
	positions: () => Readonly<Partial<Record<string, Point>>> | undefined;
	selection: () => MapSelection | undefined;
	highlights: () => MapHighlights | undefined;
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
	onNodeMove: () => ((entityId: string, position: Point) => void) | undefined;
};

type DragEvent = Parameters<NonNullable<NodeEvents<FlowNode>["onnodedragstop"]>>[0];

export type MapViewportFitOptions = {
	preserveZoom?: boolean;
};
export type MapViewportApi = {
	fit: (ids?: readonly string[], options?: MapViewportFitOptions) => Promise<void>;
};

/** Owns graph synchronization and user commands. Svelte Flow owns live node geometry. */
export class SystemMapController {
	nodes = $state.raw<FlowNode[]>([]);
	edges = $state.raw<FlowEdge[]>([]);
	loading = $state(false);
	error = $state<string>();
	showLabels = $state(false);

	private viewportApi?: MapViewportApi;

	private layoutEngine = new SystemMapLayoutEngine();
	private hoveredEdgeId?: string;
	private mounted = false;
	private revision = 0;
	private dragStart?: { id: string; position: Point };
	private options: Options;

	get nodesDraggable() {
		return Boolean(this.options.onNodeMove());
	}

	constructor(options: Options) {
		this.options = options;

		watch(
			() => this.options.graph(),
			() => this.syncGraph()
		);

		watch(
			() => this.options.positions(),
			(positions, previous) => {
				this.nodes = this.nodes.map((node) => {
					const point = positions?.[node.id];
					const old = previous?.[node.id];
					if (!validPosition(point)) {
						return node;
					}

					if (point.x === old?.x && point.y === old?.y) {
						return node;
					}

					return { ...node, position: { ...point } };
				});
			}
		);

		watch(
			() => [this.options.selection(), this.options.highlights()] as const,
			() => this.syncInteractionState()
		);

		onMount(() => {
			this.mounted = true;
			this.syncGraph();

			return () => {
				this.mounted = false;
				this.cancelLayout();
				this.layoutEngine.dispose();
			};
		});
	}

	private cancelLayout() {
		this.revision += 1;
		this.loading = false;
	}

	private buildModel = () => graphToFlow(this.options.graph(), this.nodes, this.options.positions());

	private syncGraph() {
		if (!this.mounted) {
			return;
		}

		this.cancelLayout();

		const model = this.buildModel();
		if (this.nodes.length === 0 && model.nodes.length > 0) {
			// Keep staging positions off screen while the worker places the initial graph.
			void this.runLayout(model);
			return;
		}

		this.nodes = model.nodes;
		this.edges = model.edges;
		this.hoveredEdgeId = undefined;
		this.syncInteractionState();
	}

	private async runLayout(model = this.buildModel()) {
		if (model.nodes.length === 0 || !this.mounted) {
			return;
		}

		this.revision += 1;
		const revision = this.revision;

		this.loading = true;
		this.error = undefined;

		try {
			// Layout runs once for the initial graph. Pan and zoom only change the viewport.
			const positions = await this.layoutEngine.layout(model.nodes, model.edges);
			if (!this.mounted || revision !== this.revision) {
				return;
			}

			this.nodes = model.nodes.map((node) => {
				const suppliedPosition = this.options.positions()?.[node.id];
				let position = positions.get(node.id);
				if (validPosition(suppliedPosition)) {
					position = suppliedPosition;
				}

				if (!position) {
					return node;
				}

				return { ...node, position };
			});

			this.edges = model.edges;
			this.hoveredEdgeId = undefined;
			this.syncInteractionState();
			await tick();

			if (this.mounted && revision === this.revision) {
				await this.viewportApi?.fit();
				await tick();
			}
		} catch (error) {
			if (revision !== this.revision) {
				return;
			}

			// A failed worker still leaves an inspectable graph and an explicit retry.
			this.nodes = model.nodes;
			this.edges = model.edges;
			this.hoveredEdgeId = undefined;
			this.syncInteractionState();
			this.error = "Layout failed";
			if (error instanceof Error) {
				this.error = error.message;
			}
		} finally {
			if (revision === this.revision) {
				this.loading = false;
			}
		}
	}

	private syncInteractionState() {
		const updated = applyInteractionState(this.nodes, this.edges, {
			graph: this.options.graph(),
			selection: this.options.selection(),
			highlights: this.options.highlights(),
			hoveredEdgeId: this.hoveredEdgeId,
			showLabels: this.showLabels,
		});
		this.nodes = updated.nodes;
		this.edges = updated.edges;
	}

	hoverEdge = (id: string | undefined) => {
		this.hoveredEdgeId = id;
		this.syncInteractionState();
	};

	setViewportApi = (api: MapViewportApi | undefined) => {
		this.viewportApi = api;
	};

	fit = () => {
		this.viewportApi?.fit();
	};

	private visibleSelectionIds(selection: MapSelection | undefined) {
		const ids = selectedEntityIds(this.options.graph(), selection);
		return ids.filter((id) => this.nodes.some((node) => node.id === id));
	}

	recenter = () => {
		const ids = this.visibleSelectionIds(this.options.selection());
		if (ids.length > 0) {
			this.viewportApi?.fit(ids, { preserveZoom: true });
			return;
		}
		this.viewportApi?.fit(undefined, { preserveZoom: true });
	};

	reveal = (selection: MapSelection) => {
		const ids = this.visibleSelectionIds(selection);
		if (ids.length > 0) {
			this.viewportApi?.fit(ids);
		}
	};

	retry = () => {
		void this.runLayout();
	};

	toggleLabels = () => {
		this.showLabels = !this.showLabels;
		this.syncInteractionState();
	};

	selectNode = (id: string, event: Event) => {
		this.select({ kind: "entity", entityId: id }, event);
	};

	selectEdge = (id: string, event: Event) => {
		const connection = this.edges.find((edge) => edge.id === id)?.data?.connection;
		if (!connection) {
			return;
		}
		const relationshipIds = connection.relationshipIds;
		if (relationshipIds.length === 1) {
			this.select({ kind: "relationship", relationshipId: relationshipIds[0] }, event);
			return;
		}
		this.select({ kind: "summary", relationshipIds }, event);
	};

	clearSelection = () => {
		this.options.onSelectionChange(undefined);
	};

	private select(selection: MapSelection, { target }: Event) {
		let trigger: HTMLElement | undefined;
		if (target instanceof Element) {
			trigger = target.closest<HTMLElement>(".svelte-flow__node, .svelte-flow__edge") ?? undefined;
		}

		this.options.onSelectionChange(selection, trigger);
	}

	onNodeDragStart = ({ targetNode }: DragEvent) => {
		this.cancelLayout();
		if (!targetNode) {
			return;
		}

		this.dragStart = {
			id: targetNode.id,
			position: { ...targetNode.position },
		};
	};

	onNodeDragStop = ({ targetNode }: DragEvent) => {
		const start = this.dragStart;
		this.dragStart = undefined;

		if (!targetNode || !start || targetNode.id !== start.id) {
			return;
		}

		const nodeStillExists = this.nodes.some((node) => node.id === targetNode.id);
		if (!nodeStillExists) {
			return;
		}

		const position = { ...targetNode.position };
		const moved = position.x !== start.position.x || position.y !== start.position.y;
		if (moved) {
			this.options.onNodeMove()?.(targetNode.id, position);
		}
	};

	keydown = (event: KeyboardEvent) => {
		if (event.target instanceof Element) {
			const input = event.target.closest("input, textarea, select, button, [contenteditable=true]");
			if (input) return;
		}

		if (event.key === "Escape") {
			this.clearSelection();
			return;
		}

		const isActivationKey = event.key === "Enter" || event.key === " ";
		if (!isActivationKey || !(event.target instanceof Element)) {
			return;
		}

		const target = event.target;
		const node = target.closest<HTMLElement>(".svelte-flow__node");
		const edge = target.closest<HTMLElement>(".svelte-flow__edge");
		const id = (node ?? edge)?.getAttribute("data-id");
		if (!id) return;

		event.preventDefault();
		if (node) {
			this.selectNode(id, event);
		} else {
			this.selectEdge(id, event);
		}
	};
}
