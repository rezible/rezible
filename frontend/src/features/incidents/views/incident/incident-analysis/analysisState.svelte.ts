import { createMutation, createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import {
	addSystemAnalysisComponentMutation,
	createSystemAnalysisRelationshipMutation,
	deleteSystemAnalysisComponentMutation,
	deleteSystemAnalysisRelationshipMutation,
	getSystemAnalysisOptions,
	updateSystemAnalysisComponentMutation,
	updateSystemAnalysisRelationshipMutation,
	type AddSystemAnalysisComponentAttributes,
	type CreateSystemAnalysisRelationshipAttributes,
	type UpdateSystemAnalysisComponentAttributes,
	type UpdateSystemAnalysisRelationshipAttributes,
} from "$lib/api";
import { Context } from "runed";
import { useIncidentViewController } from "../controller.svelte";
import type { ComponentProps } from "svelte";

import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";
import SystemDiagramContextMenu from "./system-diagram/SystemDiagramContextMenu.svelte";

type ContextMenuProps = {
	timeline?: ComponentProps<typeof IncidentTimelineContextMenu>;
	diagram?: ComponentProps<typeof SystemDiagramContextMenu>;
}

export class IncidentAnalysisState {
	view = useIncidentViewController();
	analysisId = $derived(this.view.systemAnalysisId || "");

	contextMenu = $state<ContextMenuProps>({});

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

	private addAnalysisComponentMutation = createMutation(() => ({
		...addSystemAnalysisComponentMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));

	addComponent(attributes: AddSystemAnalysisComponentAttributes) {
		return this.addAnalysisComponentMutation.mutateAsync({
			path: { id: this.analysisId },
			body: { attributes }
		});
	}

	private updateAnalysisComponentMut = createMutation(() => ({
		...updateSystemAnalysisComponentMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));

	updateComponent(id: string, attributes: UpdateSystemAnalysisComponentAttributes) {
		return this.updateAnalysisComponentMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeAnalysisComponentMut = createMutation(() => ({
		...deleteSystemAnalysisComponentMutation(),
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		},
	}));
	async removeComponent(id: string) {
		return this.removeAnalysisComponentMut.mutate({ path: { id } })
	}

	private createRelationshipMut = createMutation(() => ({ 
		...createSystemAnalysisRelationshipMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async createRelationship(attributes: CreateSystemAnalysisRelationshipAttributes) {
		return this.createRelationshipMut.mutate({ path: { id: this.analysisId }, body: { attributes } });
	}

	private updateRelationshipMut = createMutation(() => ({
		...updateSystemAnalysisRelationshipMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async updateRelationship(id: string, attributes: UpdateSystemAnalysisRelationshipAttributes) {
		return this.updateRelationshipMut.mutate({ path: { id }, body: { attributes } });
	}

	private removeRelationshipMut = createMutation(() => ({
		...deleteSystemAnalysisRelationshipMutation(), 
		onSuccess: () => {
			this.invalidateAnalysisQuery();
		}, 
	}));
	async removeRelationship(id: string) {
		return this.removeRelationshipMut.mutate({ path: { id }});
	}
}

const analysisCtx = new Context<IncidentAnalysisState>("incidentAnalysis");
export const setIncidentAnalysis = (s: IncidentAnalysisState) => analysisCtx.set(s);
export const useIncidentAnalysis = () => analysisCtx.get();
