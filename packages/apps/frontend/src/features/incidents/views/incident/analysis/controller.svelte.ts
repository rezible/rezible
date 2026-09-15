import type { ComponentProps } from "svelte";
import { Context } from "runed";
import { page } from "$app/state";
import { goto } from "$app/navigation";
import { useSystemAnalysisController } from "$components/system-analysis";
import type {
	GraphInteraction,
	GraphSelection,
	GraphHighlights,
} from "$components/system-analysis/diagram/diagramController.svelte";
import { initEventDialog } from "./incident-timeline/event-dialog/controller.svelte";
import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";

type ContextMenuProps = { timeline?: ComponentProps<typeof IncidentTimelineContextMenu> };
type Selection = GraphSelection & { entryId?: string };

export class IncidentAnalysisController {
	analysis = useSystemAnalysisController();

	entryEditor = initEventDialog(() => this.analysis.refreshEntries());
	
	contextMenu = $state.raw<ContextMenuProps>({});
	
	selectedEntryId = $derived(page.url.searchParams.get("entry"));
	selectedNodeId = $derived(page.url.searchParams.get("node"));
	selectedEdgeId = $derived(page.url.searchParams.get("edge"));
	selectedEntry = $derived(this.analysis.entries.find((entry) => entry.id === this.selectedEntryId));
	selectedNode = $derived(this.analysis.analysisNodes.find((node) => node.id === this.selectedNodeId));
	selectedEdge = $derived(this.analysis.analysisEdges.find((edge) => edge.id === this.selectedEdgeId));
	
	attachedEntries = $derived(
		this.selectedNodeId
			? (this.analysis.attachments.byNodeId.get(this.selectedNodeId) ?? [])
			: this.selectedEdgeId
				? (this.analysis.attachments.byEdgeId.get(this.selectedEdgeId) ?? [])
				: []
	);

	relationships = $derived(
		this.selectedNode
			? this.analysis.analysisEdges.filter((edge) => {
					const relationship = edge.attributes.knowledgeRelationship.attributes;
					return (
						relationship.sourceEntityId === this.selectedNode?.attributes.knowledgeEntity.id ||
						relationship.targetEntityId === this.selectedNode?.attributes.knowledgeEntity.id
					);
				})
			: []
	);

	subjects = $derived(
		(this.selectedEntry?.attributes.subjects ?? []).map((subject) => {
			const node = this.analysis.analysisNodes.find(
				(node) => node.attributes.knowledgeEntity.id === subject.attributes.knowledgeEntityId
			);
			const edge = this.analysis.analysisEdges.find(
				(edge) =>
					edge.attributes.knowledgeRelationship.id === subject.attributes.knowledgeRelationshipId
			);
			return { subject, node, edge };
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

	graphInteraction = $derived.by<GraphInteraction>(() => ({
		selection: { nodeId: this.selectedNodeId ?? undefined, edgeId: this.selectedEdgeId ?? undefined },
		highlights: this.graphHighlights,
		onReady: (focus) => {
			this.focusGraph = focus;
		},
		select: (selection, trigger) => this.updateSelection(selection, trigger),
	}));

	private focusGraph?: (subjects: GraphHighlights) => Promise<void>;

	selectionUrl(selection: Selection) {
		const url = new URL(page.url);
		for (const key of ["entry", "node", "edge"] as const) {
			const id = selection[`${key}Id`];
			if (id) url.searchParams.set(key, id);
			else url.searchParams.delete(key);
		}
		return `${url.pathname}${url.search}${url.hash}`;
	}

	private async updateSelection(selection: Selection, trigger?: HTMLElement) {
		if (trigger) this.returnFocus = trigger;
		this.dismissedSelection = undefined;
		const url = this.selectionUrl(selection);
		if (url !== `${page.url.pathname}${page.url.search}${page.url.hash}`)
			await goto(url, { noScroll: true, keepFocus: true, state: page.state });
	}

	select = async (selection: Selection, trigger?: HTMLElement) => {
		await this.updateSelection(selection, trigger);
		await this.focusGraph?.(this.graphHighlights);
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
