import type { ComponentProps } from "svelte";
import { Context } from "runed";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import {
	addSystemAnalysisEdgeMutation,
	addSystemAnalysisNodeMutation,
	deleteSystemAnalysisEdgeMutation,
	deleteSystemAnalysisNodeMutation,
	listSystemAnalysisEdgesOptions,
	listSystemAnalysisNodesOptions,
	updateSystemAnalysisEdgeMutation,
	updateSystemAnalysisNodeMutation,
	type AddSystemAnalysisEdgeAttributes,
	type AddSystemAnalysisNodeAttributes,
	type UpdateSystemAnalysisEdgeAttributes,
	type UpdateSystemAnalysisNodeAttributes,
} from "$lib/api";
import { useIncidentView } from "$features/incidents/views/incident/controller.svelte";

import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";
import SystemDiagramContextMenu from "./system-diagram/SystemDiagramContextMenu.svelte";

type ContextMenuProps = {
	timeline?: ComponentProps<typeof IncidentTimelineContextMenu>;
	diagram?: ComponentProps<typeof SystemDiagramContextMenu>;
};

export class IncidentAnalysisController {
	view = useIncidentView();
	analysisId = $derived(this.view.systemAnalysisId || "");

	private analysisNodesQuery = createQuery(() => ({
		...listSystemAnalysisNodesOptions({ path: { id: this.analysisId } }),
		enabled: !!this.analysisId,
	}));
	analysisNodes = $derived(this.analysisNodesQuery.data?.data ?? []);

	private analysisEdgesQuery = createQuery(() => ({
		...listSystemAnalysisEdgesOptions({ path: { id: this.analysisId } }),
		enabled: !!this.analysisId,
	}));
	analysisEdges = $derived(this.analysisEdgesQuery.data?.data ?? []);

	private refetchGraph() {
		this.analysisNodesQuery.refetch();
		this.analysisEdgesQuery.refetch();
	}

	private addAnalysisNodeMutation = createMutation(() => ({
		...addSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));

	addNode(attributes: AddSystemAnalysisNodeAttributes) {
		return this.addAnalysisNodeMutation.mutateAsync({
			path: { id: this.analysisId },
			body: { attributes },
		});
	}

	private updateAnalysisNodeMut = createMutation(() => ({
		...updateSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));

	updateNode(id: string, attributes: UpdateSystemAnalysisNodeAttributes) {
		return this.updateAnalysisNodeMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeAnalysisNodeMut = createMutation(() => ({
		...deleteSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));
	async removeNode(id: string) {
		return this.removeAnalysisNodeMut.mutate({ path: { id } });
	}

	private addEdgeMut = createMutation(() => ({
		...addSystemAnalysisEdgeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));
	async addEdge(attributes: AddSystemAnalysisEdgeAttributes) {
		return this.addEdgeMut.mutate({ path: { id: this.analysisId }, body: { attributes } });
	}

	private updateEdgeMut = createMutation(() => ({
		...updateSystemAnalysisEdgeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));
	async updateEdge(id: string, attributes: UpdateSystemAnalysisEdgeAttributes) {
		return this.updateEdgeMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeEdgeMut = createMutation(() => ({
		...deleteSystemAnalysisEdgeMutation(),
		onSuccess: () => {
			this.refetchGraph();
		},
	}));
	async removeEdge(id: string) {
		return this.removeEdgeMut.mutate({ path: { id } });
	}

	contextMenu = $state.raw<ContextMenuProps>({});

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
