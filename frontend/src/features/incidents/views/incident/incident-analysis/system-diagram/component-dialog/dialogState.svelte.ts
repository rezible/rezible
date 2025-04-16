import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemComponentMutation,
	updateSystemComponentMutation,
	type SystemComponentConstraint,
	type SystemComponentControl,
	type SystemComponentSignal,
	type CreateSystemComponentAttributes,
	type SystemAnalysisComponent,
	type SystemComponent,
	type SystemComponentAttributes,
	type UpdateSystemComponentAttributes,
} from "$lib/api";
import type { XYPosition } from "@xyflow/svelte";

import { useIncidentAnalysis } from "../../analysisState.svelte";
import { useSystemDiagram } from "../diagramState.svelte";
import { Context } from "runed";

const emptyComponentAttributes = (): SystemComponentAttributes => ({
	constraints: [],
	controls: [],
	description: "",
	kindId: "",
	name: "",
	properties: {},
	signals: []
})

type ComponentDialogView = "closed" | "add" | "create" | "edit";

const createComponentAttributesState = () => {
	let name = $state<SystemComponentAttributes["name"]>("");
	let kindId = $state<SystemComponentAttributes["kindId"]>("");
	let description = $state<SystemComponentAttributes["description"]>("");
	let constraints = $state<SystemComponentAttributes["constraints"]>([]);
	let controls = $state<SystemComponentAttributes["controls"]>([]);
	let signals = $state<SystemComponentAttributes["signals"]>([]);
	let properties = $state<SystemComponentAttributes["properties"]>({});

	let valid = $state(false);

	const init = (c?: SystemComponent) => {
		const a = c ? c.attributes : emptyComponentAttributes();
		name = a.name;
		kindId = a.kindId;
		description = a.description;
		constraints = a.constraints;
		controls = a.controls;
		signals = a.signals;
		properties = a.properties;
		valid = true;
	}

	const onUpdate = () => {
		// TODO: actually check if attributes valid;
		valid = !!name && !!kindId;
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
		get kindId() { return kindId },
		set kindId(id: string) { kindId = id; onUpdate() },
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get constraints() { return constraints },
		updateConstraint,
		get controls() { return controls },
		updateControl,
		get signals() { return signals },
		updateSignal,
		snapshot(): SystemComponentAttributes {
			return $state.snapshot({ name, kindId, description, constraints, controls, signals, properties })
		},
		get valid() { return valid },
	}
}

export const componentAttributes = createComponentAttributesState();

export class ComponentDialogState {
	editingComponent = $state<SystemAnalysisComponent>();
	selectedAddComponent = $state<SystemComponent>();
	addingPosition = $state<XYPosition>();

	analysis = useIncidentAnalysis();

	view = $state<ComponentDialogView>("closed");
	previousView = $state<ComponentDialogView>("closed");

	open = $derived(this.view !== "closed");

	creatingToAdd = $derived(this.view === "create" && this.previousView === "add");

	onAddComponent = (c: SystemComponent) => {};

	setView(v: ComponentDialogView) {
		this.previousView = $state.snapshot(this.view);
		this.view = v;
	}

	editValid = $derived(componentAttributes.valid && (this.view === "create" || this.view === "edit"));
	addValid = $derived(!!this.selectedAddComponent && this.view === "add");

	clear() {
		this.setView("closed");
		this.editingComponent = undefined;
		this.selectedAddComponent = undefined;
		this.addingPosition = undefined;
		componentAttributes.init();
	};

	goBack() {
		if (this.creatingToAdd) {
			this.setView("add");
			componentAttributes.init();
			return;
		}
		this.clear();
	}

	onSuccess({ data }: { data: SystemComponent }) {
		if (this.creatingToAdd) {
			this.goBack();
			this.selectedAddComponent = data;
			return;
		}
		this.clear();
	}

	createComponentMut = createMutation(() => ({ ...createSystemComponentMutation(), onSuccess: this.onSuccess }));
	updateComponentMut = createMutation(() => ({ ...updateSystemComponentMutation(), onSuccess: this.onSuccess }));
	
	loading = $derived(this.createComponentMut.isPending || this.updateComponentMut.isPending);

	setAdding(pos?: XYPosition) {
		this.setView("add");
		this.addingPosition = pos;
	}

	setSelectedAddComponent(c?: SystemComponent) {this.selectedAddComponent = c};

	setCreating() {
		this.setView("create");
	}

	setEditing(sc: SystemAnalysisComponent) {
		this.setView("edit");
		this.editingComponent = sc;

		componentAttributes.init($state.snapshot(sc.attributes.component));
	};

	doCreate() {
		const attr = componentAttributes.snapshot();
		const attributes: CreateSystemComponentAttributes = {
			name: attr.name,
			kindId: attr.kindId,
			description: attr.description,
			properties: attr.properties,
			constraints: attr.constraints.map(c => c.attributes),
			controls: attr.controls.map(c => c.attributes),
			signals: attr.signals.map(s => s.attributes),
		};
		this.createComponentMut.mutate({ body: { attributes } });
	}

	doUpdate() {
		if (!this.editingComponent) return;
		const id = this.editingComponent.attributes.component.id;
		const attr = componentAttributes.snapshot();
		const attributes: UpdateSystemComponentAttributes = {
			name: attr.name,
			kindId: attr.kindId,
			description: attr.description,
			properties: attr.properties,
		};
		this.updateComponentMut.mutate({ path: { id }, body: { attributes } });
	}

	async doAdd() {
		if (!this.selectedAddComponent) return;
		const component = $state.snapshot(this.selectedAddComponent);
		const pos = $state.snapshot(this.addingPosition);
		if (pos) {
			const added = await this.analysis.addComponent(component, pos);
			if (added) {
				this.clear();
			} else {
				console.log("failed to add");
			}
		} else {
			this.onAddComponent(component);
			this.clear();
		}
	}

	onConfirm() {
		if (this.view === "create" && componentAttributes.valid) {
			this.doCreate();
		} else if (this.view === "edit" && componentAttributes.valid) {
			this.doUpdate();
		} else if (this.view === "add" && !!this.selectedAddComponent) {
			this.doAdd();
		} else {
			console.error("invalid state to confirm", $state.snapshot(this.view));
			this.clear();
		}
	};
};

const componentDialogCtx = new Context<ComponentDialogState>("componentDialogState");
export const setComponentDialog = (r: ComponentDialogState) => componentDialogCtx.set(r);
export const useComponentDialog = () => componentDialogCtx.get();