import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { onDestroy } from "svelte";
import { useQueryClient } from "@tanstack/svelte-query";
import { z } from "zod";

import { type KnowledgeGraphEntitySummary } from "$lib/api";
import { getMapCategoryDisplay, isArchitectureCategory } from "$features/system/lib/system-map/category";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { MapSelection } from "$features/system/lib/system-map/presentation";
import { reconcileSelection } from "$features/system/lib/system-map/selection";
import { GraphQueryController, type PublishedGraph, type StandaloneGraphQuery } from "./loader.svelte";

type EntityCategory = KnowledgeGraphEntitySummary["category"];

export const systemMapCategoryOptions = [
	"system_function",
	"system",
	"container",
	"infrastructure",
	"component",
	"code",
	"actor",
	"process",
	"concern",
	"decision",
	"event",
	"signal",
] as const satisfies readonly EntityCategory[];

const filterSchema = z.object({
	categories: z.array(z.enum(systemMapCategoryOptions)).default([]).catch([]),
	kinds: z.array(z.string()).default([]).catch([]),
});

const emptyGraph: GraphSubset = {
	entities: [],
	relationships: [],
	unresolvedRelationships: [],
	enumeration: { scope: "query", stopReason: "exhausted" },
};

const normalizeQuery = (query: StandaloneGraphQuery): StandaloneGraphQuery => ({
	categories: [...new Set(query.categories)].sort(),
	kinds: [...new Set(query.kinds)].sort(),
});

const queryKey = (query: StandaloneGraphQuery) => JSON.stringify([query.categories, query.kinds]);

const queryDescription = (query: StandaloneGraphQuery): string => {
	const categories = query.categories.length
		? query.categories.map((category) => getMapCategoryDisplay(category).categoryLabel).join(", ")
		: "all categories";
	const kinds = query.kinds.length 
		? query.kinds.join(", ") 
		: "all kinds";
	return `${categories}; ${kinds}`;
};

const stopReasonDescription: Record<GraphSubset["enumeration"]["stopReason"], string> = {
	exhausted: "The available pages were exhausted.",
	"entity-limit": "The entity limit stopped this load; more records may be available.",
	"relationship-limit": "The relationship limit stopped this load; more records may be available.",
	"request-limit": "The request limit stopped this load; more records may be available.",
};

export class SystemMapViewController {
	readonly graphQuery = new GraphQueryController(useQueryClient());
	private params = useSearchParams(filterSchema, { noScroll: true });

	filters = $derived(filterSchema.parse(this.params));
	requestedQuery = $derived.by(() =>
		normalizeQuery({ categories: this.filters.categories, kinds: this.filters.kinds })
	);
	private requestedQueryKey = $derived(queryKey(this.requestedQuery));
	published = $derived(this.graphQuery.published);
	graph = $derived(this.published?.graph ?? emptyGraph);
	loaded = $derived(this.graphQuery.loaded);
	hasArchitecturalEntities = $derived(
		this.graph.entities.some((entity) => isArchitectureCategory(entity.category))
	);
	selection = $state<MapSelection>();
	positions = $state.raw<Partial<Record<string, Point>>>({});
	inspectorOpen = $state(true);

	availableKinds = $derived.by(() => {
		const kinds = new Set(this.filters.kinds);
		for (const entity of this.published?.graph.entities ?? []) kinds.add(entity.kind);
		return [...kinds].sort();
	});

	hasFilters = $derived(this.filters.categories.length > 0 || this.filters.kinds.length > 0);
	showingPriorQuery = $derived(
		this.published !== undefined && queryKey(this.published.query) !== this.requestedQueryKey
	);
	priorQueryLabel = $derived(this.published ? queryDescription(this.published.query) : "previous filters");
	requestedQueryLabel = $derived(queryDescription(this.requestedQuery));
	partialDescription = $derived.by(() => {
		const graph = this.published?.graph;
		if (!graph || !this.graphQuery.partial) return undefined;

		const details = [
			graph.enumeration.stopReason !== "exhausted"
				? stopReasonDescription[graph.enumeration.stopReason]
				: undefined,
			this.graphQuery.unresolvedRelationshipCount > 0
				? `${this.graphQuery.unresolvedRelationshipCount} unresolved relationships point to entities outside this loaded subset.`
				: undefined,
		].filter((detail): detail is string => detail !== undefined);
		return details.join(" ");
	});

	constructor() {
		watch(
			() => this.requestedQueryKey,
			() => {
				void this.graphQuery.load(this.requestedQuery);
			}
		);
		watch(
			() => this.graphQuery.published,
			(published) => this.acceptPublication(published)
		);

		onDestroy(() => this.dispose());
	}

	isCategorySelected(category: EntityCategory): boolean {
		return this.filters.categories.includes(category);
	}

	categoryChanged = (category: EntityCategory, checked: boolean) => {
		const categories = new Set(this.filters.categories);
		if (checked) categories.add(category);
		else categories.delete(category);
		this.params.update({ categories: [...categories].sort() });
	};

	isKindSelected(kind: string): boolean {
		return this.filters.kinds.includes(kind);
	}

	kindChanged = (kind: string, checked: boolean) => {
		const kinds = new Set(this.filters.kinds);
		if (checked) kinds.add(kind);
		else kinds.delete(kind);
		this.params.update({ kinds: [...kinds].sort() });
	};

	clearFilters = () => this.params.update({ categories: [], kinds: [] });
	refresh = () => void this.graphQuery.load(this.requestedQuery);
	toggleInspector = () => (this.inspectorOpen = !this.inspectorOpen);
	setSelection = (selection: MapSelection | undefined) => (this.selection = selection);

	moveNode = (entityId: string, position: Point) => {
		this.positions = { ...this.positions, [entityId]: position };
	};

	private acceptPublication(published: PublishedGraph | undefined): void {
		if (!published) return;

		this.selection = reconcileSelection(published.graph, this.selection);
		const entityIds = new Set(published.graph.entities.map((entity) => entity.id));
		this.positions = Object.fromEntries(
			Object.entries(this.positions).filter(([entityId]) => entityIds.has(entityId))
		);
	}

	dispose(): void {
		this.graphQuery.dispose();
		this.params.cleanup();
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const initSystemMapViewController = () => ctx.set(new SystemMapViewController());
export const useSystemMapViewController = () => ctx.get();
