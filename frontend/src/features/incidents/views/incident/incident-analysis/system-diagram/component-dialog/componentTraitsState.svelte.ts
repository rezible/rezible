import { type SystemComponentConstraint, createSystemComponentConstraintMutation, updateSystemComponentConstraintMutation, type SystemComponentSignal, createSystemComponentSignalMutation, updateSystemComponentSignalMutation, type SystemComponentControl, createSystemComponentControlMutation, updateSystemComponentControlMutation } from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { componentAttributes, useComponentDialog } from "./dialogState.svelte";

const emptyTrait = () => ({ id: "", attributes: { label: "", description: "" } });

export class ComponentTraitsState {
	dialog = useComponentDialog();

	componentId = $derived(this.dialog.editingComponent?.id);

	// Constraints
	constraint = $state<SystemComponentConstraint>();
	createConstraintMut = createMutation(() => ({ ...createSystemComponentConstraintMutation(), onSuccess: this.onConstraintUpdated }));
	updateConstraintMut = createMutation(() => ({ ...updateSystemComponentConstraintMutation(), onSuccess: this.onConstraintUpdated }));

	savingConstraint = $derived(this.createConstraintMut.isPending || this.updateConstraintMut.isPending);

	onConstraintUpdated({ data: updatedConstraint }: { data: SystemComponentConstraint }) {
		componentAttributes.updateConstraint(updatedConstraint)
		this.constraint = undefined;
	}

	editConstraint(c?: SystemComponentConstraint) {
		this.constraint = c ? $state.snapshot(c) : emptyTrait();
	};
	clearConstraint() { this.constraint = undefined };

	saveConstraint() {
		if (!this.constraint) return false;
		const cstr = $state.snapshot(this.constraint);
		if (!this.componentId) return this.onConstraintUpdated({ data: cstr });

		const body = { attributes: cstr.attributes };
		if (!cstr.id) return this.updateConstraintMut.mutate({ path: { id: cstr.id }, body });
		return this.createConstraintMut.mutate({ path: { id: this.componentId }, body });
	};

	// Signals
	signal = $state<SystemComponentSignal>();
	createSignalMut = createMutation(() => ({ ...createSystemComponentSignalMutation(), onSuccess: this.onSignalUpdated }));
	updateSignalMut = createMutation(() => ({ ...updateSystemComponentSignalMutation(), onSuccess: this.onSignalUpdated }));
	savingSignal = $derived(this.createSignalMut.isPending || this.updateSignalMut.isPending);

	onSignalUpdated({ data: updatedSignal }: { data: SystemComponentSignal }) {
		componentAttributes.updateSignal(updatedSignal)
		this.signal = undefined;
	}

	editSignal(s?: SystemComponentSignal) {
		this.signal = s ? $state.snapshot(s) : emptyTrait();
	};

	clearSignal() {
		this.signal = undefined;
	}

	saveSignal() {
		if (!this.signal) return false;
		const sig = $state.snapshot(this.signal);
		if (!this.componentId) return this.onSignalUpdated({ data: sig });

		const body = { attributes: sig.attributes };
		if (sig.id) return this.updateSignalMut.mutate({ path: { id: sig.id }, body });
		return this.createSignalMut.mutateAsync({ path: { id: this.componentId }, body });
	};

	// Controls
	control = $state<SystemComponentControl>();
	createControlMut = createMutation(() => ({ ...createSystemComponentControlMutation(), onSuccess: this.onControlUpdated }));
	updateControlMut = createMutation(() => ({ ...updateSystemComponentControlMutation(), onSuccess: this.onControlUpdated }));
	savingControl = $derived(this.createControlMut?.isPending || this.updateControlMut.isPending);

	onControlUpdated({ data: updatedControl }: { data: SystemComponentControl }) {
		componentAttributes.updateControl(updatedControl);
		this.control = undefined;
	}

	editControl(c?: SystemComponentControl) {
		this.control = c ? $state.snapshot(c) : emptyTrait();
	};
	clearControl() {
		this.control = undefined;
	};
	saveControl() {
		if (!this.control) return false;
		const ctrl = $state.snapshot(this.control);
		if (!this.componentId) return this.onControlUpdated({ data: ctrl });

		const body = { attributes: ctrl.attributes };
		if (!ctrl.id) return this.updateControlMut.mutate({ path: { id: ctrl.id }, body });
		return this.createControlMut.mutate({ path: { id: this.componentId }, body });
	};

	pending = $derived(this.savingSignal || this.savingConstraint || this.savingControl);
};
