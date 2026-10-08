import {
	expandKnowledgeGraphRelationshipsOptions,
	selectKnowledgeGraphEntitiesOptions,
	type ApiError,
	type KnowledgeGraphEntitySummary,
} from "$lib/api";
import type { QueryClient } from "@tanstack/svelte-query";
import type { GraphEntity, GraphRelationship, GraphSubset } from "$features/system/lib/system-map/graph";

const ENTITY_PAGE_LIMIT = 100;
const RELATIONSHIP_PAGE_LIMIT = 200;
const MAX_ENTITIES = 500;
const MAX_RELATIONSHIPS = 2_000;
const MAX_REQUESTS = 30;

type EntityCategory = KnowledgeGraphEntitySummary["category"];
type GraphQueryStatus = "idle" | "loading" | "success" | "error";

export type StandaloneGraphQuery = {
	readonly categories: readonly EntityCategory[];
	readonly kinds: readonly string[];
};

export type PublishedGraph = {
	readonly query: StandaloneGraphQuery;
	readonly graph: GraphSubset;
};

type LoadRun = {
	readonly generation: number;
	readonly abortController: AbortController;
	requestCount: number;
	cancelCurrentRequest?: () => Promise<void>;
};

type AbortSignalLink = {
	readonly signal: AbortSignal;
	dispose(): void;
};

function normalizeQuery(query: StandaloneGraphQuery): StandaloneGraphQuery {
	return Object.freeze({
		categories: Object.freeze([...new Set(query.categories)].sort()),
		kinds: Object.freeze([...new Set(query.kinds)].sort()),
	});
}

function hasCursor(cursor: string | undefined): cursor is string {
	return cursor !== undefined && cursor.length > 0;
}

function compareIds(left: { id: string }, right: { id: string }): number {
	return left.id < right.id ? -1 : left.id > right.id ? 1 : 0;
}

function retainEntities(
	incoming: readonly KnowledgeGraphEntitySummary[],
	retained: Map<string, GraphEntity>
): boolean {
	let overflow = false;
	for (const entity of [...incoming].sort(compareIds)) {
		const record = Object.freeze({
			id: entity.id,
			category: entity.category,
			kind: entity.kind,
		});
		if (retained.has(record.id)) {
			retained.set(record.id, record);
		} else if (retained.size < MAX_ENTITIES) {
			retained.set(record.id, record);
		} else {
			overflow = true;
		}
	}
	return overflow;
}

function retainRelationships(
	incoming: readonly {
		id: string;
		sourceId: string;
		targetId: string;
		predicate: string;
	}[],
	retained: Map<string, GraphRelationship>
): boolean {
	let overflow = false;
	for (const relationship of [...incoming].sort(compareIds)) {
		const record = Object.freeze({
			id: relationship.id,
			source: relationship.sourceId,
			target: relationship.targetId,
			predicate: relationship.predicate,
		});
		if (retained.has(record.id)) {
			retained.set(record.id, record);
		} else if (retained.size < MAX_RELATIONSHIPS) {
			retained.set(record.id, record);
		} else {
			overflow = true;
		}
	}
	return overflow;
}

function linkAbortSignals(...signals: AbortSignal[]): AbortSignalLink {
	const controller = new AbortController();
	const abort = () => controller.abort();
	for (const signal of signals) {
		if (signal.aborted) {
			controller.abort();
		} else {
			signal.addEventListener("abort", abort, { once: true });
		}
	}
	return {
		signal: controller.signal,
		dispose() {
			signals.forEach((signal) => signal.removeEventListener("abort", abort));
		},
	};
}

// The API client rejects with an ApiError; anything else thrown while loading is internal.
function toLoaderError(error: unknown): ApiError {
	if (!!error && typeof error === "object" && !(error instanceof Error)) {
		return error as ApiError;
	}
	return { code: "internal" };
}

function makeGraph(
	entitiesById: Map<string, GraphEntity>,
	relationshipsById: Map<string, GraphRelationship>,
	stopReason: GraphSubset["enumeration"]["stopReason"]
): GraphSubset {
	const entities = Object.freeze([...entitiesById.values()].sort(compareIds));
	const entityIds = new Set(entities.map((entity) => entity.id));
	const relationships: GraphRelationship[] = [];
	const unresolvedRelationships: GraphRelationship[] = [];
	for (const relationship of [...relationshipsById.values()].sort(compareIds)) {
		if (entityIds.has(relationship.source) && entityIds.has(relationship.target)) {
			relationships.push(relationship);
		} else {
			unresolvedRelationships.push(relationship);
		}
	}
	return Object.freeze({
		entities,
		relationships: Object.freeze(relationships),
		unresolvedRelationships: Object.freeze(unresolvedRelationships),
		enumeration: Object.freeze({ scope: "query" as const, stopReason }),
	});
}

export class GraphQueryController {
	private readonly queryClient: QueryClient;
	private currentStatus = $state<GraphQueryStatus>("idle");
	private currentError = $state.raw<ApiError>();
	private currentPublished = $state.raw<PublishedGraph>();
	private generation = 0;
	private activeRun: LoadRun | undefined;
	private disposed = false;

	get loaded(): boolean {
		return this.currentPublished !== undefined;
	}

	get entityCount(): number {
		return this.currentPublished?.graph.entities.length ?? 0;
	}

	get relationshipCount(): number {
		return this.currentPublished?.graph.relationships.length ?? 0;
	}

	get unresolvedRelationshipCount(): number {
		return this.currentPublished?.graph.unresolvedRelationships.length ?? 0;
	}

	get partial(): boolean {
		const graph = this.currentPublished?.graph;
		return (
			graph !== undefined &&
			(graph.enumeration.stopReason !== "exhausted" || graph.unresolvedRelationships.length > 0)
		);
	}

	constructor(queryClient: QueryClient) {
		this.queryClient = queryClient;
	}

	get status(): GraphQueryStatus {
		return this.currentStatus;
	}

	get error(): ApiError | undefined {
		return this.currentError;
	}

	get published(): PublishedGraph | undefined {
		return this.currentPublished;
	}

	load(query: StandaloneGraphQuery): Promise<void> {
		if (this.disposed) return Promise.resolve();

		const normalizedQuery = normalizeQuery(query);
		const previousRun = this.activeRun;
		const run: LoadRun = {
			generation: ++this.generation,
			abortController: new AbortController(),
			requestCount: 0,
		};
		this.activeRun = run;
		const previousCancellation = previousRun ? this.cancelRun(previousRun) : Promise.resolve();

		if (this.isCurrent(run)) this.currentStatus = "loading";
		if (this.isCurrent(run)) this.currentError = undefined;
		return this.executeLoad(run, normalizedQuery, previousCancellation);
	}

	dispose(): void {
		if (this.disposed) return;
		this.disposed = true;
		this.generation += 1;
		const run = this.activeRun;
		this.activeRun = undefined;
		if (run) void this.cancelRun(run);
	}

	private isCurrent(run: LoadRun): boolean {
		return !this.disposed && this.activeRun === run && this.generation === run.generation;
	}

	private cancelRun(run: LoadRun): Promise<void> {
		run.abortController.abort();
		return run.cancelCurrentRequest?.().catch(() => undefined) ?? Promise.resolve();
	}

	private async executeLoad(
		run: LoadRun,
		query: StandaloneGraphQuery,
		previousCancellation: Promise<void>
	): Promise<void> {
		try {
			await previousCancellation;
			if (!this.isCurrent(run)) return;

			const entitiesById = new Map<string, GraphEntity>();
			const relationshipsById = new Map<string, GraphRelationship>();
			let entityCursor: string | undefined;
			let entityOverflow = false;
			let entityLimit = false;
			let relationshipLimit = false;
			let requestLimit = false;

			while (this.isCurrent(run)) {
				if (run.requestCount >= MAX_REQUESTS) {
					requestLimit = true;
					break;
				}

				const entityPageResponse = await this.fetchEntityPage(run, query, entityCursor);
				if (!this.isCurrent(run)) return;
				const entityPage = entityPageResponse.data;
				entityOverflow ||= retainEntities(entityPage.entities, entitiesById);
				const nextEntityCursor = hasCursor(entityPage.nextCursor) ? entityPage.nextCursor : undefined;
				let relationshipCursor: string | undefined;

				while (this.isCurrent(run)) {
					if (relationshipsById.size >= MAX_RELATIONSHIPS) {
						relationshipLimit = true;
						if (run.requestCount >= MAX_REQUESTS) requestLimit = true;
						break;
					}
					if (run.requestCount >= MAX_REQUESTS) {
						requestLimit = true;
						break;
					}

					const relationshipPageResponse = await this.fetchRelationshipPage(
						run,
						entityPage.entitySelectionRef,
						relationshipCursor
					);
					if (!this.isCurrent(run)) return;
					const relationshipPage = relationshipPageResponse.data;
					if (retainRelationships(relationshipPage.relationships, relationshipsById)) {
						relationshipLimit = true;
						break;
					}

					if (!hasCursor(relationshipPage.nextCursor)) break;
					relationshipCursor = relationshipPage.nextCursor;
					if (relationshipsById.size >= MAX_RELATIONSHIPS) {
						relationshipLimit = true;
						if (run.requestCount >= MAX_REQUESTS) requestLimit = true;
						break;
					}
					if (run.requestCount >= MAX_REQUESTS) {
						requestLimit = true;
						break;
					}
				}

				if (relationshipLimit || requestLimit) break;
				if (entityOverflow) entityLimit = true;
				if (nextEntityCursor) {
					if (relationshipsById.size >= MAX_RELATIONSHIPS) relationshipLimit = true;
					if (run.requestCount >= MAX_REQUESTS) requestLimit = true;
					if (entityOverflow || entitiesById.size >= MAX_ENTITIES) entityLimit = true;
					if (relationshipLimit || requestLimit || entityLimit) break;
					entityCursor = nextEntityCursor;
					continue;
				}
				if (entityOverflow) entityLimit = true;
				break;
			}

			if (!this.isCurrent(run)) return;
			const stopReason = relationshipLimit
				? "relationship-limit"
				: requestLimit
					? "request-limit"
					: entityLimit
						? "entity-limit"
						: "exhausted";
			const graph = makeGraph(entitiesById, relationshipsById, stopReason);
			if (!this.isCurrent(run)) return;
			this.currentPublished = Object.freeze({ query, graph });
			if (this.isCurrent(run)) this.currentStatus = "success";
		} catch (error) {
			if (!this.isCurrent(run) || run.abortController.signal.aborted) return;
			if (this.isCurrent(run)) this.currentError = toLoaderError(error);
			if (this.isCurrent(run)) this.currentStatus = "error";
		} finally {
			if (this.activeRun === run) this.activeRun = undefined;
		}
	}

	private async fetchEntityPage(run: LoadRun, query: StandaloneGraphQuery, cursor: string | undefined) {
		const options = selectKnowledgeGraphEntitiesOptions({
			query: {
				...(query.categories.length > 0 ? { category: [...query.categories] } : {}),
				...(query.kinds.length > 0 ? { kind: [...query.kinds] } : {}),
				...(hasCursor(cursor) ? { cursor } : {}),
				limit: ENTITY_PAGE_LIMIT,
			},
		});
		const generatedQueryFn = options.queryFn;
		if (typeof generatedQueryFn !== "function")
			throw new Error("Generated entity query has no query function");
		const cancelRequest = () =>
			this.queryClient.cancelQueries({ queryKey: options.queryKey, exact: true }, { silent: true });
		run.cancelCurrentRequest = cancelRequest;
		run.requestCount += 1;
		try {
			return await this.queryClient.fetchQuery({
				...options,
				staleTime: 0,
				retry: false,
				queryFn: async (context) => {
					const linked = linkAbortSignals(context.signal, run.abortController.signal);
					try {
						return await generatedQueryFn({ ...context, signal: linked.signal });
					} finally {
						linked.dispose();
					}
				},
			});
		} finally {
			if (run.cancelCurrentRequest === cancelRequest) run.cancelCurrentRequest = undefined;
		}
	}

	private async fetchRelationshipPage(
		run: LoadRun,
		entitySelectionRef: string,
		cursor: string | undefined
	) {
		const options = expandKnowledgeGraphRelationshipsOptions({
			query: {
				entitySelectionRef,
				...(hasCursor(cursor) ? { cursor } : {}),
				limit: RELATIONSHIP_PAGE_LIMIT,
			},
		});
		const generatedQueryFn = options.queryFn;
		if (typeof generatedQueryFn !== "function") {
			throw new Error("Generated relationship query has no query function");
		}
		const cancelRequest = () =>
			this.queryClient.cancelQueries({ queryKey: options.queryKey, exact: true }, { silent: true });
		run.cancelCurrentRequest = cancelRequest;
		run.requestCount += 1;
		try {
			return await this.queryClient.fetchQuery({
				...options,
				staleTime: 0,
				retry: false,
				queryFn: async (context) => {
					const linked = linkAbortSignals(context.signal, run.abortController.signal);
					try {
						return await generatedQueryFn({ ...context, signal: linked.signal });
					} finally {
						linked.dispose();
					}
				},
			});
		} finally {
			if (run.cancelCurrentRequest === cancelRequest) run.cancelCurrentRequest = undefined;
		}
	}
}
