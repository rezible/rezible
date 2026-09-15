import {
	getKnowledgeGraphViewOptions,
	getKnowledgeGraphEntityOptions,
	listKnowledgeGraphEntitiesOptions,
	type KnowledgeGraphEntity,
	type KnowledgeGraphRelationship,
} from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { MarkerType, type XYPosition } from "@xyflow/svelte";
import {
	SystemDiagramController,
	type SystemDiagramNode,
	type SystemDiagramEdge,
	type GraphSelection,
} from "$components/system-diagram";
import { Context, watch } from "runed";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useSearchParams } from "runed/kit";
import { SvelteMap, SvelteSet } from "svelte/reactivity";
import { z } from "zod";

const maxEntities = 200;
const maxRelationships = 400;

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

export function makeEntityLabel({ attributes }: KnowledgeGraphEntity) {
	return (
		attributes.latestState?.displayName ||
		attributes.aliases[0]?.attributes.resourceRef.resourceRef ||
		attributes.kind
	);
}

export class SystemMapViewController {
	private params = useSearchParams(paramsSchema, { pushHistory: true });

	private entities = new SvelteMap<string, KnowledgeGraphEntity>();
	private relationships = new SvelteMap<string, KnowledgeGraphRelationship>();
	private positions = new SvelteMap<string, XYPosition>();

	search = $state("");
	searchOpen = $state(false);
	diagram: SystemDiagramController;
	private framedFocus?: string;

	focusId = $derived(this.params.focus);
	selectedId = $derived(this.params.selected);
	depth = $derived(this.params.depth);
	displayMode = $derived(this.params.view);
	kindFilter = $derived(this.params.kind);
	predicateFilter = $derived(this.params.predicate);

	private viewQuery = createQuery(() =>
		getKnowledgeGraphViewOptions({
			query: {
				depth: this.params.depth,
				entityId: this.focusId || undefined,
			},
		})
	);

	private searchQuery = createPaginatedQuery({
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
	private selectionQuery = createQuery(() => ({
		...getKnowledgeGraphEntityOptions({ path: { id: this.selectedId } }),
		enabled: !!this.selectedId && !this.entities.has(this.selectedId),
	}));

	searchResults = $derived(this.searchQuery.query.data?.data ?? []);
	searching = $derived(this.searchQuery.query.isLoading || this.searchQuery.query.isFetching);
	loading = $derived(this.viewQuery.isLoading || this.viewQuery.isFetching);
	error = $derived(this.viewQuery.error);
	searchError = $derived(this.searchQuery.query.error);
	truncated = $derived(this.viewQuery.data?.data?.truncated ?? false);
	hasGraph = $derived(this.entities.size > 0);

	/** The focused subject is missing from the loaded neighborhood (e.g. unknown or unauthorized id). */
	focusMissing = $derived(
		!!this.focusId &&
			!this.loading &&
			!this.viewQuery.data?.data.entities.some((entity) => entity.id === this.focusId) &&
			!this.error
	);

	selected = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.selectedId === "") {
			return undefined;
		}
		const entity = this.entities.get(this.selectedId) ?? this.selectionQuery.data?.data;
		if (entity) {
			return { kind: "entity", entity };
		}
		return undefined;
	});
	selectionLoading = $derived(!!this.selectedId && !this.selected && this.selectionQuery.isPending);
	selectionError = $derived(this.selected ? undefined : this.selectionQuery.error);
	retrySelection = () => this.selectionQuery.refetch();

	/** Entity kinds present in the loaded neighborhood, for the subject-type filter. */
	availableKinds = $derived.by(() => {
		const kinds = new SvelteSet<string>();
		for (const entity of this.entities.values()) {
			kinds.add(entity.attributes.kind);
		}
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
		if (this.kindFilter === "") {
			return entities;
		}
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
		if (!this.selected || this.selected.kind !== "entity") {
			return [];
		}
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

	inspectedRelationship = $state<KnowledgeGraphRelationship>();
	private inspectionTrigger?: HTMLElement;
	private focusFallback?: HTMLElement;

	/** What the inspector shows: an explicit relationship inspection or the URL-selected subject. */
	inspector = $derived.by<SystemMapSelection | undefined>(() => {
		if (this.inspectedRelationship) {
			return { kind: "relationship", relationship: this.inspectedRelationship };
		}
		return this.selected;
	});

	hasFilters = $derived(this.kindFilter !== "" || this.predicateFilter !== "");

	constructor() {
		this.diagram = new SystemDiagramController({
			selection: () =>
				this.inspectedRelationship
					? { edgeId: this.inspectedRelationship.id }
					: { nodeId: this.selectedId },
			select: (selection, trigger) => this.selectDiagram(selection, trigger),
			onNodeMove: (id, position) => this.positions.set(id, position),
		});
		watch(
			() => [this.diagram.nodes, this.focusId] as const,
			([nodes, focus]) => {
				const id = focus || "";
				if (
					this.framedFocus === id ||
					!nodes.length ||
					(id && !nodes.some((node) => node.id === id))
				) {
					return;
				}
				this.framedFocus = id;
				if (id) {
					void this.diagram.focus({ nodeIds: [id], edgeIds: [] });
				} else {
					void this.diagram.fit();
				}
			}
		);
		watch(
			() => [this.viewQuery.data?.data, this.viewQuery.dataUpdatedAt] as const,
			([view]) => {
				if (view) {
					this.mergeView(view.entities, view.relationships);
				}
			}
		);
		watch(
			() => [this.kindFilter, this.predicateFilter],
			() => this.rebuildGraph()
		);
	}

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

	private selectDiagram(selection: GraphSelection, trigger?: HTMLElement) {
		if (selection.nodeId) {
			const entity = this.entities.get(selection.nodeId);
			if (entity) {
				this.selectEntity(entity);
			}
		} else if (selection.edgeId) {
			const relationship = this.relationships.get(selection.edgeId);
			if (relationship) {
				this.selectRelationship(relationship);
			}
		} else {
			this.clearSelection();
		}
		if (trigger) {
			this.inspectionTrigger = trigger;
		}
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

	restoreInspectionFocus = (event: Event) => {
		let target = this.inspectionTrigger;
		if (!target?.isConnected) {
			target = this.focusFallback;
		}
		if (!target?.isConnected) {
			return;
		}
		event.preventDefault();
		target.focus({ preventScroll: true });
	};

	setFocusFallback = (target: HTMLElement) => {
		this.focusFallback = target;
		return () => {
			this.focusFallback = undefined;
		};
	};

	recenter() {
		const id = this.selectedId || this.focusId;
		const nodeIsVisible = id && this.diagram.nodes.some((node) => node.id === id);
		if (nodeIsVisible) {
			void this.diagram.focus({ nodeIds: [id], edgeIds: [] });
		} else {
			void this.diagram.fit();
		}
	}

	entityLabel(id: string) {
		const entity = this.entities.get(id);
		return entity ? makeEntityLabel(entity) : "Unknown subject";
	}

	connectionCount(id: string) {
		let count = 0;
		for (const rel of this.relationships.values()) {
			if (rel.attributes.sourceEntityId === id || rel.attributes.targetEntityId === id) {
				count++;
			}
		}
		return count;
	}

	reset() {
		this.params.update({ focus: "", selected: "", kind: "", predicate: "" });
		this.inspectedRelationship = undefined;
		this.entities.clear();
		this.relationships.clear();
		this.positions.clear();
		this.framedFocus = undefined;
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
			if (this.entities.size >= maxEntities && !this.entities.has(entity.id)) {
				break;
			}
			this.entities.set(entity.id, entity);
		}
		for (const rel of relationships) {
			if (this.relationships.size >= maxRelationships && !this.relationships.has(rel.id)) {
				break;
			}
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
		if (rootId && this.entities.has(rootId)) {
			distances.set(rootId, 0);
		}
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
			if (this.positions.has(entity.id)) {
				continue;
			}
			const layer = distances.get(entity.id) ?? 5;
			const entries = byLayer.get(layer) ?? [];
			entries.push(entity);
			byLayer.set(layer, entries);
		}

		for (const [layer, entities] of [...byLayer.entries()].sort(([a], [b]) => a - b)) {
			entities.sort((a, b) => makeEntityLabel(a).localeCompare(makeEntityLabel(b)));
			const height = (entities.length - 1) * 120;
			entities.forEach((entity, index) => {
				this.positions.set(entity.id, {
					x: layer * 320,
					y: index * 120 - height / 2,
				});
			});
		}

		const displayIds = new SvelteSet(this.displayEntities.map((entity) => entity.id));
		const nodes: SystemDiagramNode[] = [];
		for (const entity of this.displayEntities) {
			const position = this.positions.get(entity.id);
			if (!position) {
				continue;
			}
			nodes.push({
				id: entity.id,
				type: "entity",
				position,
				data: { entity },
			});
		}
		const edges: SystemDiagramEdge[] = [];
		for (const relationship of this.displayRelationships) {
			const { id, attributes: attrs } = relationship;
			if (!displayIds.has(attrs.sourceEntityId) || !displayIds.has(attrs.targetEntityId)) {
				continue;
			}
			const label = attrs.latestState?.displayName || attrs.predicate.replaceAll("_", " ");
			edges.push({
				id,
				type: "relationship",
				source: attrs.sourceEntityId,
				target: attrs.targetEntityId,
				label,
				data: { relationship },
				markerEnd: MarkerType.ArrowClosed,
			});
		}
		this.diagram.setGraph(nodes, edges);
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export function initSystemMapViewController() {
	return ctx.set(new SystemMapViewController());
}

export function useSystemMapViewController() {
	return ctx.get();
}
