import {
	getKnowledgeGraphViewOptions,
	listKnowledgeGraphEntitiesOptions,
	type ErrorModel,
	type KnowledgeGraphEntity,
	type KnowledgeGraphRelationship,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { MarkerType, type Edge, type Node, type Viewport } from "@xyflow/svelte";
import { Context, watch } from "runed";
import { SvelteMap, SvelteSet } from "svelte/reactivity";

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

export class SystemMapViewController {
	private session = useUserSessionState();
	private entities = new SvelteMap<string, KnowledgeGraphEntity>();
	private relationships = new SvelteMap<string, KnowledgeGraphRelationship>();

	rootId = $state<string>();
	search = $state("");
	searchOpen = $state(false);
	selected = $state<SystemMapSelection>();
	nodes = $state.raw<Node<SystemMapNodeData>[]>([]);
	edges = $state.raw<Edge<SystemMapEdgeData>[]>([]);
	viewport = $state<Viewport>({ x: 80, y: 80, zoom: 0.9 });

	private viewQuery = createQuery(() => ({
		...getKnowledgeGraphViewOptions({
			query: { depth: 1, entityId: this.rootId },
		}),
	}));

	private searchQuery = createQuery(() => ({
		...listKnowledgeGraphEntitiesOptions({
			query: {
				search: this.search.trim(),
				limit: 20,
				offset: 0,
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

	homeEntityId = $derived<string | undefined>(undefined);
	searchResults = $derived(this.searchQuery.data?.data ?? []);
	searching = $derived(this.searchQuery.isLoading || this.searchQuery.isFetching);
	loading = $derived(this.viewQuery.isLoading || this.viewQuery.isFetching);
	error = $derived((this.viewQuery.error ?? this.searchQuery.error) as ErrorModel | undefined);
	hasGraph = $derived(this.nodes.length > 0);

	explore(entity: KnowledgeGraphEntity, reset = false) {
		if (reset) {
			this.entities.clear();
			this.relationships.clear();
		}
		this.entities.set(entity.id, entity);
		this.rootId = entity.id;
		this.searchOpen = false;
		this.search = "";
		this.selected = { kind: "entity", entity };
		this.rebuildGraph(entity.id);
	}

	expand(entity: KnowledgeGraphEntity) {
		this.rootId = entity.id;
		this.selected = { kind: "entity", entity };
		this.focusEntity(entity.id);
	}

	startHere(entity: KnowledgeGraphEntity) {
		this.explore(entity, true);
	}

	reset() {
		const home = this.homeEntityId;
		if (!home) return;
		const entity = this.entities.get(home);
		this.entities.clear();
		this.relationships.clear();
		this.rootId = home;
		this.selected = entity ? { kind: "entity", entity } : undefined;
		this.rebuildGraph(home);
	}

	selectEntity(entity: KnowledgeGraphEntity) {
		this.selected = { kind: "entity", entity };
	}

	selectRelationship(relationship: KnowledgeGraphRelationship) {
		this.selected = { kind: "relationship", relationship };
	}

	clearSelection() {
		this.selected = undefined;
	}

	closeSearch() {
		this.searchOpen = false;
	}

	private mergeView(
		entities: KnowledgeGraphEntity[],
		relationships: KnowledgeGraphRelationship[],
	) {
		for (const entity of entities) {
			if (this.entities.size >= maxEntities && !this.entities.has(entity.id)) break;
			this.entities.set(entity.id, entity);
		}
		for (const relationship of relationships) {
			if (this.relationships.size >= maxRelationships && !this.relationships.has(relationship.id))
				break;
			if (
				this.entities.has(relationship.attributes.sourceEntityId) &&
				this.entities.has(relationship.attributes.targetEntityId)
			) {
				this.relationships.set(relationship.id, relationship);
			}
		}
		this.rebuildGraph(this.rootId);
	}

	private rebuildGraph(focusId?: string) {
		const distances = new SvelteMap<string, number>();
		if (this.rootId) distances.set(this.rootId, 0);
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

		const byLayer = new SvelteMap<number, KnowledgeGraphEntity[]>();
		for (const entity of this.entities.values()) {
			const layer = distances.get(entity.id) ?? 5;
			const entries = byLayer.get(layer) ?? [];
			entries.push(entity);
			byLayer.set(layer, entries);
		}

		const nodes: Node<SystemMapNodeData>[] = [];
		for (const [layer, entities] of [...byLayer.entries()].sort(([a], [b]) => a - b)) {
			entities.sort((a, b) => this.entityLabel(a).localeCompare(this.entityLabel(b)));
			const height = (entities.length - 1) * 120;
			entities.forEach((entity, index) => {
				nodes.push({
					id: entity.id,
					type: "entity",
					position: { x: layer * 320, y: index * 120 - height / 2 },
					data: { entity },
				});
			});
		}
		this.nodes = nodes;
		this.edges = [...this.relationships.values()].map((relationship) => ({
			id: relationship.id,
			type: "default",
			source: relationship.attributes.sourceEntityId,
			target: relationship.attributes.targetEntityId,
			label:
				relationship.attributes.latestEvidence?.attributes.subjectState.displayName ||
				relationship.attributes.kind.replaceAll("_", " "),
			data: { relationship },
			markerEnd: MarkerType.ArrowClosed,
		}));
		if (focusId) this.focusEntity(focusId);
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

	private entityLabel(entity: KnowledgeGraphEntity) {
		return (
			entity.attributes.latestEvidence?.attributes.subjectState?.displayName ||
			entity.attributes.aliases[0]?.attributes.providerSubjectRef ||
			entity.attributes.kind
		);
	}
}

const ctx = new Context<SystemMapViewController>("SystemMapViewController");
export const initSystemMapViewController = () => ctx.set(new SystemMapViewController());
export const useSystemMapViewController = () => ctx.get();
