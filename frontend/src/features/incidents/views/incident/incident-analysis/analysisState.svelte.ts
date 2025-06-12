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
	type SystemAnalysisRelationship,
	type UpdateSystemAnalysisComponentAttributes,
	type UpdateSystemAnalysisRelationshipAttributes,
} from "$lib/api";
import { Context, watch } from "runed";

export class IncidentAnalysisState {
	analysisId = $state("");

	queryClient = $state.raw<QueryClient>();

	constructor(idFn: () => (string | undefined)) {
		watch(idFn, id => {this.analysisId = id || ""});
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

	async removeRelationship(c: SystemAnalysisRelationship) {

	}
}

const analysisCtx = new Context<IncidentAnalysisState>("incidentAnalysis");
export const setIncidentAnalysis = (s: IncidentAnalysisState) => analysisCtx.set(s);
export const useIncidentAnalysis = () => analysisCtx.get();
