import {
	getKnowledgeGraphViewOptions,
	getKnowledgeGraphEntityOptions,
	listKnowledgeGraphEntitiesOptions,
	type ErrorModel,
	type KnowledgeGraphEntity,
	type KnowledgeGraphRelationship,
} from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { MarkerType, type Edge, type Node, type Viewport } from "@xyflow/svelte";
import { Context, watch, type Getter } from "runed";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useSearchParams } from "runed/kit";
import { SvelteMap, SvelteSet } from "svelte/reactivity";
import { z } from "zod";
import { GraphFraming } from "./framing";

const maxEntities = 200;
const maxRelationships = 400;

export type SystemMapNodeData = {
	entity: KnowledgeGraphEntity;
};

export type SystemMapEdgeData = {
	relationship: KnowledgeGraphRelationship;
};

export type SystemMapSelection =
	| { kind: "entity"; entity: KnowledgeGraphEntity }
	| { kind: "relationship"; relationship: KnowledgeGraphRelationship };

const paramsSchema = z.object({
	focus: z.string().default("").catch(""),
	selected: z.string().default("").catch(""),
	depth: z.number().min(1).max(4).default(2).catch(2),
	view: z.enum(["graph", "list"]).default("graph").catch("graph"),
	kind: z.string().default("").catch(""),
	predicate: z.string().default("").catch(""),
});

export const makeEntityLabel = ({ attributes: attrs }: KnowledgeGraphEntity) => {
	return (
		attrs.latestState?.displayName || attrs.aliases[0]?.attributes.resourceRef.resourceRef || attrs.kind
	);
};

export class SystemMapViewController {
	private params = useSearchParams(paramsSchema, { pushHistory: true });
	private defaultFocus: Getter<string | undefined>;

	private entities = new SvelteMap<string, KnowledgeGraphEntity>();
	private relationships = new SvelteMap<string, KnowledgeGraphRelationship>();
	private positions = new SvelteMap<string, { x: number; y: number }>();

	search = $state("");
	searchOpen = $state(false);
	nodes = $state.raw<Node<SystemMapNodeData>[]>([]);
	edges = $state.raw<Edge<SystemMapEdgeData>[]>([]);
	viewport = $state<Viewport>({ x: 80, y: 80, zoom: 0.9 });
	canvasWidth = $state(0);
	canvasHeight = $state(0);
	private framing = new GraphFraming();

	private viewQuery;
	private searchQuery;
	private selectionQuery;

	constructor(defaultFocus: Getter<string | undefined> = () => "") {
		this.defaultFocus = defaultFocus;
		watch(
			() => [this.nodes, this.canvasWidth, this.canvasHeight, this.focusId] as const,
			([nodes, width, height]) => {
				const viewport = this.framing.initial(nodes, width, height, this.focusId || "");
				if (viewport) this.viewport = viewport;
			}
		);
		this.viewQuery = createQuery(() => ({
			...getKnowledgeGraphViewOptions({
				query: {
					depth: this.params.depth,
					entityId: this.focusId || undefined,
				},
			}),
			enabled: this.focusId !== undefined,
		}));

		this.searchQuery = createPaginatedQuery({
			source: "local",
			queryOptions: () => ({
				...listKnowledgeGraphEntitiesOptions({
					query: {
						search: this.search.trim(),
						page: 1,
						pageSize: 20,
					},
				}),
				enabled: this.search.trim().length >= 2,
			}),
		});
		this.selectionQuery = createQuery(() => ({
			...getKnowledgeGraphEntityOptions({ path: { id: this.selectedId } }),
			enabled: !!this.selectedId && !this.entities.has(this.selectedId),
		}));
		watch(
			() => [this.viewQuery.data?.data, this.viewQuery.dataUpdatedAt] as const,
			([view]) => {
				if (view) this.mergeView(view.entities, view.relationships);
			}
		);
		watch(
			() => [this.kindFilter, this.predicateFilter],
			() => this.rebuildGraph()
		);
	}

	focusId = $derived.by(() => this.params.focus || this.defaultFocus());
	selectedId = $derived.by(() => this.params.selected);
	depth = $derived.by(() => this.params.depth);
	displayMode = $derived.by(() => this.params.view);
	kindFilter = $derived.by(() => this.params.kind);
	predicateFilter = $derived.by(() => this.params.predicate);

	searchResults = $derived.by(() => this.searchQuery?.query.data?.data ?? []);
	searching = $derived.by(() => this.searchQuery?.query.isLoading || this.searchQuery?.query.isFetching);
	loading = $derived.by(() => this.viewQuery?.isLoading || this.viewQuery?.isFetching);
	error = $derived.by(() => this.viewQuery?.error as ErrorModel | undefined);
	searchError = $derived.by(() => this.searchQuery?.query.error);
	truncated = $derived.by(() => this.viewQuery?.data?.data?.truncated ?? false);
	hasGraph = $derived.by(() => this.entities.size > 0);

	/** The focused subject is missing from the loaded neighborhood (e.g. unknown or unauthorized id). */
	focusMissing = $derived.by(
		() =>
			!!this.focusId &&
			!this.loading &&
			!this.viewQuery?.data?.data.entities.some((entity) => entity.id === this.focusId) &&
			!this.error
	);

	selected = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.selectedId === "") return undefined;
		const entity = this.entities.get(this.selectedId) ?? this.selectionQuery?.data?.data;
		if (entity) return { kind: "entity", entity };
		return undefined;
	});
	selectionLoading = $derived.by(
		() => !!this.selectedId && !this.selected && this.selectionQuery.isPending
	);
	selectionError = $derived.by(() => (!this.selected ? this.selectionQuery.error : undefined));
	retrySelection = () => this.selectionQuery.refetch();

	/** Entity kinds present in the loaded neighborhood, for the subject-type filter. */
	availableKinds = $derived.by(() => {
		const kinds = new SvelteSet<string>();
		for (const entity of this.entities.values()) kinds.add(entity.attributes.kind);
		return [...kinds].sort();
	});

	/** Relationship predicates present in the loaded neighborhood, for the relationship filter. */
	availablePredicates = $derived.by(() => {
		const predicates = new SvelteSet<string>();
		for (const relationship of this.relationships.values()) {
			predicates.add(relationship.attributes.predicate);
		}
		return [...predicates].sort();
	});

	displayEntities = $derived.by(() => {
		const entities = [...this.entities.values()];
		if (this.kindFilter === "") return entities;
		return entities.filter((entity) => entity.attributes.kind === this.kindFilter);
	});

	displayRelationships = $derived.by(() => {
		let relationships = [...this.relationships.values()];
		if (this.kindFilter !== "") {
			const visible = new SvelteSet(this.displayEntities.map((entity) => entity.id));
			relationships = relationships.filter(
				(rel) =>
					visible.has(rel.attributes.sourceEntityId) && visible.has(rel.attributes.targetEntityId)
			);
		}
		if (this.predicateFilter !== "") {
			relationships = relationships.filter((rel) => rel.attributes.predicate === this.predicateFilter);
		}
		return relationships;
	});

	/** Summary of a selected entity's relationships grouped by predicate. */
	relationshipSummary = $derived.by(() => {
		if (!this.selected || this.selected.kind !== "entity") return [];
		const byPredicate = new SvelteMap<string, KnowledgeGraphRelationship[]>();
		for (const rel of this.relationships.values()) {
			const { sourceEntityId, targetEntityId, predicate } = rel.attributes;
			if (sourceEntityId !== this.selected.entity.id && targetEntityId !== this.selected.entity.id) {
				continue;
			}
			const entries = byPredicate.get(predicate) ?? [];
			entries.push(rel);
			byPredicate.set(predicate, entries);
		}
		return [...byPredicate.entries()].sort(([a], [b]) => a.localeCompare(b));
	});

	setDepth(depth: number) {
		this.params.depth = depth;
	}

	setDisplayMode(mode: "graph" | "list") {
		this.params.view = mode;
	}

	setKindFilter(kind: string) {
		this.params.kind = kind;
		if (
			this.selected?.kind === "entity" &&
			kind !== "" &&
			this.selected.entity.attributes.kind !== kind
		) {
			this.clearSelection();
		}
	}

	setPredicateFilter(predicate: string) {
		this.params.predicate = predicate;
	}

	clearFilters() {
		this.params.kind = "";
		this.params.predicate = "";
	}

	hasFilters = $derived.by(() => this.kindFilter !== "" || this.predicateFilter !== "");

	focus(entity: KnowledgeGraphEntity) {
		this.entities.set(entity.id, entity);
		this.params.update({ focus: entity.id, selected: entity.id });
		this.searchOpen = false;
		this.search = "";
	}

	expand(entity: KnowledgeGraphEntity) {
		this.focus(entity);
	}

	selectEntity(entity: KnowledgeGraphEntity) {
		this.inspectionTrigger =
			document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
		this.inspectedRelationship = undefined;
		this.params.selected = entity.id;
	}

	selectRelationship(relationship: KnowledgeGraphRelationship) {
		this.inspectionTrigger =
			document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
		// Relationship selection is transient inspection, not a durable subject.
		this.inspectedRelationship = relationship;
	}

	clearSelection() {
		this.params.selected = "";
		this.inspectedRelationship = undefined;
	}

	inspectedRelationship = $state<KnowledgeGraphRelationship>();
	private inspectionTrigger?: HTMLElement;
	private focusFallback?: HTMLElement;

	restoreInspectionFocus = (event: Event) => {
		const target = this.inspectionTrigger?.isConnected
			? this.inspectionTrigger
			: this.focusFallback?.isConnected
				? this.focusFallback
				: undefined;
		if (!target) return;
		event.preventDefault();
		target.focus({ preventScroll: true });
	};

	setFocusFallback = (target: HTMLElement | undefined) => {
		this.focusFallback = target;
	};

	/** What the inspector shows: an explicit relationship inspection or the URL-selected subject. */
	inspector = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.inspectedRelationship)
			return { kind: "relationship", relationship: this.inspectedRelationship };
		return this.selected;
	});

	recenter() {
		const viewport = this.framing.fit(
			this.nodes,
			this.canvasWidth,
			this.canvasHeight,
			this.selectedId || this.focusId
		);
		if (viewport) this.viewport = viewport;
	}

	entityLabel = (id: string) => {
		const entity = this.entities.get(id);
		return entity ? makeEntityLabel(entity) : "Unknown subject";
	};

	connectionCount = (id: string) => {
		let count = 0;
		for (const rel of this.relationships.values()) {
			if (rel.attributes.sourceEntityId === id || rel.attributes.targetEntityId === id) count++;
		}
		return count;
	};

	reset() {
		this.params.update({ focus: "", selected: "", kind: "", predicate: "" });
		this.inspectedRelationship = undefined;
		this.entities.clear();
		this.relationships.clear();
		this.positions.clear();
		this.framing = new GraphFraming();
		this.viewport = { x: 80, y: 80, zoom: 0.9 };
		this.rebuildGraph();
		void this.viewQuery.refetch();
	}

	retry = () => this.viewQuery.refetch();
	retrySearch = () => this.searchQuery.query.refetch();

	closeSearch() {
		this.searchOpen = false;
	}

	private mergeView(entities: KnowledgeGraphEntity[], relationships: KnowledgeGraphRelationship[]) {
		for (const entity of entities) {
			if (this.entities.size >= maxEntities && !this.entities.has(entity.id)) break;
			this.entities.set(entity.id, entity);
		}
		for (const rel of relationships) {
			if (this.relationships.size >= maxRelationships && !this.relationships.has(rel.id)) break;
			const { sourceEntityId, targetEntityId } = rel.attributes;
			if (this.entities.has(sourceEntityId) && this.entities.has(targetEntityId)) {
				this.relationships.set(rel.id, rel);
			}
		}
		this.rebuildGraph();
	}

	private rebuildGraph() {
		const distances = new SvelteMap<string, number>();
		const rootId = this.focusId;
		if (rootId && this.entities.has(rootId)) distances.set(rootId, 0);
		for (let pass = 0; pass < 4; pass++) {
			for (const relationship of this.relationships.values()) {
				const { sourceEntityId: source, targetEntityId: target } = relationship.attributes;
				const sourceDistance = distances.get(source);
				const targetDistance = distances.get(target);
				if (sourceDistance !== undefined && targetDistance === undefined) {
					distances.set(target, sourceDistance + 1);
				}
				if (targetDistance !== undefined && sourceDistance === undefined) {
					distances.set(source, targetDistance + 1);
				}
			}
		}

		// Preserve positions of already-placed entities so progressive loads
		// never move the neighborhood the user is looking at.
		const byLayer = new SvelteMap<number, KnowledgeGraphEntity[]>();
		for (const entity of this.entities.values()) {
			if (this.positions.has(entity.id)) continue;
			const layer = distances.get(entity.id) ?? 5;
			const entries = byLayer.get(layer) ?? [];
			entries.push(entity);
			byLayer.set(layer, entries);
		}

		for (const [layer, entities] of [...byLayer.entries()].sort(([a], [b]) => a - b)) {
			entities.sort((a, b) => makeEntityLabel(a).localeCompare(makeEntityLabel(b)));
			const existing = entities.filter((entity) => this.positions.has(entity.id)).length;
			const height = (entities.length - 1) * 120;
			entities.forEach((entity, index) => {
				const offset = existing + index;
				this.positions.set(entity.id, {
					x: layer * 320,
					y: offset * 120 - height / 2,
				});
			});
		}

		const displayIds = new SvelteSet(this.displayEntities.map((entity) => entity.id));
		const nodes: Node<SystemMapNodeData>[] = [];
		for (const entity of this.displayEntities) {
			const position = this.positions.get(entity.id);
			if (!position) continue;
			nodes.push({
				id: entity.id,
				type: "entity",
				position,
				data: { entity },
			});
		}
		this.nodes = nodes;

		const edges: Edge<SystemMapEdgeData>[] = [];
		for (const relationship of this.displayRelationships) {
			const { id, attributes: attrs } = relationship;
			if (!displayIds.has(attrs.sourceEntityId) || !displayIds.has(attrs.targetEntityId)) continue;
			const label = attrs.latestState?.displayName || attrs.predicate.replaceAll("_", " ");
			edges.push({
				id: id,
				type: "default",
				source: attrs.sourceEntityId,
				target: attrs.targetEntityId,
				label: label,
				data: { relationship },
				markerEnd: MarkerType.ArrowClosed,
			});
		}
		this.edges = edges;
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const provideSystemMapViewController = (controller: SystemMapViewController) => ctx.set(controller);
export const useSystemMapViewController = () => ctx.get();
