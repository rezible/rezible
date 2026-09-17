import { getViewportForBounds, type OnMove } from "@xyflow/svelte";
import { Context, watch } from "runed";

import {
	DisplayMode,
	NodeDetailLevel,
	getMapCategoryDisplay,
} from "$features/systems/lib/system-map/category";
import type { GraphEntity, GraphSubset } from "$features/systems/lib/system-map/graph";
import {
	boundsForNodes,
	nearestNodeIdAtScreenPoint,
	screenPointFromWorld,
	viewportCenteredOn,
	worldBoundsByNodeId,
	worldCenter,
	type Point,
	type Size,
	type Viewport,
} from "$features/systems/lib/system-map/geometry";
import {
	MAP_MIN_ZOOM,
	EXPLICIT_FIT_MAX_ZOOM,
	INITIAL_VIEWPORT_ZOOM,
	minimumDetailForGraph,
	initialFrameMaxZoom,
	deriveNearbyEntityIds,
	systemMapOverviewEmptyState,
	type SystemMapOverviewEmptyState,
} from "$features/systems/lib/system-map/interaction";
import {
	buildInspectionItems,
	inspectionItemId,
	inspectionTargetAvailable,
	resolveInspectionTarget,
	type SystemMapInspection,
	type SystemMapInspectionItem,
} from "$features/systems/lib/system-map/inspection";
import { projectMap } from "$features/systems/lib/system-map/projection";
import type {
	InspectionTarget,
	MapDisplayOptions,
	MapProjection,
} from "$features/systems/lib/system-map/presentation";
import { preferredLayoutAnchorId, settleLayoutChain, MapViewportState } from "./viewport-state";
import { alignLayoutToPrevious, nearestSurvivingArchitectureAnchor } from "./layout";
import {
	connectionPresentationState,
	type ConnectionSelection,
} from "./map-connection/presentation";
import type { FlowEdge, FlowNode, FlowNodeData, LayoutResult } from "./flow-model";
import { createSystemMapLayoutEngine, type SystemMapLayoutEngine } from "./layout-engine";
import { onMount } from "svelte";

export type SystemMapSelection = {
	target: InspectionTarget;
	trigger?: HTMLElement;
};

export type SystemMapControllerOptions = {
	graph: () => GraphSubset;
	displayOptions: () => MapDisplayOptions;
	onSelect?: (selection: SystemMapSelection) => void;
	onClearSelection?: () => void;
	onViewportChange?: (viewport: Viewport) => void;
};

export type SystemMapStatus = "idle" | "loading" | "ready" | "error";

const emptyProjection: MapProjection = {
	nodes: [],
	connections: [],
	annotations: [],
	representativeByEntityId: new Map(),
};

const sameIds = (left: readonly string[], right: readonly string[]): boolean => {
	if (left.length !== right.length) return false;
	const rightIds = new Set(right);
	return left.every((id) => rightIds.has(id));
};

const entityIdsForTarget = (
	target: InspectionTarget | undefined,
	graph: GraphSubset | undefined
): string[] => {
	if (!target) return [];

	switch (target.kind) {
		case "entity":
			return [target.entityId];
		case "annotation": {
			const relationship = graph?.relationships.find(
				(candidate) => candidate.id === target.relationshipId
			);
			return relationship ? [relationship.target, relationship.source] : [target.entityId];
		}
		case "relationship": {
			const relationship = graph?.relationships.find(
				(candidate) => candidate.id === target.relationshipId
			);
			return relationship ? [relationship.source, relationship.target] : [];
		}
		case "summary": {
			const relationshipIds = new Set(target.sourceRelationshipIds);
			const ids: string[] = [];
			for (const relationship of graph?.relationships ?? []) {
				if (!relationshipIds.has(relationship.id)) continue;
				ids.push(relationship.source, relationship.target);
			}
			return [...new Set(ids)];
		}
	}
};

const nodeDataForEntity = (node: FlowNode, entity: GraphEntity): FlowNodeData => ({
	...node.data,
	entity,
});

type MoveEvent = Parameters<OnMove>[0];

type RebuildInputs = {
	graph: GraphSubset;
	displayOptions: MapDisplayOptions;
	nearbyEntityIds: readonly string[];
};

const screenPointFromMoveEvent = (event: MoveEvent): Point | undefined => {
	if (!event) return undefined;

	const point = "touches" in event ? (event.touches[0] ?? event.changedTouches[0]) : event;
	if (!point || !Number.isFinite(point.clientX) || !Number.isFinite(point.clientY)) return undefined;

	if (typeof Element !== "undefined" && event.target instanceof Element) {
		const flowElement = event.target.closest<HTMLElement>(".svelte-flow");
		const bounds = flowElement?.getBoundingClientRect();
		if (bounds) return { x: point.clientX - bounds.left, y: point.clientY - bounds.top };
	}

	return { x: point.clientX, y: point.clientY };
};

export class SystemMapController {
	// Inputs and lifecycle. Supplied source is current even while its replacement is laying out.
	private options = $state.raw<SystemMapControllerOptions>(null!);
	private disposed = $state(false);
	private attached = $state(false);
	private layoutEngine: SystemMapLayoutEngine | undefined;
	private requestId = 0;
	private initialFrameApplied = false;
	private suppressNearbyRefresh = false;
	private interaction!: MapViewportState;

	// Supplied source and inspection. Inspection follows these latest inputs, not the last canvas commit.
	private suppliedGraph = $state.raw<GraphSubset>();
	private suppliedProjection = $state.raw<MapProjection>(emptyProjection);

	// Increments for every explicit selection, including reselecting the same target.
	selectionRevision = $state(0);
	selectedTarget = $state.raw<InspectionTarget>();
	selectedItemId = $derived(inspectionItemId(this.selectedTarget));
	selectedEntityIds = $derived(entityIdsForTarget(this.selectedTarget, this.suppliedGraph));
	private selectedEdgeId = $state<string>();

	inspection = $state.raw<SystemMapInspection>();
	inspectionItems = $state.raw<SystemMapInspectionItem[]>([]);

	// Displayed diagram. These values change together only after a layout has succeeded; a failed
	// replacement leaves this snapshot visible while the supplied source and inspection stay current.
	private displayedGraph = $state.raw<GraphSubset>();
	displayedProjection = $state.raw<MapProjection>(emptyProjection);
	private displayedLayout = $state.raw<LayoutResult>();

	// Viewport and reveal state.
	private explicitRevealIds = new Set<string>();
	private derivedNearbyEntityIds: readonly string[] = [];
	private pendingRecenterEntityId: string | undefined;

	nodes = $state.raw<FlowNode[]>([]);
	edges = $state.raw<FlowEdge[]>([]);

	/** Svelte Flow-bound viewport; compare incoming moves against interaction.lastProcessedViewport. */
	viewport = $state<Viewport>({ x: 0, y: 0, zoom: INITIAL_VIEWPORT_ZOOM });
	width = $state(0);
	height = $state(0);
	status = $state<SystemMapStatus>("idle");
	error = $state<string>();
	canReveal = $derived(Boolean(this.inspection && this.inspection.visibility !== "visible"));
	/**
	 * Reactive mirrors of MapViewportState's detail values. They synchronize after each processed
	 * viewport update; keep this explicit boundary until a reactivity redesign is justified.
	 */
	continuousDetail = $state(0);
	structuralDetail = $state(0);

	loading = $derived(this.status === "loading");
	ready = $derived(this.status === "ready");
	overviewEmptyState = $derived(this.ready
		? systemMapOverviewEmptyState(this.displayedGraph, this.displayedProjection)
		: undefined
	);
	hasDiagram = $derived(this.displayedLayout !== undefined);

	constructor(options: SystemMapControllerOptions) {
		this.options = options;
		this.interaction = new MapViewportState(
			minimumDetailForGraph(options.graph()) ?? NodeDetailLevel.Landscape
		);
		this.syncDetailState();

		watch(
			() => [this.options.graph(), this.options.displayOptions()] as const,
			() => this.scheduleRebuild()
		);
	}

	mount = (viewportEl?: HTMLElement) => {
		this.attach();
		if (!viewportEl) {
			this.detach();
			return;
		}

		const updateSize = () => {
			this.width = Math.max(0, viewportEl.clientWidth);
			this.height = Math.max(0, viewportEl.clientHeight);
			if (this.applyInitialFrameIfReady()) return;
			this.refreshNearby();
		}
		updateSize();
		
		const observer = new ResizeObserver(updateSize);
		observer.observe(viewportEl);

		return () => {
			observer.disconnect();
			this.detach();
		};
	}

	private attach = () => {
		const shouldRefresh = this.disposed || !this.attached;
		this.disposed = false;
		this.attached = true;
		if (shouldRefresh) this.scheduleRebuild();
	};

	private detach = () => {
		this.attached = false;
		this.disposed = true;
		this.requestId += 1;
		this.layoutEngine?.dispose();
		this.layoutEngine = undefined;
	};

	retry = () => {
		this.scheduleRebuild();
	};

	/** Synchronizes the reactive controller mirrors at the viewport-state boundary. */
	private syncDetailState() {
		this.continuousDetail = this.interaction.continuousDetail;
		this.structuralDetail = this.interaction.structuralDetail;
	}

	private canvasSize(): Size {
		return { width: this.width, height: this.height };
	}

	// Supplied source inspection: publish source facts before layout, so inspection stays current
	// even when the displayed diagram is still the last successful snapshot.
	private publishInspection(graph: GraphSubset, projection: MapProjection) {
		this.suppliedGraph = graph;
		this.suppliedProjection = projection;
		this.inspectionItems = buildInspectionItems(graph, projection);

		if (this.selectedTarget && !inspectionTargetAvailable(graph, this.selectedTarget)) {
			this.selectedTarget = undefined;
			this.selectedEdgeId = undefined;
			this.inspection = undefined;
			this.options.onClearSelection?.();
			this.updateSelectionPresentation();
			return;
		}

		const hasSelectedEdge = projection.connections.some(({id}) => id === this.selectedEdgeId)
		if (this.selectedEdgeId && !hasSelectedEdge) {
			this.selectedEdgeId = undefined;
		}
		this.refreshInspection();
	}

	private resolveInspection(): SystemMapInspection | undefined {
		if (!this.suppliedGraph || !this.selectedTarget) return;
		if (!inspectionTargetAvailable(this.suppliedGraph, this.selectedTarget)) return;
		return resolveInspectionTarget(this.suppliedGraph, this.suppliedProjection, this.selectedTarget);
	}

	private refreshInspection() {
		this.inspection = this.resolveInspection();
	}

	// Layout scheduling and completion.
	private readInputs(): RebuildInputs {
		const graph = this.options.graph();
		const minimumDetail = minimumDetailForGraph(graph);
		this.interaction.setMinimumDetail(minimumDetail ?? NodeDetailLevel.Landscape);
		
		this.syncDetailState();

		const baselineIds = graph.entities
			.filter((entity) =>
				minimumDetail !== undefined && getMapCategoryDisplay(entity.category).level === minimumDetail
			)
			.map((entity) => entity.id);

		const sourceChanged = this.suppliedGraph !== undefined && this.suppliedGraph !== graph;
		if (sourceChanged) {
			this.derivedNearbyEntityIds = [];
			this.explicitRevealIds.clear();
			this.pendingRecenterEntityId = undefined;
		}

		const nearbyEntityIds = new Set([
			...baselineIds,
			...this.derivedNearbyEntityIds,
			...this.explicitRevealIds,
		])

		return {
			graph,
			displayOptions: this.options.displayOptions(),
			nearbyEntityIds: [...nearbyEntityIds],
		};
	}

	private project(inputs: RebuildInputs): MapProjection {
		const revealState = { 
			detail: this.structuralDetail, 
			nearbyEntityIds: inputs.nearbyEntityIds,
		};
		return projectMap(inputs.graph, revealState, inputs.displayOptions);
	}

	private scheduleRebuild() {
		if (this.disposed) return;
		
		const inputs = this.readInputs();
		const projection = this.project(inputs);
		this.publishInspection(inputs.graph, projection);
		this.requestLayout(inputs.graph, projection);
	}

	private requestLayout(graph: GraphSubset, projection: MapProjection) {
		const requestId = ++this.requestId;

		this.status = "loading";
		this.error = undefined;

		const onLayoutError = (err: unknown) => {
			if (this.isStaleLayoutRequest(requestId)) return;
			this.status = this.displayedLayout ? "ready" : "error";
			this.error = err instanceof Error ? err.message : "The system map could not be laid out.";
			console.error("System map layout failed", err);
		}

		void this.layoutAndCommit(requestId, graph, projection).catch(onLayoutError);
	}

	private isStaleLayoutRequest(requestId: number) {
		return this.disposed || requestId !== this.requestId;
	}

	private async layoutAndCommit(requestId: number, graph: GraphSubset, projection: MapProjection) {
		const layoutEngine = (this.layoutEngine ??= createSystemMapLayoutEngine());
		const layout = await layoutEngine.layout(graph, projection);
		if (this.isStaleLayoutRequest(requestId)) return;

		const stableLayout = this.alignAndCommit(graph, projection, layout);
		this.status = "ready";
		this.error = undefined;
		this.settleInitialFrameAndFollowup(requestId, graph, projection, stableLayout);
	}

	// Layout completion keeps a meaningful source anchor when representations change.
	private anchorForProjection(projection: MapProjection, layout: LayoutResult) {
		if (!this.displayedLayout) return undefined;
		const candidateIds: string[] = [];
		const preferredAnchorId = preferredLayoutAnchorId(
			this.interaction.pendingLayoutAnchorId,
			this.selectedEntityIds
		);
		if (preferredAnchorId) candidateIds.push(preferredAnchorId);

		candidateIds.push(...this.selectedEntityIds.filter((id) => id !== preferredAnchorId));

		const canvasCenter: Point = {
			x: this.width > 0 ? this.width / 2 : 0,
			y: this.height > 0 ? this.height / 2 : 0,
		};
		const centerNodeId = nearestNodeIdAtScreenPoint(
			this.displayedLayout.nodes,
			this.viewport,
			canvasCenter
		);
		if (centerNodeId) candidateIds.push(centerNodeId);

		
		const previousBounds = worldBoundsByNodeId(this.displayedLayout.nodes);
		let referenceScreenPoint: Point = canvasCenter;
		for (const entityId of candidateIds) {
			const previousId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
			const previousBoundsForCandidate = previousBounds.get(previousId);
			if (previousBoundsForCandidate) {
				const center = worldCenter(previousBoundsForCandidate);
				referenceScreenPoint = screenPointFromWorld(this.viewport, center);
				break;
			}
		}

		const seenCandidates = new Set<string>();
		for (const entityId of candidateIds) {
			if (seenCandidates.has(entityId)) continue;
			seenCandidates.add(entityId);
			const nextId = projection.representativeByEntityId.get(entityId) ?? entityId;
			const nextInLayout = layout.nodes.some(({id}) => id === nextId);

			const previousId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
			const previousInDisplay = this.displayedLayout.nodes.some(({id}) => id === previousId);

			if (nextInLayout && previousInDisplay) {
				return { nextId, previousId };
			}
		}

		return nearestSurvivingArchitectureAnchor(
			layout,
			this.displayedLayout,
			this.viewport,
			referenceScreenPoint
		);
	}

	private alignAndCommit(
		graph: GraphSubset,
		projection: MapProjection,
		layout: LayoutResult
	): LayoutResult {
		const anchor = this.anchorForProjection(projection, layout);
		const stableLayout = alignLayoutToPrevious(
			layout,
			this.displayedLayout,
			anchor?.nextId,
			anchor?.previousId
		);
		this.displayedGraph = graph;
		this.displayedLayout = stableLayout;
		this.displayedProjection = projection;
		this.nodes = this.decorateNodes(stableLayout.nodes, graph);
		this.edges = this.decorateEdges(stableLayout.edges);
		this.refreshInspection();
		return stableLayout;
	}

	/** Completes a successful layout with the initial frame, pending recenter, and nearby feedback. */
	private settleInitialFrameAndFollowup(
		requestId: number,
		graph: GraphSubset,
		projection: MapProjection,
		stableLayout: LayoutResult
	) {
		const initialFrameApplied = this.applyInitialFrameIfReady();

		if (this.pendingRecenterEntityId) {
			const wasSuppressingNearbyRefresh = this.suppressNearbyRefresh;
			this.suppressNearbyRefresh = initialFrameApplied || wasSuppressingNearbyRefresh;
			try {
				if (this.centerEntity(this.pendingRecenterEntityId)) this.pendingRecenterEntityId = undefined;
			} finally {
				this.suppressNearbyRefresh = wasSuppressingNearbyRefresh;
			}
		}

		if (initialFrameApplied) {
			if (!this.isStaleLayoutRequest(requestId)) this.interaction.clearLayoutAnchor();
			return;
		}

		settleLayoutChain({
			isCurrent: () => !this.isStaleLayoutRequest(requestId),
			updateNearby: () => this.updateDerivedNearby(graph, projection, stableLayout.nodes),
			clearAnchor: () => this.interaction.clearLayoutAnchor(),
		});
	}

	// Displayed diagram presentation: decorate the last successful layout with current selection.
	private decorateNodes(
		layoutNodes: readonly FlowNode[],
		graph: GraphSubset
	): FlowNode[] {
		const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
		const selectedEntityIds = this.selectedEntityRepresentatives();

		return layoutNodes.map((node) => {
			const entity = entitiesById.get(node.id);
			if (!entity) return node;
			return {
				...node,
				selected: selectedEntityIds.has(node.id),
				data: nodeDataForEntity(node, entity),
			};
		});
	}

	private decorateEdges(layoutEdges: readonly FlowEdge[]): FlowEdge[] {
		const selection: ConnectionSelection | undefined = this.selectedTarget
			? { edgeId: this.selectedEdgeId, endpointIds: this.selectedEntityRepresentatives() }
			: undefined;

		return layoutEdges.map((edge) => {
			const { 
				selected, 
				highlighted: isHighlighted, 
				dimmed: isDimmed,
			} = connectionPresentationState(edge, selection);
			
			let data: FlowEdge["data"];
			if (edge.data) data = { ...edge.data, isHighlighted, isDimmed };

			return { ...edge, selected, data };
		});
	}

	private updateSelectionPresentation() {
		if (!this.displayedGraph || !this.displayedLayout) return;
		this.nodes = this.decorateNodes(this.displayedLayout.nodes, this.displayedGraph);
		this.edges = this.decorateEdges(this.displayedLayout.edges);
	}

	private selectedEntityRepresentatives(): ReadonlySet<string> {
		const representatives = new Set<string>();
		if (
			!this.displayedGraph ||
			!this.selectedTarget ||
			!inspectionTargetAvailable(this.displayedGraph, this.selectedTarget)
		) {
			return representatives;
		}

		for (const entityId of entityIdsForTarget(this.selectedTarget, this.displayedGraph)) {
			representatives.add(this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId);
		}
		return representatives;
	}

	private edgeIdForTarget(target: InspectionTarget): string | undefined {
		if (target.kind === "entity" || target.kind === "annotation") return undefined;

		return this.displayedProjection.connections.find((connection) => {
			if (target.kind === "relationship") {
				return connection.sourceRelationshipIds.includes(target.relationshipId);
			}
			return sameIds(connection.sourceRelationshipIds, target.sourceRelationshipIds);
		})?.id;
	}

	// Layout completion feedback keeps nearby detail convergent without changing the source snapshot.
	private updateDerivedNearby(
		graph: GraphSubset,
		projection: MapProjection,
		nodes: readonly FlowNode[]
	): boolean {
		const next = deriveNearbyEntityIds(graph, projection, nodes, this.viewport, this.canvasSize());
		if (sameIds(this.derivedNearbyEntityIds, next)) return false;
		this.derivedNearbyEntityIds = next;
		this.scheduleRebuild();
		return true;
	}

	private refreshNearby() {
		if (this.suppressNearbyRefresh || !this.displayedGraph || !this.displayedLayout) return;
		this.updateDerivedNearby(this.displayedGraph, this.displayedProjection, this.displayedLayout.nodes);
	}

	// Viewport/reveal mechanics. The bound viewport is updated only after MapViewportState processes it.
	private updateViewport(viewport: Viewport, event: MoveEvent = null) {
		const previousViewport = this.interaction.lastProcessedViewport;
		const zoomChanged = this.interaction.hasZoomChanged(viewport);
		let pointerAnchorId: string | undefined;
		const pointerZoom = zoomChanged && Boolean(event);
		if (pointerZoom && event) {
			const screenPoint = screenPointFromMoveEvent(event);
			pointerAnchorId =
				screenPoint && this.displayedLayout
					? nearestNodeIdAtScreenPoint(this.displayedLayout.nodes, previousViewport, screenPoint)
					: undefined;
		}

		const structuralChanged = this.interaction.updateViewport(viewport, pointerZoom, pointerAnchorId);
		this.viewport = this.interaction.lastProcessedViewport;
		this.syncDetailState();

		if (this.displayedGraph && this.displayedLayout && !structuralChanged) {
			this.nodes = this.decorateNodes(this.displayedLayout.nodes, this.displayedGraph);
		}
		this.options.onViewportChange?.(this.viewport);
		this.refreshNearby();
		if (structuralChanged) {
			this.scheduleRebuild();
		}
	}

	/** Svelte Flow callback; the public arrow field preserves the controller context. */
	onMove: OnMove = (_event, viewport) => {
		this.updateViewport(viewport, _event);
	};

	private setViewport(viewport: Viewport): boolean {
		const previousRequestId = this.requestId;
		this.updateViewport(viewport);
		return this.requestId !== previousRequestId;
	}

	// Public camera command plus its private framing helper.
	private fitToMaxZoom(maxZoom: number) {
		if (!this.displayedLayout || this.width <= 0 || this.height <= 0) return;
		const bounds = boundsForNodes(this.displayedLayout.nodes);
		if (!bounds) return;
		this.setViewport(getViewportForBounds(bounds, this.width, this.height, MAP_MIN_ZOOM, maxZoom, 0.3));
	}

	fit = () => {
		this.fitToMaxZoom(EXPLICIT_FIT_MAX_ZOOM);
	};

	// Layout completion framing is a one-time initial operation, not an automatic Fit during transitions.
	private fitInitialFrame() {
		if (!this.displayedLayout || this.width <= 0 || this.height <= 0) return;
		const bounds = boundsForNodes(this.displayedLayout.nodes);
		if (!bounds) return;
		this.setViewport(
			getViewportForBounds(
				bounds,
				this.width,
				this.height,
				MAP_MIN_ZOOM,
				initialFrameMaxZoom(this.structuralDetail),
				0.3
			)
		);
	}

	private applyInitialFrameIfReady(): boolean {
		if (this.initialFrameApplied || !this.displayedLayout || this.width <= 0 || this.height <= 0) {
			return false;
		}

		this.initialFrameApplied = true;
		this.suppressNearbyRefresh = true;
		try {
			this.fitInitialFrame();
		} finally {
			this.suppressNearbyRefresh = false;
		}
		return true;
	}

	// Viewport focus helpers used by Recenter and Reveal.
	private centerEntity(entityId: string): boolean {
		if (!this.displayedLayout || this.width <= 0 || this.height <= 0) return false;
		const representativeId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
		const bounds = worldBoundsByNodeId(this.displayedLayout.nodes).get(representativeId);
		if (!bounds) return false;
		this.setViewport(viewportCenteredOn(this.viewport, worldCenter(bounds), this.canvasSize()));
		return true;
	}

	private screenPointForEntity(entityId: string): Point | undefined {
		if (!this.displayedLayout) return undefined;
		const representativeId = this.displayedProjection.representativeByEntityId.get(entityId) ?? entityId;
		const bounds = worldBoundsByNodeId(this.displayedLayout.nodes).get(representativeId);
		if (!bounds) return undefined;
		const center = worldCenter(bounds);
		return screenPointFromWorld(this.viewport, center);
	}

	// Remaining public camera and reveal commands.
	recenter = () => {
		const focusEntityId = this.selectedEntityIds[0];
		if (focusEntityId && this.centerEntity(focusEntityId)) return;

		if (!this.displayedLayout || this.width <= 0 || this.height <= 0) return;
		const bounds = boundsForNodes(this.displayedLayout.nodes);
		if (!bounds) return;
		this.setViewport(viewportCenteredOn(this.viewport, worldCenter(bounds), this.canvasSize()));
	};

	zoomBy = (factor: number) => {
		const focusEntityId = this.selectedEntityIds[0];
		const focusRepresentativeId = focusEntityId
			? (this.displayedProjection.representativeByEntityId.get(focusEntityId) ?? focusEntityId)
			: undefined;
		const focusBounds =
			focusRepresentativeId && this.displayedLayout
				? worldBoundsByNodeId(this.displayedLayout.nodes).get(focusRepresentativeId)
				: undefined;
		const nextViewport = this.interaction.prepareKeyboardZoom(
			factor,
			focusEntityId,
			focusBounds,
			this.canvasSize()
		);
		this.setViewport(nextViewport);
	};

	revealSelected = () => {
		const entityIds = entityIdsForTarget(this.selectedTarget, this.suppliedGraph);
		if (!entityIds.length) return;

		let nextDetail = this.structuralDetail;
		for (const entityId of entityIds) {
			this.explicitRevealIds.add(entityId);
			const entity = this.suppliedGraph?.entities.find((candidate) => candidate.id === entityId);
			const display = entity ? getMapCategoryDisplay(entity.category) : undefined;
			if (display?.mode === DisplayMode.Node && display.level !== undefined) {
				nextDetail = Math.max(nextDetail, display.level);
			}
		}

		this.pendingRecenterEntityId = entityIds[0];
		const transition = this.interaction.prepareReveal(
			entityIds[0],
			nextDetail,
			this.screenPointForEntity(entityIds[0]) ?? {
				x: this.width > 0 ? this.width / 2 : 0,
				y: this.height > 0 ? this.height / 2 : 0,
			}
		);
		if (transition.zoomChanged) {
			if (!this.setViewport(transition.viewport)) this.scheduleRebuild();
			return;
		}

		this.syncDetailState();
		this.scheduleRebuild();
	};

	// Public selection and keyboard commands.
	private selectTarget(target: InspectionTarget, edgeId?: string, event?: Event) {
		if (this.suppliedGraph && !inspectionTargetAvailable(this.suppliedGraph, target)) return;
		this.selectedTarget = target;
		this.selectionRevision += 1;
		this.selectedEdgeId = edgeId ?? this.edgeIdForTarget(target);
		this.refreshInspection();
		this.updateSelectionPresentation();
		const trigger = this.selectionTrigger(event);
		this.options.onSelect?.({ target, ...(trigger ? { trigger } : {}) });
	}

	selectNode = (nodeId: string, event?: Event) => {
		this.selectTarget({ kind: "entity", entityId: nodeId }, undefined, event);
	};

	selectEdge = (edgeId: string, event?: Event) => {
		const connection = this.displayedProjection.connections.find((candidate) => candidate.id === edgeId);
		if (!connection || connection.sourceRelationshipIds.length === 0) return;

		const target =
			connection.classification === "direct"
				? { kind: "relationship" as const, relationshipId: connection.sourceRelationshipIds[0] }
				: { kind: "summary" as const, sourceRelationshipIds: connection.sourceRelationshipIds };
		this.selectTarget(target, edgeId, event);
	};

	selectItem = (target: InspectionTarget) => {
		this.selectTarget(target);
	};

	clearSelection = () => {
		this.selectedTarget = undefined;
		this.selectedEdgeId = undefined;
		this.inspection = undefined;
		this.options.onClearSelection?.();
		this.updateSelectionPresentation();
	};

	keydown = (event: KeyboardEvent) => {
		const target = event.target;
		if (
			target instanceof Element &&
			target.closest("input, textarea, select, button, [contenteditable=true]")
		) {
			return;
		}

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
		if (event.key === "Escape") {
			event.preventDefault();
			this.clearSelection();
			return;
		}

		if (!["Enter", " "].includes(event.key)) return;
		if (!(target instanceof Element)) return;
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
		return event.target.closest<HTMLElement>(".svelte-flow__node, .svelte-flow__edge") ?? undefined;
	}
}

const ctx = new Context<SystemMapController>("SystemMapController");

export function initSystemMapController(options: SystemMapControllerOptions) {
	return ctx.set(new SystemMapController(options));
}

export function useSystemMapController() {
	return ctx.get();
}
