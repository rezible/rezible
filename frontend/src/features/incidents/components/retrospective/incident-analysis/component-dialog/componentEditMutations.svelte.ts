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
	type SystemComponentSignal,
	type UpdateSystemComponentAttributes,
} from "$lib/api";
import { v4 as uuidv4 } from "uuid";
import { componentDialog } from "./componentDialog.svelte";

const makeEmpty = () => ({ id: "", attributes: { label: "", description: "" } });

const createComponentEditState = () => {
	let constraint = $state<SystemComponentConstraint>();
	let signal = $state<SystemComponentSignal>();
	let control = $state<SystemComponentControl>();

	// TODO: use correct mutations
	const onSaveConstraintSuccess = () => {
		// attr.updateSignal(res.data)
		constraint = undefined;
	}

	const makeCreateConstraintMut = () => createMutation(() => ({
		...createSystemComponentMutation(),
		onSuccess: onSaveConstraintSuccess,
	}));
	const makeUpdateConstraintMut = () => createMutation(() => ({
		...updateSystemComponentMutation(),
		onSuccess: onSaveConstraintSuccess,
	}));

	// TODO: use correct mutations
	const onSaveSignalSuccess = () => {
		// attr.updateSignal(res.data)
		signal = undefined;
	}

	const makeCreateSignalMut = () => createMutation(() => ({
		...createSystemComponentMutation(),
		onSuccess: onSaveSignalSuccess,
	}));
	const makeUpdateSignalMut = () => createMutation(() => ({
		...updateSystemComponentMutation(),
		onSuccess: onSaveSignalSuccess,
	}));

	// TODO: use correct mutations
	const onSaveControlSuccess = () => {
		// attr.updateSignal(res.data)
		control = undefined;
	}

	const makeCreateControlMut = () => createMutation(() => ({
		...createSystemComponentMutation(),
		onSuccess: onSaveControlSuccess,
	}));

	const makeUpdateControlMut = () => createMutation(() => ({
		...updateSystemComponentMutation(),
		onSuccess: onSaveControlSuccess,
	}));

	const editConstraint = (c?: SystemComponentConstraint) => (constraint = c ? $state.snapshot(c) : makeEmpty());
	const clearConstraint = () => (constraint = undefined);
	const saveConstraint = () => {
		if (!constraint) return;
		const {id, attributes} = $state.snapshot(constraint);
		if (!id) {
			// createConstraintMut.mutate({body: {attributes}})
		} else {
			// updateConstraintMut.mutate({path: {id}, body: {attributes}});
		}
	};

	const editControl = (c?: SystemComponentControl) => (control = c ? $state.snapshot(c) : makeEmpty());
	const clearControl = () => (control = undefined);
	const saveControl = () => {
		if (!control) return;
		const {id, attributes} = $state.snapshot(control);
		if (!id) {
			// createControlMut.mutate({body: {attributes}})
		} else {
			// updateControlMut.mutate({path: {id}, body: {attributes}});
		}
	};

	const editSignal = (s?: SystemComponentSignal) => (signal = s ? $state.snapshot(s) : makeEmpty());
	const clearSignal = () => (signal = undefined);
	const saveSignal = () => {
		if (!signal) return;
		const {id, attributes} = $state.snapshot(signal);
		if (!id) {
			// createSignalMut.mutate({body: {attributes}})
		} else {
			// updateSignalMut.mutate({path: {id}, body: {attributes}});
		}
	};

	let createControlMut = $state<ReturnType<typeof makeCreateControlMut>>();
	let updateControlMut = $state<ReturnType<typeof makeUpdateControlMut>>();

	let createConstraintMut = $state<ReturnType<typeof makeCreateConstraintMut>>();
	let updateConstraintMut = $state<ReturnType<typeof makeUpdateConstraintMut>>();

	let createSignalMut = $state<ReturnType<typeof makeCreateSignalMut>>();
	let updateSignalMut = $state<ReturnType<typeof makeUpdateSignalMut>>();

	const savingSignal = $derived(createSignalMut?.isPending || updateSignalMut?.isPending);
	const savingConstraint = $derived(createConstraintMut?.isPending || updateConstraintMut?.isPending);
	const savingControl = $derived(createControlMut?.isPending || updateControlMut?.isPending);

	const pending = $derived(savingSignal || savingConstraint || savingControl);

	const setup = () => {
		createControlMut = makeCreateControlMut();
		updateControlMut = makeUpdateControlMut();
		createConstraintMut = makeCreateConstraintMut();
		updateConstraintMut = makeUpdateConstraintMut();
		createSignalMut = makeCreateSignalMut();
		updateSignalMut = makeUpdateSignalMut();
	}

	return {
		setup,

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

export const componentEdits = createComponentEditState();