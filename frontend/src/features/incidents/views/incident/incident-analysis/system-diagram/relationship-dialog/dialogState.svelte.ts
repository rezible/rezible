import { createMutation } from "@tanstack/svelte-query";
import { SvelteSet } from "svelte/reactivity";
import {
	createSystemAnalysisRelationshipMutation,
	updateSystemAnalysisRelationshipMutation,
	type CreateSystemAnalysisRelationshipAttributes,
	type SystemAnalysisRelationship,
	type UpdateSystemAnalysisRelationshipAttributes,
} from "$lib/api";
import { useIncidentAnalysis } from "../../analysisState.svelte";
import { Context } from "runed";
import { relationshipAttributes } from "./attributesState.svelte";

// TODO: support this
type RelationshipKind =
	| 'request'    // API/Service requests
	| 'data'       // Data flow
	| 'telemetry'  // Monitoring/metrics
	| 'control';   // Control actions

type RelationshipDialogView = "closed" | "create" | "edit";

export class RelationshipDialogState {
	analysis = useIncidentAnalysis();
	
	view = $state<RelationshipDialogView>("closed");
	relationshipId = $state<string>();

	saveEnabled = $derived(relationshipAttributes.valid && (this.view === "create" || relationshipAttributes.changed));

	setCreating(sourceId: string, targetId: string) {
		this.view = "create";
		this.relationshipId = undefined;
		relationshipAttributes.initNew(sourceId, targetId);
	}

	setEditing(rel: SystemAnalysisRelationship) {
		this.view = "edit";
		this.relationshipId = rel.id;
		relationshipAttributes.initFrom(rel.attributes);
	};

	clear() {
		this.view = "closed";
		relationshipAttributes.initNew("", "");
		this.relationshipId = undefined;
	};

	onSuccess() {
		this.clear();
		alert("invalidate analysis query data");
		// this.analysis.invalidateQueryData();
	}

	createRelationshipMut = createMutation(() => ({ ...createSystemAnalysisRelationshipMutation(), onSuccess: this.onSuccess }))
	updateRelationshipMut = createMutation(() => ({ ...updateSystemAnalysisRelationshipMutation(), onSuccess: this.onSuccess }))

	loading = $derived(this.createRelationshipMut.isPending || this.updateRelationshipMut.isPending);

	doCreate() {
		const analysisId = $state.snapshot(this.analysis.analysisId);
		if (!analysisId) return;
		const attr = relationshipAttributes.snapshot();
		const attributes: CreateSystemAnalysisRelationshipAttributes = {
			sourceId: attr.sourceId,
			targetId: attr.targetId,
			description: attr.description,
			controlActions: attr.controlActions.map(a => a.attributes),
			feedbackSignals: attr.feedbackSignals.map(a => a.attributes),
		};
		this.createRelationshipMut.mutate({ path: { id: analysisId }, body: { attributes } });
	}

	doEdit() {
		if (!this.relationshipId) return;
		const attr = relationshipAttributes.snapshot();
		const attributes: UpdateSystemAnalysisRelationshipAttributes = {
			description: attr.description,
			controlActions: attr.controlActions.map(a => a.attributes),
			feedbackSignals: attr.feedbackSignals.map(s => s.attributes),
		};
		this.updateRelationshipMut.mutate({ path: { id: $state.snapshot(this.relationshipId) }, body: { attributes } });
	}

	onConfirm() {
		if (!relationshipAttributes.valid) return;
		if (this.view == "create") {
			this.doCreate();
		} else if (this.view == "edit") {
			this.doEdit();
		} else {
			console.error("Invalid view state");
		}
	};
};

const relationshipDialogCtx = new Context<RelationshipDialogState>("relationshipDialogState");
export const setRelationshipDialog = (r: RelationshipDialogState) => relationshipDialogCtx.set(r);
export const useRelationshipDialog = () => relationshipDialogCtx.get();;
