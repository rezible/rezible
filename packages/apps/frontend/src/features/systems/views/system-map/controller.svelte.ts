import {
	getKnowledgeGraphStructureOptions,
	type KnowledgeGraphStructureEntity,
	type KnowledgeGraphStructureRelationship,
	getKnowledgeGraphEntityOptions,
	listKnowledgeGraphEntitiesOptions,
	type KnowledgeGraphEntity,
	type KnowledgeGraphRelationship,
	type KnowledgeGraphStructure,
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
	private entities = new SvelteMap<string, KnowledgeGraphStructureEntity>();
	private relationships = new SvelteMap<string, KnowledgeGraphStructureRelationship>();
	private positions = new SvelteMap<string, XYPosition>();

	diagram: SystemDiagramController;
	private framedFocus?: string;

	private params = useSearchParams(paramsSchema, { pushHistory: true });
	focusId = $derived(this.params.focus);
	selectedId = $derived(this.params.selected);
	depth = $derived(this.params.depth);
	displayMode = $derived(this.params.view);
	kindFilter = $derived(this.params.kind);
	predicateFilter = $derived(this.params.predicate);

	search = $state("");
	searchOpen = $state(false);
	private trimmedSearch = $derived(this.search.trim());

	private searchQuery = createPaginatedQuery({
		source: "local",
		queryOptions: () => ({
			...listKnowledgeGraphEntitiesOptions({
				query: {
					search: this.trimmedSearch,
					page: 1,
					pageSize: 20,
				},
			}),
			enabled: this.trimmedSearch.length >= 2,
		}),
	});

	searchResults = $derived(this.searchQuery.query.data?.data ?? []);
	searching = $derived(this.searchQuery.query.isLoading || this.searchQuery.query.isFetching);
	searchError = $derived(this.searchQuery.query.error);

	private structureQuery = createQuery(() =>
		getKnowledgeGraphStructureOptions({})
	);
	loading = $derived(this.structureQuery.isLoading || this.structureQuery.isFetching);
	error = $derived(this.structureQuery.error);
	truncated = $derived(false);
	hasGraph = $derived(this.entities.size > 0);

	/** Entity kinds present in the loaded neighborhood, for the subject-type filter. */
	availableKinds = $derived.by(() => {
		const kinds = new SvelteSet<string>();
		for (const entity of this.entities.values()) {
			kinds.add(entity.kind);
		}
		return [...kinds].sort();
	});

	/** Relationship predicates present in the loaded neighborhood, for the relationship filter. */
	availablePredicates = $derived.by(() => {
		const predicates = new SvelteSet<string>();
		for (const rel of this.relationships.values()) {
			predicates.add(rel.predicate);
		}
		return [...predicates].sort();
	});

	displayEntities = $derived.by(() => {
		const entities = [...this.entities.values()];
		if (this.kindFilter === "") {
			return entities;
		}
		return entities.filter((entity) => entity.kind === this.kindFilter);
	});

	displayRelationships = $derived.by(() => {
		let relationships = [...this.relationships.values()];
		if (this.kindFilter !== "") {
			const visible = new SvelteSet(this.displayEntities.map((entity) => entity.id));
			relationships = relationships.filter(
				(rel) =>
					visible.has(rel.sourceId) && visible.has(rel.targetId)
			);
		}
		if (this.predicateFilter !== "") {
			relationships = relationships.filter((rel) => rel.predicate === this.predicateFilter);
		}
		return relationships;
	});

	private selectionQuery = createQuery(() => ({
		...getKnowledgeGraphEntityOptions({ path: { id: this.selectedId } }),
		enabled: !!this.selectedId,
	}));
	selectedEntity = $derived.by(() => {
		if (!this.selectedId) return;
		const entity = this.entities.get(this.selectedId) ?? this.selectionQuery.data?.data;
		return !!entity ? { kind: "entity", entity } as SystemMapSelection : undefined;
	});
	selectionLoading = $derived(!!this.selectedId && !this.selectedEntity && this.selectionQuery.isPending);
	selectionError = $derived(this.selectedEntity ? undefined : this.selectionQuery.error);
	
	retrySelection = () => this.selectionQuery.refetch();

	/** Summary of a selected entity's relationships grouped by predicate. */
	selectedEntityRelationshipsSummary = $derived.by(() => {
		if (!this.selectedEntity || this.selectedEntity.kind !== "entity") {
			return [];
		}
		const byPredicate = new SvelteMap<string, KnowledgeGraphStructureRelationship[]>();
		for (const rel of this.relationships.values()) {
			if (rel.sourceId !== this.selectedEntity.entity.id && rel.targetId !== this.selectedEntity.entity.id) {
				continue;
			}
			const entries = byPredicate.get(rel.predicate) ?? [];
			entries.push(rel);
			byPredicate.set(rel.predicate, entries);
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
		return this.selectedEntity;
	});

	selection = $derived(this.inspectedRelationship
					? { edgeId: this.inspectedRelationship.id }
					: { nodeId: this.selectedId });


	hasFilters = $derived(this.kindFilter !== "" || this.predicateFilter !== "");

	constructor() {
		this.diagram = new SystemDiagramController({
			selection: () => this.selection,
			select: (selection, trigger) => this.selectDiagram(selection, trigger),
			onNodeMove: (id, position) => this.positions.set(id, position),
		});
		watch(
			() => [this.diagram.nodes, this.focusId] as const,
			([nodes, focus]) => {
				const id = focus || "";
				if (this.framedFocus === id || !nodes.length) return;
				if (id && !nodes.some((node) => node.id === id)) return;
				this.framedFocus = id;
				if (id) {
					void this.diagram.focus({ nodeIds: [id], edgeIds: [] });
				} else {
					void this.diagram.fit();
				}
			}
		);
		watch(
			() => [this.structureQuery.data?.data, this.structureQuery.dataUpdatedAt] as const,
			([structure]) => {
				if (structure) this.mergeView(structure);
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
			this.selectedEntity?.kind === "entity" &&
			kind !== "" &&
			this.selectedEntity.entity.attributes.kind !== kind
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

	focus(entityId: string) {
		// this.entities.set(entity.id, entity);
		// this.params.update({ focus: entity.id, selected: entity.id });
		this.searchOpen = false;
		this.search = "";
	}

	expand(entity: KnowledgeGraphEntity) {
		// this.focus(entity);
	}

	selectEntity(entityId: string) {
		this.inspectionTrigger =
			document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
		this.inspectedRelationship = undefined;
		this.params.selected = entityId;
	}

	selectRelationship(relationship: KnowledgeGraphRelationship) {
		this.inspectionTrigger =
			document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
		this.inspectedRelationship = relationship;
	}

	private selectDiagram(selection: GraphSelection, trigger?: HTMLElement) {
		if (selection.nodeId) {
			const entity = this.entities.get(selection.nodeId);
			if (entity) this.selectEntity(entity.id);
		} else if (selection.edgeId) {
			const relationship = this.relationships.get(selection.edgeId);
			// if (relationship) this.selectRelationship(relationship);
		} else {
			this.clearSelection();
		}
		if (trigger) {
			this.inspectionTrigger = trigger;
		}
	}

	clearSelection() {
		this.params.selected = "";
		this.inspectedRelationship = undefined;
	}

	restoreInspectionFocus = (event: Event) => {
		let target = this.inspectionTrigger;
		if (!target?.isConnected) {
			if (!this.focusFallback?.isConnected) return;
			target = this.focusFallback;
		}
		event.preventDefault();
		target.focus({ preventScroll: true });
	};

	setFocusFallback = (target: HTMLElement) => {
		this.focusFallback = target;
		return () => { this.focusFallback = undefined };
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
		return this.entities.get(id)?.kind || "Unknown subject";
	}

	connectionCount(id: string) {
		let count = 0;
		for (const rel of this.relationships.values()) {
			if (rel.sourceId === id || rel.targetId === id) {
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
		void this.structureQuery.refetch();
	}

	retry = () => this.structureQuery.refetch();
	retrySearch = () => this.searchQuery.query.refetch();

	closeSearch() {
		this.searchOpen = false;
	}

	private mergeView({entities, relationships}: KnowledgeGraphStructure) {
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
			if (this.entities.has(rel.sourceId) && this.entities.has(rel.targetId)) {
				this.relationships.set(rel.id, rel);
			}
		}
		this.rebuildGraph();
	}

	private rebuildGraph() {
		const rootId = this.focusId;

		const distances = new SvelteMap<string, number>();
		if (rootId && this.entities.has(rootId)) distances.set(rootId, 0);
		for (let pass = 0; pass < 4; pass++) {
			for (const rel of this.relationships.values()) {
				const sourceDistance = distances.get(rel.sourceId);
				const targetDistance = distances.get(rel.targetId);
				if (sourceDistance !== undefined && targetDistance === undefined) {
					distances.set(rel.targetId, sourceDistance + 1);
				}
				if (targetDistance !== undefined && sourceDistance === undefined) {
					distances.set(rel.sourceId, targetDistance + 1);
				}
			}
		}

		// Preserve positions of already-placed entities so progressive loads
		// never move the neighborhood the user is looking at.
		const byLayer = new SvelteMap<number, KnowledgeGraphStructureEntity[]>();
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
			entities.sort((a, b) => a.kind.localeCompare(b.kind));
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
			// nodes.push({
			// 	id: entity.id,
			// 	type: "entity",
			// 	position,
			// 	data: { entity },
			// });
		}
		const edges: SystemDiagramEdge[] = [];
		for (const rel of this.displayRelationships) {
			if (!displayIds.has(rel.sourceId) || !displayIds.has(rel.targetId)) {
				continue;
			}
			const label = rel.predicate.replaceAll("_", " ");
			// edges.push({
			// 	id: rel.id,
			// 	type: "relationship",
			// 	source: rel.sourceId,
			// 	target: rel.targetId,
			// 	label,
			// 	data: { relationship: rel },
			// 	markerEnd: MarkerType.ArrowClosed,
			// });
		}
		this.diagram.setGraph(nodes, edges);
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const initSystemMapViewController = () => ctx.set(new SystemMapViewController());
export const useSystemMapViewController = () => ctx.get();
