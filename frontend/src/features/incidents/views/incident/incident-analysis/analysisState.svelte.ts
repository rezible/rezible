import { createMutation, createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import {
	addSystemAnalysisComponentMutation,
	getSystemAnalysisOptions,
	type SystemAnalysisRelationship,
	type SystemComponent,
} from "$lib/api";
import type { XYPosition } from "@xyflow/svelte";
import { Context, watch } from "runed";

export class IncidentAnalysisState {
	analysisId = $state("");

	queryClient = $state<QueryClient>();

	constructor(idFn: () => (string | undefined)) {
		watch(idFn, id => {this.analysisId = id ?? ""});
		this.queryClient = useQueryClient();
	}

	private analysisQuery = createQuery(() => ({
		...getSystemAnalysisOptions({ path: { id: this.analysisId } }),
		enabled: !!this.analysisId,
	}));
	analysisData = $derived(this.analysisQuery.data?.data);

	private invalidateAnalysisQuery() {
		this.analysisQuery.refetch();
	}

	private addComponentMutation = createMutation(() => ({
		...addSystemAnalysisComponentMutation(),
		onSuccess: this.invalidateAnalysisQuery,
	}));

	async addComponent(component: SystemComponent, pos: XYPosition) {
		if (!this.analysisId) return false;

		try {
			const resp = await this.addComponentMutation.mutateAsync({
				path: { id: this.analysisId },
				body: { 
					attributes: {
						componentId: component.id,
						position: pos,
					}
				}
			});
			return resp.data;
		} catch (e) {
			return false;
		}
	}

	relationshipDialogOpen = $state(false);
	editingRelationship = $state<SystemAnalysisRelationship>();

	setRelationshipDialogOpen(open: boolean, editRel?: SystemAnalysisRelationship) {
		this.relationshipDialogOpen = open;
		this.editingRelationship = editRel;
	};
}

const analysisCtx = new Context<IncidentAnalysisState>("incidentAnalysis");
export const setIncidentAnalysis = (s: IncidentAnalysisState) => analysisCtx.set(s);
export const useIncidentAnalysis = () => analysisCtx.get();
