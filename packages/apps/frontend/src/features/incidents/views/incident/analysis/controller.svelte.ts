import type { ComponentProps } from "svelte";
import { Context } from "runed";
import { createMutation, createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import {
	addSystemAnalysisEdgeMutation,
	addSystemAnalysisNodeMutation,
	deleteSystemAnalysisEdgeMutation,
	deleteSystemAnalysisNodeMutation,
	getSystemAnalysisOptions,
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
}

export class IncidentAnalysisController {
	view = useIncidentView();
	analysisId = $derived(this.view.systemAnalysisId || "");

	queryClient = $state.raw<QueryClient>();

	constructor() {
		this.queryClient = useQueryClient();
	}

	private analysisQueryOptions = $derived(getSystemAnalysisOptions({ path: { id: this.analysisId } }));
	private analysisQuery = createQuery(() => ({
		...this.analysisQueryOptions,
		enabled: !!this.analysisId,
	}));
	analysisData = $derived(this.analysisQuery.data?.data);

	private invalidateAnalysisQuery() {
		this.analysisQuery.refetch();
	}

	private addAnalysisNodeMutation = createMutation(() => ({
		...addSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));

	addNode(attributes: AddSystemAnalysisNodeAttributes) {
		return this.addAnalysisNodeMutation.mutateAsync({
			path: { id: this.analysisId },
			body: { attributes }
		});
	}

	private updateAnalysisNodeMut = createMutation(() => ({
		...updateSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));

	updateNode(id: string, attributes: UpdateSystemAnalysisNodeAttributes) {
		return this.updateAnalysisNodeMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeAnalysisNodeMut = createMutation(() => ({
		...deleteSystemAnalysisNodeMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));
	async removeNode(id: string) {
		return this.removeAnalysisNodeMut.mutate({ path: { id } })
	}

	private addEdgeMut = createMutation(() => ({ 
		...addSystemAnalysisEdgeMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async addEdge(attributes: AddSystemAnalysisEdgeAttributes) {
		return this.addEdgeMut.mutate({ path: { id: this.analysisId }, body: { attributes } });
	}

	private updateEdgeMut = createMutation(() => ({
		...updateSystemAnalysisEdgeMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async updateEdge(id: string, attributes: UpdateSystemAnalysisEdgeAttributes) {
		return this.updateEdgeMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeEdgeMut = createMutation(() => ({
		...deleteSystemAnalysisEdgeMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async removeEdge(id: string) {
		return this.removeEdgeMut.mutate({ path: { id }});
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
