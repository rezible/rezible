import { getNodesBounds, getViewportForBounds, type Viewport, type XYPosition } from "@xyflow/svelte";
import { watch, type Getter } from "runed";
import type {
	DiagramContextMenu,
	GraphHighlights,
	GraphSelection,
	SystemDiagramEdge,
	SystemDiagramNode,
} from "./types";

type SystemDiagramOptions = {
	selection: Getter<GraphSelection | undefined>;
	highlights?: Getter<GraphHighlights | undefined>;
	select: (selection: GraphSelection, trigger?: HTMLElement) => void;
	onNodeMove?: (id: string, position: XYPosition) => void;
	onContextMenu?: (menu: DiagramContextMenu) => void;
};

type FramingRequest = {
	subjects?: GraphHighlights;
	resolve: () => void;
};

export class SystemDiagramController {
	private options = $state.raw<SystemDiagramOptions>(null!);

	nodes = $state.raw<SystemDiagramNode[]>([]);
	edges = $state.raw<SystemDiagramEdge[]>([]);
	viewport = $state<Viewport>({ x: 0, y: 0, zoom: 1 });
	width = $state(0);
	height = $state(0);

	private attached = $state(false);
	private framing = $state.raw<FramingRequest>();

	constructor(options: SystemDiagramOptions) {
		this.options = options;

		watch(
			() => [this.options.selection(), this.options.highlights?.()] as const,
			() => this.onSelectionUpdated()
		);
		watch(
			() => [this.nodes, this.edges, this.width, this.height, this.attached, this.framing],
			() => this.applyFraming()
		);
	}

	selection = $derived(this.options.selection());
	selectedNode = $derived(this.nodes.find((node) => node.id === this.selection?.nodeId));
	selectedEdge = $derived(this.edges.find((edge) => edge.id === this.selection?.edgeId));

	highlights = $derived(this.options.highlights?.());

	setGraph(nodes: SystemDiagramNode[], edges: SystemDiagramEdge[]) {
		const previous = new Map(this.nodes.map((node) => [node.id, node]));
		this.nodes = nodes.map((node) => ({
			...node,
			measured: previous.get(node.id)?.measured,
		}));
		this.edges = edges;
		this.onSelectionUpdated();
	}

	private onSelectionUpdated() {
		const nodeIds = new Set(this.highlights?.nodeIds);
		const edgeIds = new Set(this.highlights?.edgeIds);
		this.nodes = this.nodes.map((node) => ({
			...node,
			selected: node.id === this.selection?.nodeId,
			data: { ...node.data, highlighted: nodeIds.has(node.id) },
		}));
		this.edges = this.edges.map((edge) => ({
			...edge,
			selected: edge.id === this.selection?.edgeId,
			data: { ...edge.data!, highlighted: edgeIds.has(edge.id) },
		}));
	}

	select = (selection: GraphSelection, event?: Event) => {
		let trigger: HTMLElement | undefined;
		if (event?.target instanceof Element) {
			trigger =
				event.target.closest<HTMLElement>(".svelte-flow__node, .svelte-flow__edge") ?? undefined;
		}
		this.options.select(selection, trigger);
	};

	moveNode = (id: string, position: XYPosition) => {
		this.options.onNodeMove?.(id, { ...position });
	};

	contextMenu = (selection: GraphSelection, event: MouseEvent) => {
		if (!this.options.onContextMenu) return;
		event.preventDefault();
		const position = { x: event.clientX, y: event.clientY };
		this.options.onContextMenu({ selection, position });
	};

	keydown = (event: KeyboardEvent) => {
		if (!["Enter", " ", "Escape"].includes(event.key)) return;
		const target = event.target;
		if (!(target instanceof Element) || !target.matches(".svelte-flow__node, .svelte-flow__edge")) {
			return;
		}
		const id = target.getAttribute("data-id");
		if (!id) return;
		event.preventDefault();
		event.stopPropagation();
		if (event.key === "Escape") {
			this.select({}, event);
		} else if (target.matches(".svelte-flow__node")) {
			this.select({ nodeId: id }, event);
		} else {
			this.select({ edgeId: id }, event);
		}
	};

	attach = () => {
		this.attached = true;
	};

	detach = () => {
		this.attached = false;
		this.framing?.resolve();
		this.framing = undefined;
	};

	fit = () => this.requestFraming();

	focus = (subjects: GraphHighlights) => this.requestFraming(subjects);

	private requestFraming(subjects?: GraphHighlights) {
		this.framing?.resolve();
		return new Promise<void>((resolve) => {
			this.framing = { subjects, resolve };
		});
	}

	private applyFraming() {
		const request = this.framing;
		if (!request || !this.attached || this.width <= 0 || this.height <= 0) {
			return;
		}
		const ids = request.subjects ? new Set(request.subjects.nodeIds) : undefined;
		for (const edge of this.edges) {
			if (request.subjects?.edgeIds.includes(edge.id)) {
				ids?.add(edge.source);
				ids?.add(edge.target);
			}
		}
		const nodes = ids ? this.nodes.filter((node) => ids.has(node.id)) : this.nodes;
		if (!nodes.length || nodes.some((node) => !node.measured?.width || !node.measured?.height)) {
			return;
		}
		const bounds = getNodesBounds(nodes);
		this.viewport = getViewportForBounds(bounds, this.width, this.height, 0.01, 1.25, 0.3);
		this.framing = undefined;
		request.resolve();
	}
}
