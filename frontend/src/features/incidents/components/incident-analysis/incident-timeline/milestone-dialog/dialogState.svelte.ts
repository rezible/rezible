import {
	createIncidentMilestoneMutation,
	updateIncidentMilestoneMutation,
	type CreateIncidentMilestoneAttributes,
	type Incident,
	type IncidentMilestone,
	type UpdateIncidentMilestoneAttributes,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { incidentCtx } from "$features/incidents/lib/context";
import { timeline } from "$features/incidents/components/incident-analysis/incident-timeline/timeline.svelte";
import { milestoneAttributes } from "./milestoneAttributes.svelte";

type EditorDialogView = "closed" | "create" | "edit";

const createMilestoneDialogState = () => {
	let incident = $state<Incident>();
	let editingMilestone = $state<IncidentMilestone>();
	let view = $state<EditorDialogView>("closed");
	let previousView = $state<EditorDialogView>("closed");

	const open = $derived(view !== "closed");

	const setView = (v: EditorDialogView) => {
		previousView = $state.snapshot(view);
		view = v;
	}

	const clear = () => {
		setView("closed");
		editingMilestone = undefined;
	};

	const onSuccess = ({ data: milestone }: { data: IncidentMilestone }) => {
		//timeline.eventAdded(event);
		clear();
	}

	const makeCreateMutation = () => createMutation(() => ({ ...createIncidentMilestoneMutation(), onSuccess }));
	const makeUpdateMutation = () => createMutation(() => ({ ...updateIncidentMilestoneMutation(), onSuccess }));

	let createMut = $state<ReturnType<typeof makeCreateMutation>>();
	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();
	const loading = $derived(updateMut?.isPending || createMut?.isPending);

	const doCreate = () => {
		if (!incident || !createMut) return;
		const attrs = milestoneAttributes.snapshot();
		const path = { id: $state.snapshot(incident.id) };
		const attributes: CreateIncidentMilestoneAttributes = {
			kind: attrs.kind,
			timestamp: attrs.timestamp,
			title: attrs.title,
		};
		createMut.mutate({ path, body: { attributes } });
	}

	const doEdit = () => {
		if (!editingMilestone || !updateMut) return;
		const attrs = milestoneAttributes.snapshot();
		const path = { id: $state.snapshot(editingMilestone.id) };
		const attributes: UpdateIncidentMilestoneAttributes = {
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
		milestoneAttributes.init(incident);
	}

	const setEditing = (m: IncidentMilestone) => {
		setView("edit");
		editingMilestone = $state.snapshot(m);
		milestoneAttributes.init(incident, m.attributes);
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
			return open;
		},
		setCreating,
		setEditing,
		clear,
		confirm,
		get loading() {
			return loading;
		},
		get editingMilestone() {
			return editingMilestone;
		},
	};
};

export const milestoneDialog = createMilestoneDialogState();