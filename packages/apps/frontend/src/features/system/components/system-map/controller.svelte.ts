import { onMount } from "svelte";
import { Context, watch } from "runed";
import {
	detailBandFromZoom,
	resolveDetailBand,
	runtimeThreshold,
	type DetailBand,
	type DetailMode,
} from "$features/system/lib/system-map/detail";
import type { Box } from "$features/system/lib/system-map/geometry";
import type { RelationshipMode } from "$features/system/lib/system-map/rollup";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { selectedEntityIds, selectedRelationshipIds } from "$features/system/lib/system-map/selection";
import { applyMapState, projectMap, type FlowEdge, type FlowNode, type MapProjection } from "./projection";

type Options = {
	graph: () => GraphSubset;
	selection: () => MapSelection | undefined;
	highlights: () => MapHighlights | undefined;
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
	onExploreDeployment: () => ((targetId: string) => void) | undefined;
};

export type MapViewportFitOptions = {
	preserveZoom?: boolean;
	/** The least zoom the fit may choose, so that the band showing the targets is reached. */
	minZoom?: number;
};
export type MapViewportApi = {
	fit: (ids?: readonly string[], options?: MapViewportFitOptions) => Promise<void>;
};

/** Projects the graph once per publication and applies interaction state over the fixed geometry. */
export class SystemMapController {
	nodes = $state.raw<FlowNode[]>([]);
	edges = $state.raw<FlowEdge[]>([]);

	detailMode = $state<DetailMode>("auto");
	relationshipMode = $state<RelationshipMode>("focused");
	band = $state<DetailBand>("landscape");
	/** The viewport zoom, for drawing labels at a constant size on screen. */
	zoom = $state(1);
	/** The extent of the drawn map; undefined while nothing is drawn. */
	bounds = $state.raw<Box>();
	/** The zoom at which the whole map fits; the runtime threshold is relative to it. */
	private fitZoom = 1;
	private autoBand: DetailBand = "landscape";
	private projection?: MapProjection;
	private viewportApi?: MapViewportApi;
	private hoveredEdgeId?: string;
	private mounted = false;
	private options: Options;

	constructor(options: Options) {
		this.options = options;

		watch(
			() => this.options.graph(),
			() => this.syncGraph()
		);

		watch(
			() => [this.options.selection(), this.options.highlights()] as const,
			() => this.syncMapState()
		);

		onMount(() => {
			this.mounted = true;
			this.syncGraph();

			return () => {
				this.mounted = false;
			};
		});
	}

	private syncGraph() {
		if (!this.mounted) {
			return;
		}

		this.projection = projectMap(this.options.graph());
		this.hoveredEdgeId = undefined;
		this.syncMapState();

		if (this.projection.nodes.length > 0) {
			this.bounds = this.projection.bounds;
		} else {
			this.bounds = undefined;
		}
	}

	private syncMapState() {
		if (!this.projection) return;
		const updated = applyMapState(this.projection, {
			graph: this.options.graph(),
			band: this.band,
			relationshipMode: this.relationshipMode,
			selection: this.options.selection(),
			highlights: this.options.highlights(),
			hoveredEdgeId: this.hoveredEdgeId,
		});
		this.nodes = updated.nodes;
		this.edges = updated.edges;
	}

	setRelationshipMode = (mode: RelationshipMode) => {
		this.relationshipMode = mode;
		this.hoveredEdgeId = undefined;
		this.syncMapState();
	};

	setDetailMode = (mode: DetailMode) => {
		this.detailMode = mode;
		this.syncDetailBand();
	};

	viewportZoom = (zoom: number) => {
		this.zoom = zoom;
		this.autoBand = detailBandFromZoom(zoom, this.fitZoom, this.autoBand);
		this.syncDetailBand();
	};

	/** Records the whole-map fit zoom, which moves the runtime threshold with it. */
	setFitZoom = (fitZoom: number) => {
		this.fitZoom = fitZoom;
		this.autoBand = detailBandFromZoom(this.zoom, fitZoom, this.autoBand);
		this.syncDetailBand();
	};

	private syncDetailBand() {
		const band = resolveDetailBand(this.detailMode, this.autoBand);
		if (band === this.band) return;
		this.band = band;
		this.hoveredEdgeId = undefined;
		this.syncMapState();
	}

	hoverEdge = (id: string | undefined) => {
		this.hoveredEdgeId = id;
		this.syncMapState();
	};

	setViewportApi = (api: MapViewportApi | undefined) => {
		this.viewportApi = api;
	};

	fit = () => {
		this.viewportApi?.fit();
	};

	/**
	 * The nodes that stand for a selection, hidden or not: the endpoints of the connections drawn for its
	 * relationships, else its canonical endpoints or entity.
	 */
	private targetNodeIds(selection: MapSelection | undefined): string[] {
		const relationshipIds = new Set(selectedRelationshipIds(selection));
		const endpoints = new Set<string>();
		for (const edge of this.edges) {
			if (edge.data!.connection.relationshipIds.some((id) => relationshipIds.has(id))) {
				endpoints.add(edge.source);
				endpoints.add(edge.target);
			}
		}
		if (endpoints.size > 0) {
			return [...endpoints];
		}

		const ids = selectedEntityIds(this.options.graph(), selection);
		return ids.filter((id) => this.nodes.some((node) => node.id === id));
	}

	recenter = () => {
		const ids = this.targetNodeIds(this.options.selection());
		if (ids.length > 0) {
			this.viewportApi?.fit(ids, { preserveZoom: true });
			return;
		}
		this.viewportApi?.fit(undefined, { preserveZoom: true });
	};

	reveal = (selection: MapSelection) => {
		const ids = this.targetNodeIds(selection);
		if (ids.length === 0) {
			return;
		}

		// Cards show from the runtime band, so revealing one zooms in past its threshold unless a band is pinned.
		const revealsCard = this.nodes.some(
			(node) => ids.includes(node.id) && node.type === "system-map-card"
		);
		if (revealsCard && this.detailMode === "auto") {
			this.viewportApi?.fit(ids, { minZoom: runtimeThreshold(this.fitZoom) * 1.05 });
			return;
		}
		this.viewportApi?.fit(ids);
	};

	/** Whether the host offers a deployment view to explore a placement target in. */
	get canExploreDeployment(): boolean {
		return this.options.onExploreDeployment() !== undefined;
	}

	exploreDeployment = (targetId: string) => {
		this.options.onExploreDeployment()?.(targetId);
	};

	selectEntity = (entityId: string) => {
		this.options.onSelectionChange({ kind: "entity", entityId });
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

	keydown = (event: KeyboardEvent) => {
		// Escape clears the selection from buttons on the canvas, such as a chip whose popover just closed.
		if (event.target instanceof Element) {
			const textInput = event.target.closest("input, textarea, select, [contenteditable=true]");
			if (textInput) return;
		}

		if (event.key === "Escape") {
			this.clearSelection();
			return;
		}

		if (event.target instanceof Element && event.target.closest("button")) {
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

const ctx = new Context<SystemMapController>("SystemMapController");
export const setSystemMapController = (controller: SystemMapController) => ctx.set(controller);
/** For node components, which Svelte Flow renders inside the map. */
export const useSystemMapController = () => ctx.get();
