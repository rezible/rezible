import {
	getKnowledgeGraphViewOptions,
	listKnowledgeGraphEntitiesOptions,
	type ErrorModel,
	type KnowledgeGraphEntity,
	type KnowledgeGraphRelationship,
} from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { MarkerType, type Edge, type Node, type Viewport } from "@xyflow/svelte";
import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { SvelteMap } from "svelte/reactivity";
import { z } from "zod";

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
	private params = useSearchParams(paramsSchema);

	private entities = new SvelteMap<string, KnowledgeGraphEntity>();
	private relationships = new SvelteMap<string, KnowledgeGraphRelationship>();
	private positions = new SvelteMap<string, { x: number; y: number }>();

	search = $state("");
	searchOpen = $state(false);
	nodes = $state.raw<Node<SystemMapNodeData>[]>([]);
	edges = $state.raw<Edge<SystemMapEdgeData>[]>([]);
	viewport = $state<Viewport>({ x: 80, y: 80, zoom: 0.9 });

	private viewQuery = createQuery(() => ({
		...getKnowledgeGraphViewOptions({
			query: {
				depth: this.params.depth,
				entityId: this.params.focus ? this.params.focus : undefined,
			},
		}),
	}));

	private searchQuery = createQuery(() => ({
		...listKnowledgeGraphEntitiesOptions({
			query: {
				search: this.search.trim(),
				page: 1,
				pageSize: 20,
			},
		}),
		enabled: this.search.trim().length >= 2,
	}));

	constructor() {
		watch(
			() => this.viewQuery.data?.data,
			(view) => {
				if (view) this.mergeView(view.entities, view.relationships);
			}
		);
	}

	focusId = $derived(this.params.focus);
	selectedId = $derived(this.params.selected);
	depth = $derived(this.params.depth);
	displayMode = $derived(this.params.view);
	kindFilter = $derived(this.params.kind);
	predicateFilter = $derived(this.params.predicate);

	searchResults = $derived(this.searchQuery.data?.data ?? []);
	searching = $derived(this.searchQuery.isLoading || this.searchQuery.isFetching);
	loading = $derived(this.viewQuery.isLoading || this.viewQuery.isFetching);
	error = $derived((this.viewQuery.error ?? this.searchQuery.error) as ErrorModel | undefined);
	truncated = $derived(this.viewQuery.data?.data?.truncated ?? false);
	hasGraph = $derived(this.entities.size > 0);

	/** The focused subject is missing from the loaded neighborhood (e.g. unknown or unauthorized id). */
	focusMissing = $derived(
		this.focusId !== "" && !this.loading && !this.entities.has(this.focusId) && !this.error
	);

	selected = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.selectedId === "") return undefined;
		const entity = this.entities.get(this.selectedId);
		if (entity) return { kind: "entity", entity };
		return undefined;
	});

	/** Entity kinds present in the loaded neighborhood, for the subject-type filter. */
	availableKinds = $derived.by(() => {
		const kinds = new Set<string>();
		for (const entity of this.entities.values()) kinds.add(entity.attributes.kind);
		return [...kinds].sort();
	});

	/** Relationship predicates present in the loaded neighborhood, for the relationship filter. */
	availablePredicates = $derived.by(() => {
		const predicates = new Set<string>();
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
			const visible = new Set(this.displayEntities.map((entity) => entity.id));
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
		const byPredicate = new Map<string, KnowledgeGraphRelationship[]>();
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

	hasFilters = $derived(this.kindFilter !== "" || this.predicateFilter !== "");

	focus(entity: KnowledgeGraphEntity) {
		this.entities.set(entity.id, entity);
		this.params.focus = entity.id;
		this.params.selected = entity.id;
		this.searchOpen = false;
		this.search = "";
	}

	expand(entity: KnowledgeGraphEntity) {
		this.focus(entity);
	}

	selectEntity(entity: KnowledgeGraphEntity) {
		this.params.selected = entity.id;
	}

	selectRelationship(relationship: KnowledgeGraphRelationship) {
		// Relationship selection is transient inspection, not a durable subject.
		this.inspectedRelationship = relationship;
	}

	clearSelection() {
		this.params.selected = "";
		this.inspectedRelationship = undefined;
	}

	inspectedRelationship = $state<KnowledgeGraphRelationship>();

	/** What the inspector shows: an explicit relationship inspection or the URL-selected subject. */
	inspector = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.inspectedRelationship)
			return { kind: "relationship", relationship: this.inspectedRelationship };
		return this.selected;
	});

	recenter() {
		if (!this.selectedId) return;
		this.focusEntity(this.selectedId);
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
		this.params.focus = "";
		this.params.selected = "";
		this.params.kind = "";
		this.params.predicate = "";
		this.entities.clear();
		this.relationships.clear();
		this.positions.clear();
	}

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

		const displayIds = new Set(this.displayEntities.map((entity) => entity.id));
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

		if (this.selectedId) this.focusEntity(this.selectedId);
	}

	private focusEntity(id: string) {
		const node = this.nodes.find((candidate) => candidate.id === id);
		if (!node) return;
		this.viewport = {
			x: 420 - node.position.x,
			y: 280 - node.position.y,
			zoom: 1,
		};
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const initSystemMapViewController = () => ctx.set(new SystemMapViewController());
export const useSystemMapViewController = () => ctx.get();
