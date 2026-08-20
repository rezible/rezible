import {
	addSystemAnalysisEntrySubjectMutation,
	createSystemAnalysisEntryMutation,
	deleteSystemAnalysisEntrySubjectMutation,
	updateSystemAnalysisEntryMutation,
	updateSystemAnalysisEntrySubjectMutation,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context } from "runed";

import { useIncidentView } from "$features/incidents/views/incident";
import { useIncidentAnalysis } from "$features/incidents/views/incident/analysis/controller.svelte";
import {
	initEventDialogAttributes,
	type TimelineEventDialogAttributes,
} from "./attribute-panels/attributes.svelte";
import {
	timelineEntryProperties,
	type TimelineAnalysisEntry,
	type TimelineAnalysisEntryAttributes,
	type TimelineEntrySystemContext,
} from "../entry-model";

type EditorDialogView = "closed" | "create" | "edit";
export type OnEventChangedCallbackFn = () => void;

export class IncidentEventDialogController {
	incidentViewController = useIncidentView();
	analysisController = useIncidentAnalysis();
	incident = $derived(this.incidentViewController.incident);
	analysisId = $derived(this.analysisController.analysisId);

	editingEntry = $state<TimelineAnalysisEntry>();
	onEventChangedCallback: OnEventChangedCallbackFn;

	view = $state<EditorDialogView>("closed");
	previousView = $state<EditorDialogView>("closed");

	open = $derived(this.view !== "closed");

	attributes: TimelineEventDialogAttributes;

	constructor(onEventChanged: OnEventChangedCallbackFn) {
		this.onEventChangedCallback = onEventChanged;
		this.attributes = initEventDialogAttributes();
	}

	setView(v: EditorDialogView) {
		this.previousView = $state.snapshot(this.view);
		this.view = v;
	}

	clear() {
		this.setView("closed");
		this.editingEntry = undefined;
	}

	onSuccess() {
		this.onEventChangedCallback();
		this.clear();
	}

	createEntryMut = createMutation(() => createSystemAnalysisEntryMutation());
	updateEntryMut = createMutation(() => updateSystemAnalysisEntryMutation());
	addSubjectMut = createMutation(() => addSystemAnalysisEntrySubjectMutation());
	updateSubjectMut = createMutation(() => updateSystemAnalysisEntrySubjectMutation());
	deleteSubjectMut = createMutation(() => deleteSystemAnalysisEntrySubjectMutation());

	loading = $derived(
		this.createEntryMut.isPending ||
			this.updateEntryMut.isPending ||
			this.addSubjectMut.isPending ||
			this.updateSubjectMut.isPending ||
			this.deleteSubjectMut.isPending
	);

	private makeCreateAttributes(attrs: TimelineAnalysisEntryAttributes) {
		return {
			kind: attrs.kind,
			occurredAt: attrs.timestamp,
			title: attrs.title,
			body: attrs.description,
			properties: timelineEntryProperties(attrs),
		};
	}

	private makeUpdateAttributes(attrs: TimelineAnalysisEntryAttributes) {
		return {
			kind: attrs.kind,
			occurredAt: attrs.timestamp,
			title: attrs.title,
			body: attrs.description,
			properties: timelineEntryProperties(attrs),
		};
	}

	private async addSubject(entryId: string, context: TimelineEntrySystemContext) {
		await this.addSubjectMut.mutateAsync({
			path: { id: entryId },
			body: {
				attributes: {
					role: context.attributes.relationship || "related",
					knowledgeEntityId: context.attributes.knowledgeEntityId,
				},
			},
		});
	}

	private async syncSubjects(
		entryId: string,
		previous: TimelineEntrySystemContext[],
		next: TimelineEntrySystemContext[]
	) {
		const retainedSubjectIds = new Set<string>();
		const previousById = new Map(
			previous.filter((context) => !!context.id).map((context) => [context.id!, context])
		);

		for (const context of next) {
			const previousContext = context.id ? previousById.get(context.id) : undefined;
			if (
				!previousContext ||
				previousContext.attributes.knowledgeEntityId !== context.attributes.knowledgeEntityId
			) {
				await this.addSubject(entryId, context);
				continue;
			}

			retainedSubjectIds.add(previousContext.id!);
			if (previousContext.attributes.relationship !== context.attributes.relationship) {
				await this.updateSubjectMut.mutateAsync({
					path: { id: previousContext.id! },
					body: { attributes: { role: context.attributes.relationship || "related" } },
				});
			}
		}

		for (const previousContext of previous) {
			if (!previousContext.id || retainedSubjectIds.has(previousContext.id)) continue;
			await this.deleteSubjectMut.mutateAsync({ path: { id: previousContext.id } });
		}
	}

	async doCreate() {
		if (!this.incident || !this.analysisId) return;
		const attrs = this.attributes.snapshot();
		const created = await this.createEntryMut.mutateAsync({
			path: { id: this.analysisId },
			body: { attributes: this.makeCreateAttributes(attrs) },
		});
		await this.syncSubjects(created.data.id, [], attrs.systemContext);
		this.onSuccess();
	}

	async doEdit() {
		if (!this.editingEntry) return;
		const attrs = this.attributes.snapshot();
		const entryId = $state.snapshot(this.editingEntry.id);
		await this.updateEntryMut.mutateAsync({
			path: { id: entryId },
			body: { attributes: this.makeUpdateAttributes(attrs) },
		});
		await this.syncSubjects(entryId, this.editingEntry.attributes.systemContext, attrs.systemContext);
		this.onSuccess();
	}

	setCreating(attrs?: Partial<TimelineAnalysisEntryAttributes>) {
		this.setView("create");
		this.attributes.init(this.incident, attrs);
	}

	setEditing(ev: TimelineAnalysisEntry) {
		this.setView("edit");
		this.editingEntry = $state.snapshot(ev);
		this.attributes.init(this.incident, ev.attributes);
	}

	confirm() {
		if (this.view === "create") {
			this.doCreate();
		} else if (this.view === "edit") {
			this.doEdit();
		} else {
			console.error("something went wrong", $state.snapshot(this.view));
		}
	}
}

const ctx = new Context<IncidentEventDialogController>("IncidentEventDialogController");
export const initEventDialog = (onEventChanged: OnEventChangedCallbackFn) =>
	ctx.set(new IncidentEventDialogController(onEventChanged));
export const useEventDialog = () => ctx.get();
