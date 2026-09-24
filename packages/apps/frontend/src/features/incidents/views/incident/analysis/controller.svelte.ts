import type { ComponentProps } from "svelte";
import type { SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import { Context } from "runed";
import { page } from "$app/state";
import { goto } from "$app/navigation";
import { initSystemAnalysisController, type GraphInteraction } from "$components/system-analysis";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { useIncidentView } from "../controller.svelte";
import { initEventDialog } from "./incident-timeline/event-dialog/controller.svelte";
import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";

type ContextMenuProps = { timeline?: ComponentProps<typeof IncidentTimelineContextMenu> };
type IncidentSelection = { entryId?: string; mapSelection?: MapSelection };

export type InspectorEntry = {
	kind: "entry";
	attrs: SystemAnalysisEntry["attributes"];
	occurredAtLabel: string | undefined;
};

export type InspectorSubject = {
	kind: "subject";
	title: string;
	context: string;
	description: string | undefined;
};

function nodeTitle(node: SystemAnalysisNode) {
	return (
		node.attributes.labelOverride ??
		node.attributes.knowledgeEntity.attributes.latestState?.displayName ??
		"Subject"
	);
}

function mapSelectionFromUrl(): MapSelection | undefined {
	const entityId = page.url.searchParams.get("entity");
	if (entityId) return { kind: "entity", entityId };

	const summaryIds = page.url.searchParams.getAll("summary");
	if (summaryIds.length > 0) return { kind: "summary", relationshipIds: summaryIds };

	const relationshipId = page.url.searchParams.get("relationship");
	return relationshipId ? { kind: "relationship", relationshipId } : undefined;
}

function selectionKey(selection: IncidentSelection): string {
	if (selection.entryId) return `entry:${selection.entryId}`;
	const graphSelection = selection.mapSelection;
	if (!graphSelection) return "";
	if (graphSelection.kind === "entity") return `entity:${graphSelection.entityId}`;
	if (graphSelection.kind === "relationship") return `relationship:${graphSelection.relationshipId}`;
	return `summary:${JSON.stringify(graphSelection.relationshipIds)}`;
}

export class IncidentAnalysisController {
	private incident = useIncidentView();

	systemAnalysis = initSystemAnalysisController(
		() => this.incident.systemAnalysisId,
		() => ({ interaction: this.graphInteraction })
	);

	entryEditor = initEventDialog(() => this.systemAnalysis.refreshEntries());

	contextMenu = $state.raw<ContextMenuProps>({});

	selectedEntityId = $derived(page.url.searchParams.get("entity"));
	selectedEntity = $derived(
		this.selectedEntityId ? this.systemAnalysis.nodeByEntityId.get(this.selectedEntityId) : undefined
	);

	selectedRelationshipId = $derived(page.url.searchParams.get("relationship"));
	selectedEdge = $derived(
		this.selectedRelationshipId
			? this.systemAnalysis.edgeByRelationshipId.get(this.selectedRelationshipId)
			: undefined
	);

	mapSelection = $derived(mapSelectionFromUrl());

	selectedEntryId = $derived(page.url.searchParams.get("entry"));
	selectedEntry = $derived(this.systemAnalysis.entries.find((entry) => entry.id === this.selectedEntryId));

	attachedEntries = $derived.by(() => {
		const selection = this.mapSelection;
		if (!selection) return [];
		if (selection.kind === "entity") {
			return this.systemAnalysis.attachments.byEntityId.get(selection.entityId) ?? [];
		}
		const relationshipIds =
			selection.kind === "relationship" ? [selection.relationshipId] : selection.relationshipIds;
		const entries = new Map<string, SystemAnalysisEntry>();
		for (const relationshipId of relationshipIds) {
			for (const entry of this.systemAnalysis.attachments.byRelationshipId.get(relationshipId) ?? []) {
				entries.set(entry.id, entry);
			}
		}
		return [...entries.values()];
	});

	relationships = $derived.by(() => {
		if (!this.selectedEntity) return [];

		const entityId = this.selectedEntity.attributes.knowledgeEntity.id;
		return [...this.systemAnalysis.edgeByRelationshipId.values()]
			.map((edge) => ({ edge, attrs: edge.attributes.knowledgeRelationship.attributes }))
			.filter(({ attrs }) => attrs.sourceEntityId === entityId || attrs.targetEntityId === entityId)
			.map(({ edge, attrs }) => {
				const otherId =
					attrs.sourceEntityId === entityId ? attrs.targetEntityId : attrs.sourceEntityId;
				const otherNode = this.systemAnalysis.nodeByEntityId.get(otherId);
				const otherTitle = otherNode ? nodeTitle(otherNode) : "Unknown subject";
				const label =
					attrs.sourceEntityId === entityId
						? `${attrs.predicate} → ${otherTitle}`
						: `${otherTitle} → ${attrs.predicate}`;
				return { id: edge.attributes.knowledgeRelationship.id, label };
			});
	});

	subjects = $derived(
		(this.selectedEntry?.attributes.subjects ?? []).map((subject) => {
			const attrs = subject.attributes;
			const node = attrs.knowledgeEntityId
				? this.systemAnalysis.nodeByEntityId.get(attrs.knowledgeEntityId)
				: undefined;
			const edge = attrs.knowledgeRelationshipId
				? this.systemAnalysis.edgeByRelationshipId.get(attrs.knowledgeRelationshipId)
				: undefined;
			const mapSelection = node
				? { kind: "entity" as const, entityId: node.attributes.knowledgeEntity.id }
				: edge
					? {
							kind: "relationship" as const,
							relationshipId: edge.attributes.knowledgeRelationship.id,
						}
					: undefined;
			const label = node
				? nodeTitle(node)
				: (edge?.attributes.labelOverride ??
					edge?.attributes.knowledgeRelationship.attributes.predicate);
			return {
				node,
				edge,
				mapSelection,
				label,
				id: subject.id,
				role: attrs.role,
				unavailable: attrs.knowledgeEvidenceId
					? "Evidence details unavailable"
					: "Subject details unavailable",
				reference:
					attrs.knowledgeEvidenceId ??
					attrs.normalizedEventId ??
					attrs.knowledgeEntityId ??
					attrs.knowledgeRelationshipId,
			};
		})
	);

	sourceUrl = $derived.by(() => {
		const reference = this.selectedEntry?.attributes.reference;
		if (!reference) return;
		try {
			const url = new URL(reference);
			return ["http:", "https:"].includes(url.protocol) ? url.href : undefined;
		} catch {
			return;
		}
	});

	inspectorDetails = $derived.by<InspectorEntry | InspectorSubject | undefined>(() => {
		if (this.selectedEntry) {
			const attrs = this.selectedEntry.attributes;
			return {
				kind: "entry",
				attrs,
				occurredAtLabel: attrs.occurredAt ? new Date(attrs.occurredAt).toLocaleString() : undefined,
			};
		}
		if (this.selectedEntity) {
			const attrs = this.selectedEntity.attributes;
			const entityAttrs = attrs.knowledgeEntity.attributes;
			return {
				kind: "subject",
				title: nodeTitle(this.selectedEntity),
				context: `${entityAttrs.category} · ${entityAttrs.kind}`,
				description: attrs.descriptionOverride ?? entityAttrs.latestState?.description,
			};
		}
		if (this.selectedEdge) {
			const attrs = this.selectedEdge.attributes.knowledgeRelationship.attributes;
			return {
				kind: "subject",
				title:
					this.selectedEdge.attributes.labelOverride ??
					attrs.latestState?.displayName ??
					attrs.predicate,
				context: attrs.predicate,
				description:
					this.selectedEdge.attributes.descriptionOverride ?? attrs.latestState?.description,
			};
		}
		return undefined;
	});

	selectedRecordLoading = $derived(
		this.systemAnalysis.entriesQuery.isPending || this.systemAnalysis.graphLoading
	);

	selectionKey = $derived(
		selectionKey({ entryId: this.selectedEntryId ?? undefined, mapSelection: this.mapSelection })
	);
	hasSelection = $derived(!!(this.selectedEntryId || this.mapSelection));
	private dismissedSelection = $state<string>();
	private inventoryRequested = $state(false);

	inspectorOpen = $derived(
		this.inventoryRequested || (this.hasSelection && this.dismissedSelection !== this.selectionKey)
	);

	private returnFocus?: HTMLElement;
	fallbackFocus = $state.raw<HTMLButtonElement | null>(null);

	graphHighlights = $derived.by<MapHighlights>(() => {
		const subjects = this.selectedEntry?.attributes.subjects ?? [];
		return {
			entityIds: [
				...new Set(
					subjects.flatMap(({ attributes }) =>
						attributes.knowledgeEntityId ? [attributes.knowledgeEntityId] : []
					)
				),
			],
			relationshipIds: [
				...new Set(
					subjects.flatMap(({ attributes }) =>
						attributes.knowledgeRelationshipId ? [attributes.knowledgeRelationshipId] : []
					)
				),
			],
		};
	});

	graphInteraction = $derived<GraphInteraction>({
		selection: this.mapSelection,
		highlights: this.graphHighlights,
		select: (selection, trigger) => this.updateSelection({ mapSelection: selection }, trigger),
	});

	selectionUrl(selection: IncidentSelection) {
		const url = new URL(page.url);
		for (const key of ["entry", "entity", "relationship", "summary", "node", "edge"] as const) {
			url.searchParams.delete(key);
		}
		if (selection.entryId) url.searchParams.set("entry", selection.entryId);

		const graphSelection = selection.mapSelection;
		if (graphSelection?.kind === "entity") {
			url.searchParams.set("entity", graphSelection.entityId);
		} else if (graphSelection?.kind === "relationship") {
			url.searchParams.set("relationship", graphSelection.relationshipId);
		} else if (graphSelection?.kind === "summary") {
			for (const relationshipId of graphSelection.relationshipIds) {
				url.searchParams.append("summary", relationshipId);
			}
		}

		return `${url.pathname}${url.search}${url.hash}`;
	}

	private async updateSelection(selection: IncidentSelection, trigger?: HTMLElement) {
		if (trigger) this.returnFocus = trigger;
		this.dismissedSelection = undefined;
		const url = this.selectionUrl(selection);
		if (url !== `${page.url.pathname}${page.url.search}${page.url.hash}`) {
			await goto(url, { noScroll: true, keepFocus: true, state: page.state });
		}
	}

	select = (selection: IncidentSelection, trigger?: HTMLElement) =>
		this.updateSelection(selection, trigger);

	selectInspectorLink = (
		event: MouseEvent & { currentTarget: HTMLAnchorElement },
		selection: IncidentSelection
	) => {
		if (
			event.defaultPrevented ||
			event.button !== 0 ||
			event.ctrlKey ||
			event.metaKey ||
			event.shiftKey ||
			event.altKey
		) {
			return;
		}
		event.preventDefault();
		void this.select(selection, event.currentTarget);
	};

	editSelectedEntry = () => {
		if (this.selectedEntry) this.entryEditor.setEditing(this.selectedEntry);
	};

	dismissInspector = () => {
		this.closeInspector();
		this.restoreFocus();
	};

	setInspectorOpen = (open: boolean) => {
		if (!open) this.closeInspector();
	};

	restoreInspectorFocus = (event: Event) => {
		event.preventDefault();
		this.restoreFocus();
	};

	selectEntry = (id: string, trigger?: HTMLElement) => this.select({ entryId: id }, trigger);

	openInspector = (event: MouseEvent) => {
		if (event.currentTarget instanceof HTMLElement) this.returnFocus = event.currentTarget;
		this.inventoryRequested = true;
		this.dismissedSelection = undefined;
	};

	closeInspector = () => {
		this.inventoryRequested = false;
		this.dismissedSelection = this.selectionKey;
	};

	restoreFocus = () => {
		const target = this.returnFocus?.isConnected ? this.returnFocus : this.fallbackFocus;
		target?.focus({ preventScroll: true });
	};

	setContextMenu(props: ContextMenuProps) {
		this.contextMenu = props;
	}

	clearContextMenu() {
		this.contextMenu = {};
	}
}

const ctx = new Context<IncidentAnalysisController>("IncidentAnalysisController");
export const initIncidentAnalysisController = () => ctx.set(new IncidentAnalysisController());
export const useIncidentAnalysis = () => ctx.get();
