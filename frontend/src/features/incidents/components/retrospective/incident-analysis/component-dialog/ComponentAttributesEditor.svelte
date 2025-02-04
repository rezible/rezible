<script lang="ts">
	import { Button, Collapse, Icon, ListItem, SelectField, State, TextField, type MenuOption } from "svelte-ux";
	import { type SystemComponentConstraint, type SystemComponentAttributes, type SystemComponent, type SystemComponentSignal, type SystemComponentControl } from "$lib/api";
	import { type Snippet } from "svelte";
	import { v4 as uuidv4 } from "uuid";
	import { mdiBroadcast, mdiCancel, mdiCheck, mdiClose, mdiCross, mdiPencil, mdiStateMachine, mdiTune } from "@mdi/js";
	import ConfirmButtons from "$src/components/confirm-buttons/ConfirmButtons.svelte";
	import { watch } from "runed";

	type Props = {
		kind: string;
		name: string;
		description: string;
		constraints: SystemComponentConstraint[];
		signals: SystemComponentSignal[];
		controls: SystemComponentControl[];
	};
	let {
		kind = $bindable(),
		name = $bindable(),
		description = $bindable(),
		constraints = $bindable(),
		signals = $bindable(),
		controls = $bindable(),
	}: Props = $props();

	const kindOptions: MenuOption<string>[] = [
		{label: "Service", value: "service"},
	];

	let editingConstraint = $state<SystemComponentConstraint>();
	const editConstraint = (c?: SystemComponentConstraint) => {
		editingConstraint = c ? $state.snapshot(c) : {id: uuidv4(), attributes: {label: "", description: ""}};
	}
	const saveEditConstraint = () => {
		const constraint = $state.snapshot(editingConstraint);
		editingConstraint = undefined;
		if (!constraint) return;
		const idx = constraints.findIndex(c => c.id === constraint.id);
		if (idx >= 0) {
			constraints[idx] = constraint;
		} else {
			constraints.push(constraint);
		}
	}

	let editingSignal = $state<SystemComponentSignal>();
	const editSignal = (s?: SystemComponentSignal) => {
		editingSignal = s ? $state.snapshot(s) : {id: uuidv4(), attributes: {label: "", description: ""}};
	}
	const saveEditSignal = () => {
		const signal = $state.snapshot(editingSignal);
		editingSignal = undefined;
		if (!signal) return;
		const idx = signals.findIndex(c => c.id === signal.id);
		if (idx >= 0) {
			signals[idx] = signal;
		} else {
			signals.push(signal);
		}
	}

	let editingControl = $state<SystemComponentControl>();
	const editControl = (c?: SystemComponentControl) => {
		editingControl = c ? $state.snapshot(c) : {id: uuidv4(), attributes: {label: "", description: ""}};
	}
	const saveEditControl = () => {
		const control = $state.snapshot(editingControl);
		editingControl = undefined;
		if (!control) return;
		const idx = controls.findIndex(c => c.id === control.id);
		if (idx >= 0) {
			controls[idx] = control;
		} else {
			controls.push(control);
		}
	}
</script>

<div class="flex flex-row min-h-0 max-h-full h-full gap-2">
	<div class="flex flex-col gap-2 w-2/5">
		<TextField label="Name" labelPlacement="float" bind:value={name} />

		<TextField label="Description" labelPlacement="float" bind:value={description} />

		<SelectField label="Kind" labelPlacement="float" options={kindOptions} value={kind} on:change={e => kind = e.detail.value ?? ""} />
	</div>

	<div class="flex flex-col gap-2 p-1 overflow-y-auto flex-1 min-h-0 max-h-full">
		{#snippet panel(
			title: string,
			subheading: string,
			icon: string,
			content: Snippet
		)}
			<div class="p-2 border rounded">
				<Collapse
					open
					classes={{ root: "overflow-x-hidden", content: "p-2" }}
				>
					<ListItem
						slot="trigger"
						{title}
						{subheading}
						{icon}
						classes={{ root: "pl-0" }}
						avatar={{
							class: "bg-surface-content/50 text-surface-100/90",
						}}
						class="flex-1"
						noShadow
					/>
					{@render content()}
				</Collapse>
			</div>
		{/snippet}

		{#snippet attributeListItem(label: string, description: string, onClick: VoidFunction)}
			<ListItem
				title={label}
				subheading={description}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button iconOnly icon={mdiPencil} on:click={onClick} />
				</div>
			</ListItem>
		{/snippet}

		{#snippet constraintsPanel()}
			{#if !editingConstraint}
				<div class="flex flex-col gap-2">
					{#each constraints as constraint}
						{@render attributeListItem(constraint.attributes.label, constraint.attributes.description, () => editConstraint(constraint))}
					{/each}

					<Button on:click={() => editConstraint()}>Add Constraint</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField label="Label" labelPlacement="float" bind:value={editingConstraint.attributes.label} />
					<TextField label="Description" labelPlacement="float" bind:value={editingConstraint.attributes.description} />
					
					<ConfirmButtons
						onClose={() => {editingConstraint = undefined}} 
						onConfirm={saveEditConstraint}
						saveEnabled={!!editingConstraint.attributes.label}
					>
						{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
						{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
					</ConfirmButtons>
				</div>
			{/if}
		{/snippet}
		{@render panel(
			"Constraints",
			"Conditions under which this component operates normally",
			mdiStateMachine,
			constraintsPanel
		)}

		{#snippet signalsPanel()}
			{#if !editingSignal}
				<div class="flex flex-col gap-2">
					{#each signals as signal}
						{@render attributeListItem(signal.attributes.label, signal.attributes.description, () => editSignal(signal))}
					{/each}

					<Button on:click={() => editSignal()}>Add Signal</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField label="Label" labelPlacement="float" bind:value={editingSignal.attributes.label} />
					<TextField label="Description" labelPlacement="float" bind:value={editingSignal.attributes.description} />
					
					<ConfirmButtons
						onClose={() => {editingSignal = undefined}} 
						onConfirm={saveEditSignal}
						saveEnabled={!!editingSignal.attributes.label}
					>
						{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
						{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
					</ConfirmButtons>
				</div>
			{/if}
		{/snippet}
		{@render panel(
			"Signals",
			"Feedback from this component",
			mdiBroadcast,
			signalsPanel
		)}

		{#snippet controlsPanel()}
			{#if !editingControl}
				<div class="flex flex-col gap-2">
					{#each controls as control}
						{@render attributeListItem(control.attributes.label, control.attributes.description, () => editControl(control))}
					{/each}

					<Button on:click={() => editSignal()}>Add Control</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField label="Label" labelPlacement="float" bind:value={editingControl.attributes.label} />
					<TextField label="Description" labelPlacement="float" bind:value={editingControl.attributes.description} />
					
					<ConfirmButtons
						onClose={() => {editingControl = undefined}} 
						onConfirm={saveEditControl}
						saveEnabled={!!editingControl.attributes.label}
					>
						{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
						{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
					</ConfirmButtons>
				</div>
			{/if}
		{/snippet}
		{@render panel(
			"Controls",
			"Methods that can alter the behaviour of this component",
			mdiTune,
			controlsPanel
		)}
	</div>
</div>