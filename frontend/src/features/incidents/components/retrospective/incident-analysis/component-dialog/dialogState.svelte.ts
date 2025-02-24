import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemComponentMutation,
	updateSystemComponentMutation,
	type CreateSystemComponentAttributes,
	type CreateSystemComponentResponseBody,
	type SystemAnalysisComponent,
	type SystemComponent,
	type SystemComponentAttributes,
	type SystemComponentConstraint,
	type SystemComponentControl,
	type SystemComponentKind,
	type SystemComponentSignal,
	type UpdateSystemComponentAttributes,
} from "$lib/api";
import { analysis } from "$features/incidents/components/retrospective/incident-analysis/analysis.svelte";

const defaultComponentKind = {
	id: "",
	attributes: {
		description: "",
		label: ""
	},
};

const defaultAttributes: SystemComponentAttributes = {
	constraints: [],
	controls: [],
	description: "",
	kind: defaultComponentKind,
	name: "",
	properties: {},
	signals: []
}

const createComponentAttributesState = () => {
	let name = $state<SystemComponentAttributes["name"]>("");
	let kind = $state<SystemComponentKind>(defaultComponentKind);
	let description = $state<SystemComponentAttributes["description"]>("");
	let constraints = $state<SystemComponentAttributes["constraints"]>([]);
	let controls = $state<SystemComponentAttributes["controls"]>([]);
	let signals = $state<SystemComponentAttributes["signals"]>([]);
	let properties = $state<SystemComponentAttributes["properties"]>({});
	
	let valid = $state(false);

	const init = (c?: SystemComponent) => {
		const a = c ? c.attributes : defaultAttributes;
		name = a.name;
		kind = a.kind;
		description = a.description;
		constraints = a.constraints;
		controls = a.controls;
		signals = a.signals;
		properties = a.properties;
		valid = true;
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = !!name && !!kind;
	}

	const updateConstraint = (c: SystemComponentConstraint) => {
		const idx = constraints.findIndex(v => v.id === c.id);
		if (idx >= 0) { constraints[idx] = c }
		else { constraints.push(c) }
		onUpdate();
	}

	const updateSignal = (s: SystemComponentSignal) => {
		const idx = constraints.findIndex(v => v.id === s.id);
		if (idx >= 0) { signals[idx] = s }
		else { signals.push(s) }
		onUpdate();
	}

	const updateControl = (c: SystemComponentControl) => {
		const idx = controls.findIndex(v => v.id === c.id);
		if (idx >= 0) { controls[idx] = c }
		else { controls.push(c) }
		onUpdate();
	}

	// this is gross but oh well
	return {
		init,
		get name() { return name },
		set name(n: string) { name = n; onUpdate(); },
		get kind() { return kind },
		setKind(k?: SystemComponentKind) { kind = (k ?? defaultComponentKind); onUpdate(); },
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get constraints() { return constraints },
		updateConstraint,
		get controls() { return controls },
		updateControl,
		get signals() { return signals },
		updateSignal,
		asAttributes(): SystemComponentAttributes {
			return {
				name: $state.snapshot(name),
				kind: $state.snapshot(kind),
				description: $state.snapshot(description),
				constraints: $state.snapshot(constraints),
				controls: $state.snapshot(controls),
				signals: $state.snapshot(signals),
				properties: $state.snapshot(properties),
			}
		},
		get valid() { return valid },
	}
}

type ComponentDialogView = "closed" | "add" | "create" | "edit";

const createComponentDialogState = () => {
	let editingComponent = $state<SystemAnalysisComponent>();
	let componentAttributes = createComponentAttributesState();
	let selectedAddComponent = $state<SystemComponent>();

	let view = $state<ComponentDialogView>("closed");
	let previousView = $state<ComponentDialogView>("closed");

	const setView = (v: ComponentDialogView) => {
		previousView = $state.snapshot(view);
		view = v;
	}

	const editValid = $derived(componentAttributes.valid && (view === "create" || view === "edit"));
	const addValid = $derived(!!selectedAddComponent && view === "add");

	const clear = () => {
		setView("closed");
		editingComponent = undefined;
		selectedAddComponent = undefined;
		componentAttributes = createComponentAttributesState();
	};

	const goBack = () => {
		if (view === "create" && previousView === "add") {
			setView("add");
			componentAttributes = createComponentAttributesState();
			return;
		}
		clear();
	}

	const makeCreateMutation = () => createMutation(() => ({
		...createSystemComponentMutation(), 
		onSuccess: (body: CreateSystemComponentResponseBody) => {
			if (view === "create" && previousView === "add") {
				goBack();
				selectedAddComponent = body.data;
				return;
			}
			clear();
		}
	}));
	const makeUpdateMutation = () => createMutation(() => ({
		...updateSystemComponentMutation(),
		onSuccess: clear,
	}));

	let createMut = $state<ReturnType<typeof makeCreateMutation>>();
	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();

	const loading = $derived(createMut?.isPending || updateMut?.isPending);

	const setup = () => {
		createMut = makeCreateMutation();
		updateMut = makeUpdateMutation();
	};

	const setAdding = () => {
		setView("add");
	}

	const setSelectedAddComponent = (c?: SystemComponent) => {
		selectedAddComponent = c;
	}

	const setCreating = () => {
		setView("create");
	}

	const setEditing = (sc: SystemAnalysisComponent) => {
		setView("edit");
		editingComponent = sc;
		componentAttributes.init($state.snapshot(sc.attributes.component));
	};

	const doCreate = () => {
		const attr = componentAttributes.asAttributes();
		const reqAttributes: CreateSystemComponentAttributes = {
			name: attr.name,
		};
		createMut?.mutate({ body: { attributes: reqAttributes } });
	}

	const doEdit = () => {
		if (!editingComponent) return;
		const componentId = editingComponent.attributes.component.id;
		const reqAttributes: UpdateSystemComponentAttributes = {
			
		};
		updateMut?.mutate({
			path: { id: componentId },
			body: { attributes: reqAttributes },
		});
	}

	const onConfirm = () => {
		if (view === "create" && componentAttributes.valid) {
			doCreate();
		} else if (view === "edit" && componentAttributes.valid) {
			doEdit();
		} else if (view === "add" && !!selectedAddComponent) {
			analysis.setAddingComponent($state.snapshot(selectedAddComponent));
			clear();
		} else {
			console.error("invalid state to confirm", $state.snapshot(view));
			clear();
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
		setAdding,
		setSelectedAddComponent,
		get selectedAddComponent() {
			return selectedAddComponent;
		},
		setCreating,
		setEditing,
		get componentAttributes() {
			return componentAttributes
		},
		clear,
		goBack,
		onConfirm,
		get loading() {
			return loading;
		},
		get stateValid() {
			return editValid || addValid;
		},
		get editingComponent() {
			return editingComponent;
		},
	};
};

export const componentDialog = createComponentDialogState();