import { createIncidentEventMutation, updateIncidentEventMutation, type CreateIncidentEventAttributes, type CreateIncidentEventResponseBody, type Incident, type IncidentEvent, type UpdateIncidentEventAttributes } from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { eventAttributes } from "./attribute-panels/eventAttributes.svelte";
import { incidentCtx } from "$src/features/incidents/lib/context";

type EditorDialogView = "closed" | "create" | "edit";

const createEventEditorDialogState = () => {
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

	const onMutationSuccess = ({data: event}: {data: IncidentEvent}) => {
		console.log("event!", event);
		clear();
	}

	const makeCreateMutation = () => createMutation(() => ({
		...createIncidentEventMutation(), 
		onSuccess: onMutationSuccess,
	}));
	const makeUpdateMutation = () => createMutation(() => ({
		...updateIncidentEventMutation(),
		onSuccess: onMutationSuccess,
	}));

	let createMut = $state<ReturnType<typeof makeCreateMutation>>();
	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();

	const loading = $derived(updateMut?.isPending || createMut?.isPending);

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
		const attrs = eventAttributes.snapshot();
		if (view === "create" && createMut && incident?.id) {
			const path = {id: $state.snapshot(incident.id)};
			const attributes: CreateIncidentEventAttributes = {
				kind: attrs.kind,
				timestamp: attrs.timestamp,
				title: attrs.title,
			}
			createMut.mutate({path, body: {attributes}});
		} else if (view === "edit" && updateMut && editingEvent) {
			const path = {id: editingEvent.id};
			const attributes: UpdateIncidentEventAttributes = {
				kind: attrs.kind,
				timestamp: attrs.timestamp,
				title: attrs.title,
			};
			updateMut.mutate({path, body: {attributes}});
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

export const eventDialog = createEventEditorDialogState();