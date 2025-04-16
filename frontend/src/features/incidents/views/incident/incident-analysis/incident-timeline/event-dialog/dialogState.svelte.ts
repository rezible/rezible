import {
	createIncidentEventMutation,
	updateIncidentEventMutation,
	type CreateIncidentEventAttributes,
	type Incident,
	type IncidentEvent,
	type UpdateIncidentEventAttributes,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

import { useIncidentViewState } from "../../../viewState.svelte";
import { TimelineState, useIncidentTimeline } from "../timelineState.svelte";
import { eventAttributes, TimelineEventDialogAttributesState } from "./attribute-panels/eventAttributesState.svelte";

type EditorDialogView = "closed" | "create" | "edit";

export class EventDialogState {
	timeline = $state<TimelineState>();
	incident = $state<Incident>();
	editingEvent = $state<IncidentEvent>();
	view = $state<EditorDialogView>("closed");
	previousView = $state<EditorDialogView>("closed");

	open = $derived(this.view !== "closed");

	constructor() {
		this.timeline = useIncidentTimeline();
		const viewState = useIncidentViewState();
		watch(() => viewState.incident, inc => {this.incident = inc});
	}

	setView(v: EditorDialogView) {
		this.previousView = $state.snapshot(this.view);
		this.view = v;
	}

	clear() {
		this.setView("closed");
		this.editingEvent = undefined;
	};

	onSuccess({ data: event }: { data: IncidentEvent }) {
		this.timeline?.onEventAdded(event);
		this.clear();
	}

	createEventMut = createMutation(() => ({ ...createIncidentEventMutation(), onSuccess: this.onSuccess }));
	updateEventMut = createMutation(() => ({ ...updateIncidentEventMutation(), onSuccess: this.onSuccess }));
	
	loading = $derived(this.createEventMut.isPending || this.updateEventMut.isPending);

	doCreate() {
		if (!this.incident) return;
		const attrs = eventAttributes.snapshot();
		const path = { id: $state.snapshot(this.incident.id) };
		const attributes: CreateIncidentEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			isKey: attrs.isKey,
			title: attrs.title,
		};
		this.createEventMut.mutate({ path, body: { attributes } });
	}

	doEdit() {
		if (!this.editingEvent) return;
		const attrs = eventAttributes.snapshot();
		const path = { id: $state.snapshot(this.editingEvent.id) };
		const attributes: UpdateIncidentEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			title: attrs.title,
		};
		this.updateEventMut.mutate({ path, body: { attributes } });
	}

	setCreating() {
		this.setView("create");
		eventAttributes.init(this.incident);
	}

	setEditing(ev: IncidentEvent) {
		this.setView("edit");
		this.editingEvent = $state.snapshot(ev);
		eventAttributes.init(this.incident, ev.attributes);
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

const eventDialogCtx = new Context<EventDialogState>("incidentEventDialog");
export const setEventDialog = (s: EventDialogState) => eventDialogCtx.set(s);
export const useEventDialog = () => eventDialogCtx.get();