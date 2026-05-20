import {
	createIncidentTimelineEventMutation,
	updateIncidentTimelineEventMutation,
	type CreateIncidentTimelineEventAttributes,
	type Incident,
	type IncidentTimelineEvent,
	type IncidentTimelineEventAttributes,
	type UpdateIncidentTimelineEventAttributes,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context } from "runed";

import { useIncidentView } from "$features/incidents/views/incident";
import { initEventDialogAttributes, type TimelineEventDialogAttributes } from "./attribute-panels/attributes.svelte";

type EditorDialogView = "closed" | "create" | "edit";
export type OnEventAddedCallbackFn = (e: IncidentTimelineEvent) => void;

export class IncidentEventDialogController {
	incidentViewController = useIncidentView();
	incident = $derived(this.incidentViewController.incident);
	
	editingEvent = $state<IncidentTimelineEvent>();
	onEventAddedCallback: OnEventAddedCallbackFn;

	view = $state<EditorDialogView>("closed");
	previousView = $state<EditorDialogView>("closed");

	open = $derived(this.view !== "closed");

	attributes: TimelineEventDialogAttributes;

	constructor(onEventAdded: OnEventAddedCallbackFn) {
		this.onEventAddedCallback = onEventAdded;
		this.attributes = initEventDialogAttributes();
	}

	setView(v: EditorDialogView) {
		this.previousView = $state.snapshot(this.view);
		this.view = v;
	}

	clear() {
		this.setView("closed");
		this.editingEvent = undefined;
	};

	onSuccess({ data: event }: { data: IncidentTimelineEvent }) {
		this.onEventAddedCallback(event);
		this.clear();
	}

	createEventMut = createMutation(() => ({ ...createIncidentTimelineEventMutation(), onSuccess: e => {this.onSuccess(e)} }));
	updateEventMut = createMutation(() => ({ ...updateIncidentTimelineEventMutation(), onSuccess: e => {this.onSuccess(e)} }));
	
	loading = $derived(this.createEventMut.isPending || this.updateEventMut.isPending);

	doCreate() {
		if (!this.incident) return;
		const attrs = this.attributes.snapshot();
		const path = { id: $state.snapshot(this.incident.id) };
		const attributes: CreateIncidentTimelineEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			isKey: attrs.isKey,
			title: attrs.title,
		};
		this.createEventMut.mutate({ path, body: { attributes } });
	}

	doEdit() {
		if (!this.editingEvent) return;
		const attrs = this.attributes.snapshot();
		const path = { id: $state.snapshot(this.editingEvent.id) };
		const attributes: UpdateIncidentTimelineEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			title: attrs.title,
		};
		this.updateEventMut.mutate({ path, body: { attributes } });
	}

	setCreating(attrs?: Partial<IncidentTimelineEventAttributes>) {
		this.setView("create");
		this.attributes.init(this.incident, attrs);
	}

	setEditing(ev: IncidentTimelineEvent) {
		this.setView("edit");
		this.editingEvent = $state.snapshot(ev);
		this.attributes.init(this.incident, ev.attributes);
	};

	confirm() {
		if (this.view === "create") {
			this.doCreate();
		} else if (this.view === "edit") {
			this.doEdit();
		} else {
			console.error("something went wrong", $state.snapshot(this.view));
		}
	};
}

const ctx = new Context<IncidentEventDialogController>("IncidentEventDialogController");
export const initEventDialog = (onEventAdded: OnEventAddedCallbackFn) => ctx.set(new IncidentEventDialogController(onEventAdded));
export const useEventDialog = () => ctx.get();
