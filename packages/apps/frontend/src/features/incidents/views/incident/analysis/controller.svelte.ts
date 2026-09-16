import type { ComponentProps } from "svelte";
import type { SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import { Context } from "runed";
import { page } from "$app/state";
import { goto } from "$app/navigation";
import { initSystemAnalysisController, type GraphInteraction } from "$components/system-analysis";
import type { GraphSelection, GraphHighlights } from "$components/system-diagram";
import { useIncidentView } from "../controller.svelte";
import { initEventDialog } from "./incident-timeline/event-dialog/controller.svelte";
import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";

type ContextMenuProps = { timeline?: ComponentProps<typeof IncidentTimelineContextMenu> };
type Selection = GraphSelection & { entryId?: string };

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

export class IncidentAnalysisController {
	private incident = useIncidentView();

	systemAnalysis = initSystemAnalysisController(
		() => this.incident.systemAnalysisId,
		() => ({ interaction: this.graphInteraction })
	);

	entryEditor = initEventDialog(() => this.systemAnalysis.refreshEntries());

	contextMenu = $state.raw<ContextMenuProps>({});

	selectedNodeId = $derived(page.url.searchParams.get("node"));
	selectedNode = $derived(
		this.systemAnalysis.analysisNodes.find((node) => node.id === this.selectedNodeId)
	);

	selectedEdgeId = $derived(page.url.searchParams.get("edge"));
	selectedEdge = $derived(
		this.systemAnalysis.analysisEdges.find((edge) => edge.id === this.selectedEdgeId)
	);

	selectedEntryId = $derived(page.url.searchParams.get("entry"));
	selectedEntry = $derived(this.systemAnalysis.entries.find((entry) => entry.id === this.selectedEntryId));

	attachedEntries = $derived.by(() => {
		if (this.selectedNodeId) {
			return this.systemAnalysis.attachments.byNodeId.get(this.selectedNodeId) ?? [];
		}
		if (this.selectedEdgeId) {
			return this.systemAnalysis.attachments.byEdgeId.get(this.selectedEdgeId) ?? [];
		}
		return [];
	});

	relationships = $derived.by(() => {
		if (!this.selectedNode) {
			return [];
		}
		const entityId = this.selectedNode.attributes.knowledgeEntity.id;
		return this.systemAnalysis.analysisEdges
			.filter((edge) => {
				const attrs = edge.attributes.knowledgeRelationship.attributes;
				return attrs.sourceEntityId === entityId || attrs.targetEntityId === entityId;
			})
			.map((edge) => {
				const attrs = edge.attributes.knowledgeRelationship.attributes;
				const otherId =
					attrs.sourceEntityId === entityId ? attrs.targetEntityId : attrs.sourceEntityId;
				const otherNode = this.systemAnalysis.analysisNodes.find(
					(node) => node.attributes.knowledgeEntity.id === otherId
				);
				const otherTitle = otherNode ? nodeTitle(otherNode) : "Unknown subject";
				const label =
					attrs.sourceEntityId === entityId
						? `${attrs.predicate} → ${otherTitle}`
						: `${otherTitle} → ${attrs.predicate}`;
				return { id: edge.id, label };
			});
	});

	subjects = $derived(
		(this.selectedEntry?.attributes.subjects ?? []).map((subject) => {
			const node = this.systemAnalysis.analysisNodes.find(
				(node) => node.attributes.knowledgeEntity.id === subject.attributes.knowledgeEntityId
			);
			const edge = this.systemAnalysis.analysisEdges.find(
				(edge) =>
					edge.attributes.knowledgeRelationship.id === subject.attributes.knowledgeRelationshipId
			);
			const attrs = subject.attributes;
			const selection = node ? { nodeId: node.id } : edge ? { edgeId: edge.id } : undefined;
			const label = node
				? nodeTitle(node)
				: edge?.attributes.knowledgeRelationship.attributes.predicate;
			return {
				node,
				edge,
				selection,
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
		if (!reference) {
			return;
		}
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
		if (this.selectedNode) {
			const attrs = this.selectedNode.attributes;
			const entityAttrs = attrs.knowledgeEntity.attributes;
			return {
				kind: "subject",
				title: nodeTitle(this.selectedNode),
				context: `${entityAttrs.category} · ${entityAttrs.kind}`,
				description: attrs.descriptionOverride ?? entityAttrs.latestState?.description,
			};
		}
		if (this.selectedEdge) {
			const attrs = this.selectedEdge.attributes.knowledgeRelationship.attributes;
			return {
				kind: "subject",
				title: attrs.latestState?.displayName ?? attrs.predicate,
				context: attrs.predicate,
				description: attrs.latestState?.description,
			};
		}
		return undefined;
	});

	selectedRecordLoading = $derived(
		this.systemAnalysis.entriesQuery.isPending || this.systemAnalysis.graphLoading
	);

	selectionKey = $derived([this.selectedEntryId, this.selectedNodeId, this.selectedEdgeId].join(":"));
	hasSelection = $derived(!!(this.selectedEntryId || this.selectedNodeId || this.selectedEdgeId));
	private dismissedSelection = $state<string>();

	inspectorOpen = $derived(this.hasSelection && this.dismissedSelection !== this.selectionKey);

	private returnFocus?: HTMLElement;
	fallbackFocus = $state.raw<HTMLButtonElement | null>(null);

	graphHighlights = $derived.by<GraphHighlights>(() => {
		if (this.selectedNodeId || this.selectedEdgeId) {
			return {
				nodeIds: this.selectedNodeId ? [this.selectedNodeId] : [],
				edgeIds: this.selectedEdgeId ? [this.selectedEdgeId] : [],
			};
		}
		return {
			nodeIds: this.subjects.flatMap(({ node }) => (node ? [node.id] : [])),
			edgeIds: this.subjects.flatMap(({ edge }) => (edge ? [edge.id] : [])),
		};
	});

	graphInteraction = $derived<GraphInteraction>({
		selection: { nodeId: this.selectedNodeId ?? undefined, edgeId: this.selectedEdgeId ?? undefined },
		highlights: this.graphHighlights,
		select: (selection, trigger) => this.updateSelection(selection, trigger),
	});

	selectionUrl(selection: Selection) {
		const url = new URL(page.url);
		for (const key of ["entry", "node", "edge"] as const) {
			const id = selection[`${key}Id`];
			if (id) {
				url.searchParams.set(key, id);
			} else {
				url.searchParams.delete(key);
			}
		}
		return `${url.pathname}${url.search}${url.hash}`;
	}

	private async updateSelection(selection: Selection, trigger?: HTMLElement) {
		if (trigger) {
			this.returnFocus = trigger;
		}
		this.dismissedSelection = undefined;
		const url = this.selectionUrl(selection);
		if (url !== `${page.url.pathname}${page.url.search}${page.url.hash}`) {
			await goto(url, { noScroll: true, keepFocus: true, state: page.state });
		}
	}

	select = async (selection: Selection, trigger?: HTMLElement) => {
		await this.updateSelection(selection, trigger);
		await this.systemAnalysis.diagram.focus(this.graphHighlights);
	};

	selectInspectorLink = (
		event: MouseEvent & { currentTarget: HTMLAnchorElement },
		selection: Selection
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
		if (this.selectedEntry) {
			this.entryEditor.setEditing(this.selectedEntry);
		}
	};

	dismissInspector = () => {
		this.closeInspector();
		this.restoreFocus();
	};

	setInspectorOpen = (open: boolean) => {
		if (!open) {
			this.closeInspector();
		}
	};

	restoreInspectorFocus = (event: Event) => {
		event.preventDefault();
		this.restoreFocus();
	};

	selectEntry = (id: string, trigger?: HTMLElement) => this.select({ entryId: id }, trigger);

	openInspector = () => {
		this.dismissedSelection = undefined;
	};

	closeInspector = () => {
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
