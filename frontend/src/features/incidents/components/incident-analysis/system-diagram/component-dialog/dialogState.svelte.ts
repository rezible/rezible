import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemComponentConstraintMutation,
	createSystemComponentControlMutation,
	createSystemComponentSignalMutation,
	createSystemComponentMutation,
	updateSystemComponentMutation,
	updateSystemComponentConstraintMutation,
	updateSystemComponentControlMutation,
	updateSystemComponentSignalMutation,
	type SystemComponentConstraint,
	type SystemComponentControl,
	type SystemComponentSignal,
	type CreateSystemComponentAttributes,
	type SystemAnalysisComponent,
	type SystemComponent,
	type SystemComponentAttributes,
	type SystemComponentKind,
	type UpdateSystemComponentAttributes,
} from "$lib/api";
import type { XYPosition } from "@xyflow/svelte";
import { analysis } from "$features/incidents/components/incident-analysis/analysisState.svelte";
import { diagram } from "$features/incidents/components/incident-analysis/system-diagram/diagram.svelte";

const emptyComponentKind = () => ({
	id: "",
	attributes: {
		description: "",
		label: ""
	},
});

const emptyComponentAttributes = (): SystemComponentAttributes => ({
	constraints: [],
	controls: [],
	description: "",
	kind: emptyComponentKind(),
	name: "",
	properties: {},
	signals: []
})

const emptyTrait = () => ({ id: "", attributes: { label: "", description: "" } });

type ComponentDialogView = "closed" | "add" | "create" | "edit";

const createComponentAttributesState = () => {
	let name = $state<SystemComponentAttributes["name"]>("");
	let kind = $state<SystemComponentKind>(emptyComponentKind());
	let description = $state<SystemComponentAttributes["description"]>("");
	let constraints = $state<SystemComponentAttributes["constraints"]>([]);
	let controls = $state<SystemComponentAttributes["controls"]>([]);
	let signals = $state<SystemComponentAttributes["signals"]>([]);
	let properties = $state<SystemComponentAttributes["properties"]>({});

	let valid = $state(false);

	const init = (c?: SystemComponent) => {
		const a = c ? c.attributes : emptyComponentAttributes();
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
		// TODO: actually check if attributes valid;
		valid = !!name && !!kind.id;
	}

	const updateKind = (k?: SystemComponentKind) => {
		kind = (k ?? emptyComponentKind());
		onUpdate(); 
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

	return {
		init,
		get name() { return name },
		set name(n: string) { name = n; onUpdate(); },
		get kind() { return kind },
		updateKind,
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get constraints() { return constraints },
		updateConstraint,
		get controls() { return controls },
		updateControl,
		get signals() { return signals },
		updateSignal,
		snapshot(): SystemComponentAttributes {
			return $state.snapshot({ name, kind, description, constraints, controls, signals, properties })
		},
		get valid() { return valid },
	}
}

export const componentAttributes = createComponentAttributesState();

const createComponentDialogState = () => {
	let editingComponent = $state<SystemAnalysisComponent>();
	let selectedAddComponent = $state<SystemComponent>();
	let addingPosition = $state<XYPosition>();

	let view = $state<ComponentDialogView>("closed");
	let previousView = $state<ComponentDialogView>("closed");

	const creatingToAdd = $derived(view === "create" && previousView === "add");

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
		addingPosition = undefined;
		componentAttributes.init();
	};

	const goBack = () => {
		if (creatingToAdd) {
			setView("add");
			componentAttributes.init();
			return;
		}
		clear();
	}

	const onSuccess = ({ data }: { data: SystemComponent }) => {
		console.log("success", creatingToAdd);
		if (creatingToAdd) {
			goBack();
			selectedAddComponent = data;
			return;
		}
		clear();
	}
	const makeCreateMutation = () => createMutation(() => ({ ...createSystemComponentMutation(), onSuccess }));
	type CreateMutation = ReturnType<typeof makeCreateMutation>;

	const makeUpdateMutation = () => createMutation(() => ({ ...updateSystemComponentMutation(), onSuccess }));
	type UpdateMutation = ReturnType<typeof makeUpdateMutation>;

	let createMut = $state<CreateMutation>();
	let updateMut = $state<UpdateMutation>();

	const loading = $derived(createMut?.isPending || updateMut?.isPending);

	const setup = () => {
		createMut = makeCreateMutation();
		updateMut = makeUpdateMutation();
		componentTraits.setup();
	};

	const setAdding = (pos?: XYPosition) => {
		setView("add");
		addingPosition = pos;
	}

	const setSelectedAddComponent = (c?: SystemComponent) => selectedAddComponent = c;

	const setCreating = () => setView("create");

	const setEditing = (sc: SystemAnalysisComponent) => {
		setView("edit");
		editingComponent = sc;
		componentAttributes.init($state.snapshot(sc.attributes.component));
	};

	const doCreate = () => {
		const attr = componentAttributes.snapshot();
		const attributes: CreateSystemComponentAttributes = {
			name: attr.name,
			kindId: attr.kind.id,
			description: attr.description,
			properties: attr.properties,
			constraints: attr.constraints.map(c => c.attributes),
			controls: attr.controls.map(c => c.attributes),
			signals: attr.signals.map(s => s.attributes),
		};
		createMut?.mutate({ body: { attributes } });
	}

	const doUpdate = () => {
		if (!editingComponent) return;
		const id = editingComponent.attributes.component.id;
		const attr = componentAttributes.snapshot();
		const attributes: UpdateSystemComponentAttributes = {
			name: attr.name,
			kindId: attr.kind.id,
			description: attr.description,
			properties: attr.properties,
		};
		updateMut?.mutate({ path: { id }, body: { attributes } });
	}

	const doAdd = () => {
		if (!selectedAddComponent) return;
		const component = $state.snapshot(selectedAddComponent);
		const pos = $state.snapshot(addingPosition);
		if (pos) {
			analysis.addComponent(component, pos);
			// TODO: check if success then clear
			clear();
		} else {
			diagram.setAddingComponent(component);
			clear();
		}
	}

	const onConfirm = () => {
		if (view === "create" && componentAttributes.valid) {
			doCreate();
		} else if (view === "edit" && componentAttributes.valid) {
			doUpdate();
		} else if (view === "add" && !!selectedAddComponent) {
			doAdd();
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

const createComponentTraitsState = () => {
	const componentId = $derived(componentDialog.editingComponent?.id);

	// Constraints
	let constraint = $state<SystemComponentConstraint>();
	let createConstraintMut = $state<ReturnType<typeof makeCreateConstraintMut>>();
	let updateConstraintMut = $state<ReturnType<typeof makeUpdateConstraintMut>>();
	const savingConstraint = $derived(createConstraintMut?.isPending || updateConstraintMut?.isPending);

	const onConstraintUpdated = ({ data: updatedConstraint }: { data: SystemComponentConstraint }) => {
		componentAttributes.updateConstraint(updatedConstraint)
		constraint = undefined;
	}
	const makeCreateConstraintMut = () => createMutation(() => ({ ...createSystemComponentConstraintMutation(), onSuccess: onConstraintUpdated }));
	const makeUpdateConstraintMut = () => createMutation(() => ({ ...updateSystemComponentConstraintMutation(), onSuccess: onConstraintUpdated }));

	const editConstraint = (c?: SystemComponentConstraint) => (constraint = c ? $state.snapshot(c) : emptyTrait());
	const clearConstraint = () => (constraint = undefined);
	const saveConstraint = () => {
		if (!constraint || !createConstraintMut || !updateConstraintMut) return false;
		const cstr = $state.snapshot(constraint);
		if (!componentId) return onConstraintUpdated({ data: cstr });

		const body = { attributes: cstr.attributes };
		if (!cstr.id) return updateConstraintMut.mutate({ path: { id: cstr.id }, body });
		return createConstraintMut.mutate({ path: { id: componentId }, body });
	};

	// Signals
	let signal = $state<SystemComponentSignal>();
	let createSignalMut = $state<ReturnType<typeof makeCreateSignalMut>>();
	let updateSignalMut = $state<ReturnType<typeof makeUpdateSignalMut>>();
	const savingSignal = $derived(createSignalMut?.isPending || updateSignalMut?.isPending);

	const onSignalUpdated = ({ data: updatedSignal }: { data: SystemComponentSignal }) => {
		componentAttributes.updateSignal(updatedSignal)
		signal = undefined;
	}
	const makeCreateSignalMut = () => createMutation(() => ({ ...createSystemComponentSignalMutation(), onSuccess: onSignalUpdated }));
	const makeUpdateSignalMut = () => createMutation(() => ({ ...updateSystemComponentSignalMutation(), onSuccess: onSignalUpdated }));

	const editSignal = (s?: SystemComponentSignal) => (signal = s ? $state.snapshot(s) : emptyTrait());
	const clearSignal = () => (signal = undefined);
	const saveSignal = () => {
		if (!signal || !createSignalMut || !updateSignalMut) return false;
		const sig = $state.snapshot(signal);
		if (!componentId) return onSignalUpdated({ data: sig });

		const body = { attributes: sig.attributes };
		if (sig.id) return updateSignalMut.mutate({ path: { id: sig.id }, body });
		return createSignalMut.mutateAsync({ path: { id: componentId }, body });
	};

	// Controls
	let control = $state<SystemComponentControl>();
	let createControlMut = $state<ReturnType<typeof makeCreateControlMut>>();
	let updateControlMut = $state<ReturnType<typeof makeUpdateControlMut>>();
	const savingControl = $derived(createControlMut?.isPending || updateControlMut?.isPending);

	const onControlUpdated = ({ data: updatedControl }: { data: SystemComponentControl }) => {
		componentAttributes.updateControl(updatedControl);
		control = undefined;
	}
	const makeCreateControlMut = () => createMutation(() => ({ ...createSystemComponentControlMutation(), onSuccess: onControlUpdated }));
	const makeUpdateControlMut = () => createMutation(() => ({ ...updateSystemComponentControlMutation(), onSuccess: onControlUpdated }));

	const editControl = (c?: SystemComponentControl) => (control = c ? $state.snapshot(c) : emptyTrait());
	const clearControl = () => (control = undefined);
	const saveControl = () => {
		if (!control || !createControlMut || !updateControlMut) return false;
		const ctrl = $state.snapshot(control);
		if (!componentId) return onControlUpdated({ data: ctrl });

		const body = { attributes: ctrl.attributes };
		if (!ctrl.id) return updateControlMut.mutate({ path: { id: ctrl.id }, body });
		return createControlMut.mutate({ path: { id: componentId }, body });
	};

	const pending = $derived(savingSignal || savingConstraint || savingControl);

	return {
		setup: () => {
			createControlMut = makeCreateControlMut();
			updateControlMut = makeUpdateControlMut();
			createConstraintMut = makeCreateConstraintMut();
			updateConstraintMut = makeUpdateConstraintMut();
			createSignalMut = makeCreateSignalMut();
			updateSignalMut = makeUpdateSignalMut();
		},

		get constraint() { return constraint },
		editConstraint,
		clearConstraint,
		saveConstraint,

		get control() { return control },
		editControl,
		clearControl,
		saveControl,

		get signal() { return signal },
		editSignal,
		clearSignal,
		saveSignal,

		get pending() { return pending },
	};
};

export const componentTraits = createComponentTraitsState();