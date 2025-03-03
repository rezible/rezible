import {
	createIncidentEventMutation,
	updateIncidentEventMutation,
	type CreateIncidentEventAttributes,
	type Incident,
	type IncidentEvent,
	type UpdateIncidentEventAttributes,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { incidentCtx } from "$features/incidents/lib/context";
import { eventAttributes } from "./attribute-panels/eventAttributes.svelte";

type EditorDialogView = "closed" | "create" | "edit";

const createEventDialogState = () => {
	let incident = $state<Incident>();
	let editingEvent = $state<IncidentEvent>();
	let view = $state<EditorDialogView>("closed");
	let previousView = $state<EditorDialogView>("closed");

	const setView = (v: EditorDialogView) => {
		previousView = $state.snapshot(view);
		view = v;
	}

	const clear = () => {
		setView("closed");
	};

	const onSuccess = ({ data: event }: { data: IncidentEvent }) => {
		console.log("event!", event);
		clear();
	}

	const makeCreateMutation = () => createMutation(() => ({ ...createIncidentEventMutation(), onSuccess }));
	const makeUpdateMutation = () => createMutation(() => ({ ...updateIncidentEventMutation(), onSuccess }));

	let createMut = $state<ReturnType<typeof makeCreateMutation>>();
	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();
	const loading = $derived(updateMut?.isPending || createMut?.isPending);

	const doCreate = () => {
		if (!incident || !createMut) return;
		const attrs = eventAttributes.snapshot();
		const path = { id: $state.snapshot(incident.id) };
		const attributes: CreateIncidentEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			isKey: attrs.isKey,
			title: attrs.title,
		};
		createMut.mutate({ path, body: { attributes } });
	}

	const doEdit = () => {
		if (!editingEvent || !updateMut) return;
		const attrs = eventAttributes.snapshot();
		const path = { id: $state.snapshot(editingEvent.id) };
		const attributes: UpdateIncidentEventAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			title: attrs.title,
		};
		updateMut.mutate({ path, body: { attributes } });
	}

	const setup = () => {
		incident = incidentCtx.get();
		updateMut = makeUpdateMutation();
		createMut = makeCreateMutation();
	};

	const setCreating = () => {
		setView("create");
		eventAttributes.init(incident);
	}

	const setEditing = (ev: IncidentEvent) => {
		setView("edit");
		editingEvent = $state.snapshot(ev);
		eventAttributes.init(incident, ev.attributes);
	};

	const confirm = () => {
		if (view === "create") {
			doCreate();
		} else if (view === "edit") {
			doEdit();
		} else {
			console.error("something went wrong", $state.snapshot(view), !!createMut, !!updateMut);
		}
	};

	return {
		setup,
		get view() {
			return view;
		},
		get previousView() {
			return previousView;
		},
		get open() {
			return view !== "closed";
		},
		setCreating,
		setEditing,
		clear,
		confirm,
		get loading() {
			return loading;
		},
		get editingEvent() {
			return editingEvent;
		},
	};
};

export const eventDialog = createEventDialogState();