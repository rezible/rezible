import type { IncidentEvent } from "$lib/api";
import { eventAttributes } from "./attribute-panels/eventAttributes.svelte";

type EditorDialogView = "closed" | "create" | "edit";

const createEventEditorDialogState = () => {
	let editingEvent = $state<any>();
	let view = $state<EditorDialogView>("closed");
	let previousView = $state<EditorDialogView>("closed");

	const setView = (v: EditorDialogView) => {
		previousView = $state.snapshot(view);
		view = v;
	}

	const clear = () => {
		setView("closed");
	};

	/*
	const makeUpdateMutation = () => createMutation(() => ({
		...updateSystemComponentMutation(),
		onSuccess: clear,
	}));
	const makeCreateMutation = () => createMutation(() => ({
		...createSystemComponentMutation(), 
		onSuccess: (body: CreateSystemComponentResponseBody) => {
			if (view === "create" && previousView === "add") {
				goBack();
				selectedAddComponent = body.data;
			}
		}
	}));

	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();
	let createMut = $state<ReturnType<typeof makeCreateMutation>>();

	*/
	const saveEnabled = $derived(true);
	const loading = $derived(false);//updateMut?.isPending || createMut?.isPending);

	const setup = () => {
		// updateMut = makeUpdateMutation();
		// createMut = makeCreateMutation();
	};

	const setCreating = () => {
		setView("create");
		eventAttributes.init();
	}

	const setEditing = (ev: IncidentEvent) => {
		setView("edit");
		eventAttributes.init(ev.attributes);
	};

	const confirm = () => {
		clear();
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
		get saveEnabled() {
			return saveEnabled;
		},
		get editingEvent() {
			return editingEvent;
		},
	};
};

export const eventDialog = createEventEditorDialogState();