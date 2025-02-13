import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemComponentConstraintMutation,
	createSystemComponentControlMutation,
	createSystemComponentSignalMutation,
	updateSystemComponentConstraintMutation,
	updateSystemComponentControlMutation,
	updateSystemComponentSignalMutation,
	type CreateSystemComponentSignalResponseBody,
	type SystemComponentConstraint,
	type SystemComponentControl,
	type SystemComponentSignal,
	type UpdateSystemComponentSignalResponseBody,
} from "$lib/api";
import { componentDialog } from "./componentDialog.svelte";

const emptyTrait = () => ({ id: "", attributes: { label: "", description: "" } });

const createComponentTraitsState = () => {
	const attrs = $derived(componentDialog.componentAttributes);

	let componentId = $state("");

	let constraint = $state<SystemComponentConstraint>();
	let signal = $state<SystemComponentSignal>();
	let control = $state<SystemComponentControl>();

	const onConstraintUpdated = ({ data: updatedConstraint }: { data: SystemComponentConstraint }) => {
		attrs.updateConstraint(updatedConstraint)
		constraint = undefined;
	}

	const makeCreateConstraintMut = () => createMutation(() => ({
		...createSystemComponentConstraintMutation(),
		onSuccess: onConstraintUpdated,
	}));
	const makeUpdateConstraintMut = () => createMutation(() => ({
		...updateSystemComponentConstraintMutation(),
		onSuccess: onConstraintUpdated,
	}));

	const onSignalUpdated = ({ data: updatedSignal }: { data: SystemComponentSignal }) => {
		attrs.updateSignal(updatedSignal)
		signal = undefined;
	}

	const makeCreateSignalMut = () => createMutation(() => ({
		...createSystemComponentSignalMutation(),
		onSuccess: onSignalUpdated,
	}));
	const makeUpdateSignalMut = () => createMutation(() => ({
		...updateSystemComponentSignalMutation(),
		onSuccess: onSignalUpdated,
	}));

	const onControlSaved = ({ data: updatedControl }: { data: SystemComponentControl }) => {
		attrs.updateControl(updatedControl);
		control = undefined;
	}

	const makeCreateControlMut = () => createMutation(() => ({
		...createSystemComponentControlMutation(),
		onSuccess: onControlSaved,
	}));

	const makeUpdateControlMut = () => createMutation(() => ({
		...updateSystemComponentControlMutation(),
		onSuccess: onControlSaved,
	}));

	let createConstraintMut = $state<ReturnType<typeof makeCreateConstraintMut>>();
	let updateConstraintMut = $state<ReturnType<typeof makeUpdateConstraintMut>>();

	const editConstraint = (c?: SystemComponentConstraint) => (constraint = c ? $state.snapshot(c) : emptyTrait());
	const clearConstraint = () => (constraint = undefined);
	const saveConstraint = () => {
		if (!constraint || !createConstraintMut || !updateConstraintMut) return false;
		const { id: constraintId, attributes } = $state.snapshot(constraint);
		const body = { attributes };
		if (!constraintId) return createConstraintMut.mutate({ path: { id: componentId }, body })
		return updateConstraintMut.mutate({ path: { componentId, constraintId }, body });
	};

	let createControlMut = $state<ReturnType<typeof makeCreateControlMut>>();
	let updateControlMut = $state<ReturnType<typeof makeUpdateControlMut>>();

	const editControl = (c?: SystemComponentControl) => (control = c ? $state.snapshot(c) : emptyTrait());
	const clearControl = () => (control = undefined);
	const saveControl = () => {
		if (!control || !createControlMut || !updateControlMut) return false;
		const { id: controlId, attributes } = $state.snapshot(control);
		const body = { attributes };
		if (!controlId) return createControlMut.mutate({ path: { id: componentId }, body });
		return updateControlMut.mutate({ path: { componentId, controlId }, body });
	};

	let createSignalMut = $state<ReturnType<typeof makeCreateSignalMut>>();
	let updateSignalMut = $state<ReturnType<typeof makeUpdateSignalMut>>();

	const editSignal = (s?: SystemComponentSignal) => (signal = s ? $state.snapshot(s) : emptyTrait());
	const clearSignal = () => (signal = undefined);
	const saveSignal = () => {
		if (!signal || !createSignalMut || !updateSignalMut) return false;
		const { id: signalId, attributes } = $state.snapshot(signal);
		const body = { attributes };
		if (!signalId) return createSignalMut.mutateAsync({ path: { id: componentId }, body })
		return updateSignalMut.mutate({ path: { componentId, signalId }, body: { attributes } });
	};

	const savingSignal = $derived(createSignalMut?.isPending || updateSignalMut?.isPending);
	const savingConstraint = $derived(createConstraintMut?.isPending || updateConstraintMut?.isPending);
	const savingControl = $derived(createControlMut?.isPending || updateControlMut?.isPending);

	const pending = $derived(savingSignal || savingConstraint || savingControl);

	const setupMutations = () => {
		createControlMut = makeCreateControlMut();
		updateControlMut = makeUpdateControlMut();
		createConstraintMut = makeCreateConstraintMut();
		updateConstraintMut = makeUpdateConstraintMut();
		createSignalMut = makeCreateSignalMut();
		updateSignalMut = makeUpdateSignalMut();
	}

	return {
		setupMutations,

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