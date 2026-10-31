import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { onDestroy, tick } from "svelte";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { z } from "zod";

import type { SystemMapHandle } from "$features/system/components/system-map";
import { knowledgeModelOptions } from "$features/system/lib/knowledge-model";
import { getMapCategoryDisplay } from "$features/system/lib/system-map/category";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import { emptyMapModel, mapCategories, mapModel } from "$features/system/lib/system-map/model";
import type { MapSelection } from "$features/system/lib/system-map/presentation";
import { reconcileSelection } from "$features/system/lib/system-map/selection";
import { GraphQueryController, type PublishedGraph, type StandaloneGraphQuery } from "./loader.svelte";

const filterSchema = z.object({
	categories: z.array(z.string()).default([]).catch([]),
	kinds: z.array(z.string()).default([]).catch([]),
	/** Structure is the default and leaves the URL; the back button returns to it from deployment. */
	view: z.enum(["structure", "deployment"]).default("structure").catch("structure"),
});

export type MapViewKind = z.infer<typeof filterSchema>["view"];

const emptyGraph: GraphSubset = {
	model: emptyMapModel,
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
		: "structure and deployment categories";
	const kinds = query.kinds.length ? query.kinds.join(", ") : "all kinds";
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
	private modelQuery = createQuery(() => knowledgeModelOptions());
	private params = useSearchParams(filterSchema, { noScroll: true });

	filters = $derived(filterSchema.parse(this.params));
	requestedQuery = $derived.by(() =>
		normalizeQuery({ categories: this.filters.categories, kinds: this.filters.kinds })
	);
	private requestedQueryKey = $derived(queryKey(this.requestedQuery));
	published = $derived(this.graphQuery.published);
	graph = $derived(this.published?.graph ?? emptyGraph);
	loaded = $derived(this.graphQuery.loaded);
	hasArchitecturalEntities = $derived.by(() => {
		const categories = new Set(mapCategories(this.graph.model));
		return this.graph.entities.some((entity) => categories.has(entity.category));
	});
	/** Every category the model knows, for the filter. */
	categoryOptions = $derived(
		this.modelQuery.data ? [...mapModel(this.modelQuery.data.data).categories.keys()] : []
	);
	selection = $state<MapSelection>();
	inspectorOpen = $state(true);
	mapView = $derived(this.filters.view);
	/** The placement target the deployment view opens at, from Explore deployment. */
	deploymentFocusTargetId = $state<string>();
	mapHandle = $state<SystemMapHandle>();

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

	isCategorySelected(category: string): boolean {
		return this.filters.categories.includes(category);
	}

	categoryChanged = (category: string, checked: boolean) => {
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

	setMapView = (view: MapViewKind) => {
		this.deploymentFocusTargetId = undefined;
		this.params.update({ view });
	};

	exploreDeployment = (targetId: string) => {
		this.deploymentFocusTargetId = targetId;
		this.params.update({ view: "deployment" });
	};

	/** Shows the structure map, then moves its viewport to the entity once the map is visible again. */
	showInStructure = async (entityId: string) => {
		this.setMapView("structure");
		// The hidden map has no size until the frame after it is shown measures it.
		await tick();
		await new Promise((resolve) => requestAnimationFrame(resolve));
		await new Promise((resolve) => requestAnimationFrame(resolve));
		this.mapHandle?.reveal({ kind: "entity", entityId });
	};

	private acceptPublication(published: PublishedGraph | undefined): void {
		if (!published) return;

		this.selection = reconcileSelection(published.graph, this.selection);
	}

	dispose(): void {
		this.graphQuery.dispose();
		this.params.cleanup();
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const initSystemMapViewController = () => ctx.set(new SystemMapViewController());
export const useSystemMapViewController = () => ctx.get();
