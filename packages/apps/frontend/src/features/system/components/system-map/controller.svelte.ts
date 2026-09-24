import { getViewportForBounds, type EdgeEvents, type NodeEvents, type OnMove } from "@xyflow/svelte";
import { watch } from "runed";

import {
	getMapCategoryDisplay,
	isArchitectureCategory,
	NodeDetailLevel,
} from "$features/system/lib/system-map/category";
import { layerOpacityAtDetail } from "$features/system/lib/system-map/interaction";
import type { GraphEntity, GraphSubset } from "$features/system/lib/system-map/graph";
import {
	boundsForNodes,
	nearestNodeIdAtScreenPoint,
	screenPointFromWorld,
	viewportCenteredOn,
	worldBoundsByNodeId,
	worldCenter,
	worldPositionByNodeId,
	type Point,
	type Size,
	type Viewport,
} from "$features/system/lib/system-map/geometry";
import {
	deriveNearbyEntityIds,
	EXPLICIT_FIT_MAX_ZOOM,
	INITIAL_VIEWPORT_ZOOM,
	initialFrameMaxZoom,
	MAP_MIN_ZOOM,
	minimumDetailForGraph,
	systemMapOverviewEmptyState,
} from "$features/system/lib/system-map/interaction";
import type {
	MapHighlights,
	MapProjection,
	MapSelection,
} from "$features/system/lib/system-map/presentation";
import { projectFullArchitecture, projectMap } from "$features/system/lib/system-map/projection";
import {
	selectedEntityIds,
	selectedRelationshipIds,
	selectionIsAvailable,
} from "$features/system/lib/system-map/selection";
import {
	alignLayoutToPrevious,
	nearestSurvivingArchitectureAnchor,
	presentMapProjection,
	systemMapLayoutInputKey,
	type MapRouteTransition,
	type LayoutPositionInputs,
} from "./layout";
import {
	CompactNodeDragSession,
	layoutRequestIsStale,
	nextLayoutRequestId,
	nodeCanBeDragged,
	nodesWithPendingDropPosition,
	type PendingNodeDrop,
} from "./drag-state";
import {
	connectionEndpointIds,
	connectionEndpointOpacity,
	connectionPresentationState,
	type ConnectionSelection,
} from "./map-connection/presentation";
import { connectionRouteWithNodeFeedback } from "./map-connection/geometry";
import type { FlowEdge, FlowNode, LayoutResult } from "./flow-model";
import { createSystemMapLayoutEngine, type SystemMapLayoutEngine } from "./layout-engine";
import { ProjectedRouteCache } from "./routing";
import { MapViewportState } from "./viewport-state";

export type SystemMapControllerOptions = {
	graph: () => GraphSubset;
	positions: () => Readonly<Partial<Record<string, Point>>> | undefined;
	selection: () => MapSelection | undefined;
	highlights: () => MapHighlights | undefined;
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
	onNodeMove: () => ((entityId: string, position: Point) => void) | undefined;
};

export type SystemMapStatus = "idle" | "loading" | "ready" | "error";

type LayoutSource = {
	graph: GraphSubset;
	fullProjection: MapProjection;
	positions: Readonly<Partial<Record<string, Point>>> | undefined;
	key: string;
};

type DisplayedCanvas = LayoutSource & { fullLayout: LayoutResult; routeCache: ProjectedRouteCache };

const emptyProjection: MapProjection = {
	nodes: [],
	connections: [],
	representativeByEntityId: new Map(),
};

const sameIds = (left: readonly string[], right: readonly string[]): boolean => {
	if (left.length !== right.length) return false;
	const rightIds = new Set(right);
	return left.every((id) => rightIds.has(id));
};

const nodeDataForEntity = (node: FlowNode, entity: GraphEntity): FlowNode["data"] => ({...node.data, entity});

type MoveEvent = Parameters<OnMove>[0];
type EdgePointerEvent = Parameters<NonNullable<EdgeEvents<FlowEdge>["onedgepointerenter"]>>[0];
type NodeDragEvent = Parameters<NonNullable<NodeEvents<FlowNode>["onnodedrag"]>>[0];
type ViewportSetter = (viewport: Viewport) => Promise<boolean>;

export class SystemMapController {
	private options = $state.raw<SystemMapControllerOptions>(null!);
	private disposed = true;
	private attached = false;
	private layoutEngine: SystemMapLayoutEngine | undefined;
	private requestId = 0;
	private pendingLayoutKey: string | undefined;
	private attemptedLayoutKey: string | undefined;
	private dragSession = new CompactNodeDragSession();
	private pendingDrop: PendingNodeDrop | undefined;
	private initialFrameApplied = false;
	private initialFrameRequested = false;
	private suppressNearbyRefresh = false;
	private viewportSetter: ViewportSetter | undefined;
	private interaction: MapViewportState;
	private source = $state.raw<LayoutSource>();
	private displayed = $state.raw<DisplayedCanvas>();
	private displayedProjection = $state.raw<MapProjection>(emptyProjection);
	private presentationLayout = $state.raw<LayoutResult>();
	private explicitRevealIds = new Set<string>();
	private derivedNearbyEntityIds: readonly string[] = [];
	private pendingRecenterEntityId: string | undefined;
	showAllConnectionLabels = $state(false);
	private hoveredConnectionId = $state<string>();

	nodes = $state.raw<FlowNode[]>([]);
	edges = $state.raw<FlowEdge[]>([]);
	viewport = $state<Viewport>({ x: 0, y: 0, zoom: INITIAL_VIEWPORT_ZOOM });
	width = $state(0);
	height = $state(0);
	status = $state<SystemMapStatus>("idle");
	error = $state<string>();
	continuousDetail = $state(0);
	structuralDetail = $state(0);

	loading = $derived(this.status === "loading");
	ready = $derived(this.status === "ready");
	nodesDraggable = $derived(Boolean(this.options.onNodeMove()));
	hasDiagram = $derived(Boolean(this.displayed?.fullLayout.nodes.length));
	showingOlderSuppliedData = $derived(
		Boolean(this.error && this.hasDiagram && this.displayed?.key !== this.source?.key)
	);
	canReveal = $derived(this.canRevealSelection());
	overviewEmptyState = $derived(
		!this.hasDiagram && this.status !== "loading" && this.source
			? systemMapOverviewEmptyState(this.source.graph, this.displayedProjection)
			: undefined
	);

	constructor(options: SystemMapControllerOptions) {
		this.options = options;
		this.source = this.createLayoutSource(options.graph(), options.positions());
		this.interaction = new MapViewportState(
			minimumDetailForGraph(options.graph()) ?? NodeDetailLevel.Landscape
		);
		this.syncDetailState();

		watch(
			() => [this.options.graph(), this.options.positions()] as const,
			() => this.acceptLayoutInput()
		);
		watch(
			() => this.options.onNodeMove(),
			() => this.updateSelectionPresentation()
		);
		watch(
			() => [this.options.selection(), this.options.highlights()] as const,
			() => this.updateSelectionPresentation()
		);
	}

	private createLayoutSource(
		graph: GraphSubset,
		positions: Readonly<Partial<Record<string, Point>>> | undefined
	): LayoutSource {
		const fullProjection = projectFullArchitecture(graph);
		const positionSnapshot = positions ? { ...positions } : undefined;
		return {
			graph,
			fullProjection,
			positions: positionSnapshot,
			key: systemMapLayoutInputKey(graph, fullProjection, positionSnapshot),
		};
	}

	mount = (viewportEl?: HTMLElement) => {
		this.disposed = false;
		this.attached = true;
		if (!viewportEl) return () => this.detach();

		const updateSize = () => {
			this.width = Math.max(0, viewportEl.clientWidth);
			this.height = Math.max(0, viewportEl.clientHeight);
			if (this.applyInitialFrameIfReady()) return;
			this.refreshNearby();
		};
		updateSize();
		const observer = new ResizeObserver(updateSize);
		observer.observe(viewportEl);
		this.acceptLayoutInput();

		return () => {
			observer.disconnect();
			this.detach();
		};
	};

	private detach() {
		this.attached = false;
		this.disposed = true;
		this.viewportSetter = undefined;
		this.hoveredConnectionId = undefined;
		this.requestId = nextLayoutRequestId(this.requestId);
		this.dragSession.cancel();
		this.pendingDrop = undefined;
		this.layoutEngine?.dispose();
		this.layoutEngine = undefined;
	}

	setViewportApi = (setViewport: ViewportSetter | undefined) => {
		this.viewportSetter = setViewport;
		if (setViewport) this.applyInitialFrameIfReady();
	};

	retry = () => {
		if (!this.source) return;
		this.attemptedLayoutKey = undefined;
		this.scheduleFullLayout(this.source, true);
	};

	toggleConnectionLabels = () => {
		this.showAllConnectionLabels = !this.showAllConnectionLabels;
		this.updateConnectionPresentation();
	};

	private syncDetailState() {
		this.continuousDetail = this.interaction.continuousDetail;
		this.structuralDetail = this.interaction.structuralDetail;
	}

	private canvasSize(): Size {
		return { width: this.width, height: this.height };
	}

	private currentSelection(): MapSelection | undefined {
		return this.options.selection();
	}

	private currentHighlights(): MapHighlights | undefined {
		return this.options.highlights();
	}

	private canRevealSelection(): boolean {
		const graph = this.displayed?.graph;
		const selection = this.currentSelection();
		return Boolean(
			graph &&
			selectionIsAvailable(graph, selection) &&
			selectedEntityIds(graph, selection).some((entityId) => {
				const entity = graph.entities.find((candidate) => candidate.id === entityId);
				return entity ? isArchitectureCategory(entity.category) : false;
			})
		);
	}

	private acceptLayoutInput() {
		if (this.disposed) return;
		const next = this.createLayoutSource(this.options.graph(), this.options.positions());
		this.source = next;
		if (next.key !== this.displayed?.key) {
			const suppliedIds = new Set(next.graph.entities.map((entity) => entity.id));
			this.explicitRevealIds = new Set(
				[...this.explicitRevealIds].filter((entityId) => suppliedIds.has(entityId))
			);
			this.interaction.setMinimumDetail(minimumDetailForGraph(next.graph) ?? NodeDetailLevel.Landscape);
			this.syncDetailState();
		}

		if (next.key === this.displayed?.key && this.displayed) {
			this.requestId = nextLayoutRequestId(this.requestId);
			this.pendingLayoutKey = undefined;
			this.attemptedLayoutKey = next.key;
			this.displayed = {
				...this.displayed,
				graph: next.graph,
				fullProjection: next.fullProjection,
				positions: next.positions,
			};
			this.status = "ready";
			this.error = undefined;
			this.updatePresentation();
			this.applyInitialFrameIfReady();
			return;
		}

		if (this.pendingLayoutKey === next.key || this.attemptedLayoutKey === next.key) return;
		if (this.dragSession.isActive) {
			this.requestId = nextLayoutRequestId(this.requestId);
			this.pendingLayoutKey = undefined;
			this.status = this.displayed ? "ready" : "loading";
			return;
		}
		if (this.attached) this.scheduleFullLayout(next);
	}

	private layoutPositionInputs(source: LayoutSource): LayoutPositionInputs {
		return {
			...(source.positions ? { positions: source.positions } : {}),
			...(this.displayed
				? {
						previousWorldPositions: Object.fromEntries(
							worldPositionByNodeId(this.displayed.fullLayout.nodes)
						),
					}
				: {}),
		};
	}

	private scheduleFullLayout(source: LayoutSource, force = false) {
		if (this.disposed || !this.attached) return;
		if (!force && (source.key === this.pendingLayoutKey || source.key === this.attemptedLayoutKey))
			return;

		this.attemptedLayoutKey = source.key;
		if (source.fullProjection.nodes.length === 0) {
			this.requestId = nextLayoutRequestId(this.requestId);
			this.pendingLayoutKey = undefined;
			this.displayed = {
				...source,
				fullLayout: { nodes: [], edges: [] },
				routeCache: new ProjectedRouteCache(),
			};
			this.status = "ready";
			this.error = undefined;
			this.updatePresentation();
			return;
		}

		const requestId = (this.requestId = nextLayoutRequestId(this.requestId));
		this.pendingLayoutKey = source.key;
		this.status = "loading";
		this.error = undefined;
		void this.layoutAndCommit(requestId, source, this.layoutPositionInputs(source)).catch(
			(error: unknown) => {
				if (layoutRequestIsStale(requestId, this.requestId, this.disposed)) return;
				this.pendingLayoutKey = undefined;
				this.status = this.displayed ? "ready" : "error";
				this.error = error instanceof Error ? error.message : "The system map could not be laid out.";
				console.error("System map layout failed", error);
			}
		);
	}

	private async layoutAndCommit(
		requestId: number,
		source: LayoutSource,
		positionInputs: LayoutPositionInputs
	) {
		const layoutEngine = (this.layoutEngine ??= createSystemMapLayoutEngine());
		const layout = await layoutEngine.layout(source.graph, source.fullProjection, positionInputs);
		if (layoutRequestIsStale(requestId, this.requestId, this.disposed)) return;
		if (this.dragSession.isActive) {
			this.pendingLayoutKey = undefined;
			this.status = this.displayed ? "ready" : "loading";
			return;
		}

		const latest = this.source;
		if (!latest || latest.key !== source.key) {
			this.pendingLayoutKey = undefined;
			this.acceptLayoutInput();
			return;
		}

		const alignedLayout = this.alignLayoutToPrevious(layout);
		const pendingHint = this.pendingDrop ? latest.positions?.[this.pendingDrop.entityId] : undefined;
		if (
			this.pendingDrop &&
			pendingHint?.x === this.pendingDrop.position.x &&
			pendingHint?.y === this.pendingDrop.position.y
		) {
			this.pendingDrop = undefined;
		}

		this.pendingLayoutKey = undefined;
		this.displayed = {
			...latest,
			fullLayout: alignedLayout,
			routeCache: new ProjectedRouteCache(),
		};
		this.status = "ready";
		this.error = undefined;
		this.updatePresentation();
		this.applyInitialFrameIfReady();
		this.refreshNearby();
	}

	private alignLayoutToPrevious(layout: LayoutResult): LayoutResult {
		const previous = this.displayed;
		if (!previous?.fullLayout.nodes.length || !layout.nodes.length) return layout;

		const nextIds = new Set(layout.nodes.map((node) => node.id));
		const selectedAnchor = selectedEntityIds(previous.graph, this.currentSelection()).find(
			(entityId) =>
				nextIds.has(entityId) && previous.fullLayout.nodes.some((node) => node.id === entityId)
		);
		const previousPresentation = this.presentationLayout?.nodes ?? previous.fullLayout.nodes;
		const screenCenter = { x: this.width / 2, y: this.height / 2 };
		const visibleCenterId = nearestNodeIdAtScreenPoint(previousPresentation, this.viewport, screenCenter);
		const directAnchor =
			selectedAnchor ?? (visibleCenterId && nextIds.has(visibleCenterId) ? visibleCenterId : undefined);
		const previousLayout = this.pendingDrop
			? {
					...previous.fullLayout,
					nodes: nodesWithPendingDropPosition(previous.fullLayout.nodes, this.pendingDrop),
				}
			: previous.fullLayout;
		if (directAnchor) return alignLayoutToPrevious(layout, previousLayout, directAnchor);

		const fallbackAnchor = nearestSurvivingArchitectureAnchor(
			layout,
			previousLayout,
			this.viewport,
			screenCenter
		);
		return fallbackAnchor
			? alignLayoutToPrevious(layout, previousLayout, fallbackAnchor.nextId, fallbackAnchor.previousId)
			: layout;
	}

	private projectionFor(graph: GraphSubset): MapProjection {
		return projectMap(graph, {
			detail: this.structuralDetail,
			visualDetail: this.continuousDetail,
			nearbyEntityIds: this.nearbyEntityIds(),
		});
	}

	private nearbyEntityIds(): string[] {
		return [...this.derivedNearbyEntityIds, ...this.explicitRevealIds];
	}

	private routeTransitionFor(graph: GraphSubset): MapRouteTransition {
		const nearbyEntityIds = this.nearbyEntityIds();
		const lowerLevel = Math.max(NodeDetailLevel.Landscape, Math.ceil(this.continuousDetail) - 1);
		const upperLevel = Math.min(NodeDetailLevel.Implementation, lowerLevel + 1);
		return {
			lowerProjection: projectMap(graph, {
				detail: lowerLevel,
				visualDetail: lowerLevel,
				nearbyEntityIds,
			}),
			upperProjection: projectMap(graph, {
				detail: upperLevel,
				visualDetail: upperLevel,
				nearbyEntityIds,
			}),
			upperLevel,
			progress: layerOpacityAtDetail(upperLevel, this.continuousDetail),
		};
	}

	private updatePresentation() {
		if (!this.displayed) {
			this.displayedProjection = emptyProjection;
			this.nodes = [];
			this.edges = [];
			return;
		}
		const projection = this.projectionFor(this.displayed.graph);
		this.displayedProjection = projection;
		this.presentationLayout = presentMapProjection(
			this.displayed.fullLayout,
			projection,
			this.displayed.routeCache,
			this.routeTransitionFor(this.displayed.graph)
		);
		this.updateSelectionPresentation();
	}

	private representativeIdsForEntityIds(graph: GraphSubset, ids: readonly string[]): ReadonlySet<string> {
		const suppliedIds = new Set(graph.entities.map((entity) => entity.id));
		const representatives = new Set<string>();
		for (const entityId of ids) {
			if (!suppliedIds.has(entityId)) continue;
			const representativeId = this.displayedProjection.representativeByEntityId.get(entityId);
			if (representativeId) representatives.add(representativeId);
		}
		return representatives;
	}

	private connectionPresentationOptions(): {
		selection?: ConnectionSelection;
		highlightedRelationshipIds: ReadonlySet<string>;
		highlightedEndpointIds: ReadonlySet<string>;
		hoveredConnectionId?: string;
		showAllConnectionLabels: boolean;
	} {
		const graph = this.displayed?.graph;
		const selection = this.currentSelection();
		const highlights = this.currentHighlights();
		const visibleRelationshipIds = new Set(
			(this.presentationLayout?.edges ?? []).flatMap(
				(edge) => edge.data?.connection.sourceRelationshipIds ?? []
			)
		);
		const selectedRelationshipIdSet = new Set(
			selectedRelationshipIds(selection).filter((relationshipId) =>
				visibleRelationshipIds.has(relationshipId)
			)
		);
		const selectedEndpoints = graph
			? this.representativeIdsForEntityIds(graph, selectedEntityIds(graph, selection))
			: new Set<string>();
		const highlightedEndpoints =
			graph && highlights
				? this.representativeIdsForEntityIds(graph, highlights.entityIds)
				: new Set<string>();
		const highlightedRelationshipIdSet = new Set(
			(highlights?.relationshipIds ?? []).filter((relationshipId) =>
				visibleRelationshipIds.has(relationshipId)
			)
		);
		const selectionIsCurrent = Boolean(graph && selection && selectionIsAvailable(graph, selection));

		return {
			...(selectionIsCurrent
				? {
						selection: {
							relationshipIds: selectedRelationshipIdSet,
							endpointIds: selectedEndpoints,
						},
					}
				: {}),
			highlightedRelationshipIds: highlightedRelationshipIdSet,
			highlightedEndpointIds: highlightedEndpoints,
			hoveredConnectionId: this.hoveredConnectionId,
			showAllConnectionLabels: this.showAllConnectionLabels,
		};
	}

	private decorateNodes(layoutNodes: readonly FlowNode[], layoutEdges: readonly FlowEdge[]): FlowNode[] {
		const graph = this.displayed?.graph;
		if (!graph) return [...layoutNodes];
		const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
		const selection = this.currentSelection();
		const highlights = this.currentHighlights();
		const selectedEntityIdsForSelection = selectedEntityIds(graph, selection);
		const selectedRepresentatives = this.representativeIdsForEntityIds(
			graph,
			selectedEntityIdsForSelection
		);
		const selectedNodeRepresentatives =
			selection?.kind === "entity" ? selectedRepresentatives : new Set<string>();
		const attachedRepresentatives =
			selection?.kind === "entity" ? new Set<string>() : selectedRepresentatives;
		const highlightedRepresentatives = highlights
			? this.representativeIdsForEntityIds(graph, highlights.entityIds)
			: new Set<string>();
		const endpointIds = connectionEndpointIds(layoutEdges, this.connectionPresentationOptions());
		const hoveredEndpointIds = new Set(
			layoutEdges
				.filter((edge) => edge.id === this.hoveredConnectionId)
				.flatMap((edge) => [edge.source, edge.target])
		);

		return nodesWithPendingDropPosition(layoutNodes, this.pendingDrop).map((node) => {
			const entity = entitiesById.get(node.id);
			if (!entity) return node;
			const opacity = connectionEndpointOpacity(
				node.data.opacity ?? 1,
				hoveredEndpointIds.has(node.id)
			);
			const isInteractive = opacity > 0.06;
			const draggable =
				isInteractive && nodeCanBeDragged(node.data.appearance, Boolean(this.options.onNodeMove()));
			return {
				...node,
				style: `opacity: ${opacity};${isInteractive ? "" : "pointer-events: none"}`,
				focusable: isInteractive,
				selectable: isInteractive,
				domAttributes: isInteractive ? undefined : { "aria-hidden": true },
				draggable,
				selected: selectedNodeRepresentatives.has(node.id),
				data: {
					...nodeDataForEntity(node, entity),
					opacity,
					isInteractive,
					isDraggable: draggable,
					isHighlighted:
						attachedRepresentatives.has(node.id) || highlightedRepresentatives.has(node.id),
					isConnectionEndpoint: endpointIds.has(node.id),
				},
			};
		});
	}

	private decorateEdges(layoutEdges: readonly FlowEdge[]): FlowEdge[] {
		const options = this.connectionPresentationOptions();
		const worldPositions = worldPositionByNodeId(this.nodes);
		const originalWorldPositions = worldPositionByNodeId(this.presentationLayout?.nodes ?? []);
		const compactIds = new Set(
			(this.presentationLayout?.nodes ?? [])
				.filter((node) => node.data.appearance === "compact")
				.map((node) => node.id)
		);

		const movedCompactPositions = new Map<string, Point>();
		for (const entityId of compactIds) {
			const currPos = worldPositions.get(entityId);
			const ogPos = originalWorldPositions.get(entityId);
			if (!currPos || !ogPos || (currPos.x === ogPos.x && currPos.y === ogPos.y)) continue;
			movedCompactPositions.set(entityId, { 
				x: currPos.x - ogPos.x,
				y: currPos.y - ogPos.y,
			});
		}

		return layoutEdges.map((edge) => {
			const state = connectionPresentationState(edge, options);
			const data = edge.data;
			const opacity = data?.opacity ?? 1;
			const isInteractive = opacity > 0.06;

			const sourcePos = movedCompactPositions.get(edge.source);
			const targetPos = movedCompactPositions.get(edge.target);
			const route = data?.route
				? connectionRouteWithNodeFeedback(data.route, sourcePos, targetPos)
				: undefined;
			const routeVariants = data?.routeVariants?.map((variant) => ({
				...variant,
				route: connectionRouteWithNodeFeedback(variant.route, sourcePos, targetPos),
			}));

			const newEdgeData = {
				...(route ? { route } : {}),
				...(routeVariants ? { routeVariants } : {}),
				opacity,
				isHighlighted: state.highlighted,
				isHovered: state.hovered,
				isDimmed: state.dimmed,
				isLabelVisible: state.labelVisible,
			}
			
			return {
				...edge,
				focusable: isInteractive,
				class: isInteractive ? edge.class : "inactive",
				interactionWidth: isInteractive ? 24 : 0,
				domAttributes: isInteractive ? undefined : { "aria-hidden": true },
				selected: state.selected,
				zIndex: state.highlighted ? 1 : 0,
				data: data ? { ...data, ...newEdgeData } : data,
			};
		});
	}

	private updateConnectionPresentation() {
		if (!this.presentationLayout) return;
		this.edges = this.decorateEdges(this.presentationLayout.edges);
		this.nodes = this.decorateNodes(this.presentationLayout.nodes, this.presentationLayout.edges);
	}

	private updateDragRouteFeedback() {
		if (!this.presentationLayout) return;
		this.edges = this.decorateEdges(this.presentationLayout.edges);
	}

	private updateSelectionPresentation() {
		if (!this.presentationLayout) return;
		this.nodes = this.decorateNodes(this.presentationLayout.nodes, this.presentationLayout.edges);
		this.edges = this.decorateEdges(this.presentationLayout.edges);
	}

	private updateDerivedNearby(): boolean {
		const displayed = this.displayed;
		if (!displayed) return false;
		const next = deriveNearbyEntityIds(
			displayed.graph,
			displayed.fullProjection,
			displayed.fullLayout.nodes,
			this.viewport,
			this.canvasSize()
		);
		if (sameIds(this.derivedNearbyEntityIds, next)) return false;
		this.derivedNearbyEntityIds = next;
		this.updatePresentation();
		return true;
	}

	private refreshNearby() {
		if (this.suppressNearbyRefresh || !this.displayed?.fullLayout.nodes.length) return;
		this.updateDerivedNearby();
	}

	private processViewport(viewport: Viewport) {
		const previousContinuousDetail = this.interaction.continuousDetail;
		const structuralChanged = this.interaction.updateViewport(viewport);
		this.viewport = { ...viewport };
		this.syncDetailState();
		if (structuralChanged || previousContinuousDetail !== this.continuousDetail)
			this.updatePresentation();
		else this.updateSelectionPresentation();
		this.refreshNearby();

		if (this.pendingRecenterEntityId && this.centerEntity(this.pendingRecenterEntityId)) {
			this.pendingRecenterEntityId = undefined;
		}
	}

	private setViewport(viewport: Viewport, onApplied?: () => void, onFailed?: () => void): boolean {
		if (!this.viewportSetter) return false;
		void this.viewportSetter(viewport)
			.then((applied) => {
				if (!applied) {
					onFailed?.();
					return;
				}
				this.processViewport(viewport);
				onApplied?.();
			})
			.catch((error: unknown) => {
				onFailed?.();
				console.error("System map viewport update failed", error);
			});
		return true;
	}

	private fitToMaxZoom(maxZoom: number) {
		if (!this.nodes.length || this.width <= 0 || this.height <= 0) return;
		const bounds = boundsForNodes(this.nodes);
		if (!bounds) return;
		this.setViewport(getViewportForBounds(bounds, this.width, this.height, MAP_MIN_ZOOM, maxZoom, 0.3));
	}

	fit = () => this.fitToMaxZoom(EXPLICIT_FIT_MAX_ZOOM);

	private fitInitialFrame(): Viewport | undefined {
		if (!this.nodes.length || this.width <= 0 || this.height <= 0) return undefined;
		const bounds = boundsForNodes(this.nodes);
		if (!bounds) return undefined;
		return getViewportForBounds(
			bounds,
			this.width,
			this.height,
			MAP_MIN_ZOOM,
			initialFrameMaxZoom(this.structuralDetail),
			0.3
		);
	}

	private applyInitialFrameIfReady(): boolean {
		if (this.initialFrameApplied || this.initialFrameRequested || !this.nodes.length) return false;
		const viewport = this.fitInitialFrame();
		if (!viewport) return false;
		this.suppressNearbyRefresh = true;
		this.initialFrameRequested = this.setViewport(
			viewport,
			() => {
				this.initialFrameApplied = true;
				this.initialFrameRequested = false;
				this.suppressNearbyRefresh = false;
			},
			() => {
				this.initialFrameRequested = false;
				this.suppressNearbyRefresh = false;
			}
		);
		if (!this.initialFrameRequested) this.suppressNearbyRefresh = false;
		return this.initialFrameRequested;
	}

	private centerEntity(entityId: string): boolean {
		if (!this.nodes.length || this.width <= 0 || this.height <= 0) return false;
		const representativeId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
		const bounds = worldBoundsByNodeId(this.nodes).get(representativeId);
		if (!bounds) return false;
		return this.setViewport(viewportCenteredOn(this.viewport, worldCenter(bounds), this.canvasSize()));
	}

	private screenPointForEntity(entityId: string): Point | undefined {
		const representativeId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
		const bounds = worldBoundsByNodeId(this.nodes).get(representativeId);
		return bounds ? screenPointFromWorld(this.viewport, worldCenter(bounds)) : undefined;
	}

	recenter = () => {
		const graph = this.displayed?.graph;
		if (graph) {
			for (const entityId of selectedEntityIds(graph, this.currentSelection())) {
				if (this.centerEntity(entityId)) return;
			}
		}
		if (!this.nodes.length || this.width <= 0 || this.height <= 0) return;
		const bounds = boundsForNodes(this.nodes);
		if (bounds)
			this.setViewport(viewportCenteredOn(this.viewport, worldCenter(bounds), this.canvasSize()));
	};

	zoomBy = (factor: number) => {
		const graph = this.displayed?.graph;
		const selectedId = graph ? selectedEntityIds(graph, this.currentSelection())[0] : undefined;
		const representativeId = selectedId
			? (this.displayedProjection.representativeByEntityId.get(selectedId) ?? selectedId)
			: undefined;
		const focusBounds = representativeId
			? worldBoundsByNodeId(this.nodes).get(representativeId)
			: undefined;
		const nextViewport = this.interaction.prepareKeyboardZoom(factor, focusBounds, this.canvasSize());
		this.setViewport(nextViewport);
	};

	reveal = (target: MapSelection) => {
		const graph = this.displayed?.graph;
		if (!this.displayed || !graph || !selectionIsAvailable(graph, target)) return;
		const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
		const entityIds = selectedEntityIds(graph, target).filter((entityId) => {
			const entity = entitiesById.get(entityId);
			return entity ? isArchitectureCategory(entity.category) : false;
		});
		if (!entityIds.length) return;

		let nextDetail = this.structuralDetail;
		for (const entityId of entityIds) {
			this.explicitRevealIds.add(entityId);
			const level = getMapCategoryDisplay(entitiesById.get(entityId)!.category).level;
			if (level !== undefined) nextDetail = Math.max(nextDetail, level);
		}

		const focusEntityId = entityIds[0];
		this.pendingRecenterEntityId = focusEntityId;
		const screenPoint = this.screenPointForEntity(focusEntityId) ?? {
			x: this.width > 0 ? this.width / 2 : 0,
			y: this.height > 0 ? this.height / 2 : 0,
		};
		const transition = this.interaction.prepareReveal(nextDetail, screenPoint);
		this.syncDetailState();
		if (transition.zoomChanged) {
			if (!this.setViewport(transition.viewport)) this.pendingRecenterEntityId = undefined;
			return;
		}

		this.updatePresentation();
		if (this.centerEntity(focusEntityId)) this.pendingRecenterEntityId = undefined;
	};

	revealSelected = () => {
		const selection = this.currentSelection();
		if (selection) this.reveal(selection);
	};

	selectNode = (nodeId: string, event?: Event) => {
		if (!this.displayed?.graph.entities.some((entity) => entity.id === nodeId)) return;
		this.options.onSelectionChange({ kind: "entity", entityId: nodeId }, this.selectionTrigger(event));
	};

	selectEdge = (edgeId: string, event?: Event) => {
		const connection = this.presentationLayout?.edges.find((edge) => edge.id === edgeId)?.data
			?.connection;
		if (!connection || connection.sourceRelationshipIds.length === 0) return;
		const selection: MapSelection =
			connection.classification === "direct"
				? { kind: "relationship", relationshipId: connection.sourceRelationshipIds[0] }
				: { kind: "summary", relationshipIds: [...connection.sourceRelationshipIds] };
		this.options.onSelectionChange(selection, this.selectionTrigger(event));
	};

	clearSelection = (event?: Event) => {
		this.options.onSelectionChange(undefined, this.selectionTrigger(event));
	};

	onViewportChange = (viewport: Viewport) => {
		this.processViewport(viewport);
	};

	private nodesFromDragEvent(event: NodeDragEvent): FlowNode[] {
		const nodesById = new Map(this.nodes.map((node) => [node.id, node]));
		for (const node of event.nodes) nodesById.set(node.id, node);
		if (event.targetNode) nodesById.set(event.targetNode.id, event.targetNode);
		return [...nodesById.values()];
	}

	onNodeDragStart = (event: NodeDragEvent) => {
		const node = event.targetNode;
		if (!node) return;
		const nodes = this.nodesFromDragEvent(event);
		const position = worldPositionByNodeId(nodes).get(node.id);
		if (!position) return;
		if (
			!this.dragSession.start(
				node.id,
				node.data.appearance,
				position,
				Boolean(this.options.onNodeMove())
			)
		)
			return;

		this.nodes = nodes;
		this.error = undefined;
	};

	onNodeDrag = (event: NodeDragEvent) => {
		const node = event.targetNode;
		if (!node) return;
		const nodes = this.nodesFromDragEvent(event);
		const position = worldPositionByNodeId(nodes).get(node.id);
		if (!position || !this.dragSession.update(node.id, position)) return;
		this.nodes = nodes;
		this.updateDragRouteFeedback();
	};

	onNodeDragStop = (event: NodeDragEvent) => {
		const node = event.targetNode;
		if (!node) {
			this.dragSession.cancel();
			return;
		}
		const nodes = this.nodesFromDragEvent(event);
		const position = worldPositionByNodeId(nodes).get(node.id);
		if (!position) {
			this.dragSession.cancel();
			return;
		}
		const delta = this.dragSession.update(node.id, position);
		this.nodes = nodes;
		const moved = Boolean(delta && (delta.x !== 0 || delta.y !== 0));
		if (moved) this.pendingDrop = { entityId: node.id, position };
		this.dragSession.finish(node.id, position, this.options.onNodeMove());
		this.updateDragRouteFeedback();
		if (moved) this.acceptLayoutInput();
	};

	onEdgePointerEnter = (event: EdgePointerEvent) => {
		if (!this.presentationLayout?.edges.some(({ id }) => id === event.edge.id)) return;
		if (this.hoveredConnectionId === event.edge.id) return;
		this.hoveredConnectionId = event.edge.id;
		this.updateSelectionPresentation();
	};

	onEdgePointerLeave = (event: EdgePointerEvent) => {
		if (this.hoveredConnectionId !== event.edge.id) return;
		this.hoveredConnectionId = undefined;
		this.updateSelectionPresentation();
	};

	keydown = (event: KeyboardEvent) => {
		if (event.key === "Escape") {
			event.preventDefault();
			this.clearSelection(event);
			return;
		}

		const target = event.target;
		if (
			target instanceof Element &&
			target.closest("input, textarea, select, button, [contenteditable=true]")
		)
			return;

		if (["+", "="].includes(event.key)) {
			event.preventDefault();
			this.zoomBy(1.2);
			return;
		}
		if (["-", "_"].includes(event.key)) {
			event.preventDefault();
			this.zoomBy(1 / 1.2);
			return;
		}
		if (event.key.toLowerCase() === "f") {
			event.preventDefault();
			this.fit();
			return;
		}
		if (event.key.toLowerCase() === "r") {
			event.preventDefault();
			this.recenter();
			return;
		}
		if (event.key.toLowerCase() === "v") {
			event.preventDefault();
			this.revealSelected();
			return;
		}

		if (!(event.key === "Enter" || event.key === " ") || !(target instanceof Element)) return;
		const flowNode = target.closest<HTMLElement>(".svelte-flow__node");
		const flowEdge = target.closest<HTMLElement>(".svelte-flow__edge");
		const id = flowNode?.getAttribute("data-id") ?? flowEdge?.getAttribute("data-id");
		if (!id) return;
		event.preventDefault();
		event.stopPropagation();
		if (flowNode) this.selectNode(id, event);
		if (flowEdge) this.selectEdge(id, event);
	};

	private selectionTrigger(event?: Event): HTMLElement | undefined {
		if (!(event?.target instanceof Element)) return undefined;
		return (
			event.target.closest<HTMLElement>(".svelte-flow__node, .svelte-flow__edge") ??
			(event.target instanceof HTMLElement ? event.target : undefined)
		);
	}
}
