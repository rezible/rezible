<script lang="ts">
	import {
		Button,
		Collapse,
		Icon,
		ListItem,
		SelectField,
		State,
		TextField,
		type MenuOption,
	} from "svelte-ux";
	import {
		type SystemComponentConstraint,
		type SystemComponentAttributes,
		type SystemComponent,
		type SystemComponentSignal,
		type SystemComponentControl,
	} from "$lib/api";
	import { type Snippet } from "svelte";
	import { v4 as uuidv4 } from "uuid";
	import {
		mdiBroadcast,
		mdiCancel,
		mdiCheck,
		mdiClose,
		mdiCross,
		mdiPencil,
		mdiStateMachine,
		mdiTune,
	} from "@mdi/js";
	import ConfirmButtons from "$src/components/confirm-buttons/ConfirmButtons.svelte";
	import { componentDialog } from "./componentDialog.svelte";

	const attr = $derived(componentDialog.componentAttributes);

	const kindOptions: MenuOption<string>[] = [{ label: "Service", value: "service" }];

	const makeEmpty = () => ({ id: uuidv4(), attributes: { label: "", description: "" } });

	let editConstraint = $state<SystemComponentConstraint>();
	const setEditConstraint = (c?: SystemComponentConstraint) => {
		editConstraint = c ? $state.snapshot(c) : makeEmpty();
	};
	const saveEditConstraint = () => {
		if (!editConstraint) return;
		attr.updateConstraint($state.snapshot(editConstraint));
		editConstraint = undefined;
	};

	let editSignal = $state<SystemComponentSignal>();
	const setEditSignal = (s?: SystemComponentSignal) => {
		editSignal = s ? $state.snapshot(s) : makeEmpty();
	};
	const saveEditSignal = () => {
		if (!editSignal) return;
		attr.updateSignal($state.snapshot(editSignal));
		editSignal = undefined;
	};

	let editControl = $state<SystemComponentControl>();
	const setEditControl = (c?: SystemComponentControl) => {
		editControl = c ? $state.snapshot(c) : makeEmpty();
	};
	const saveEditControl = () => {
		if (!editControl) return;
		attr.updateControl($state.snapshot(editControl));
		editControl = undefined;
	};
</script>

<div class="flex flex-row min-h-0 max-h-full h-full gap-2">
	<div class="flex flex-col gap-2 w-2/5">
		<TextField label="Name" labelPlacement="float" bind:value={attr.name} />

		<TextField label="Description" labelPlacement="float" bind:value={attr.description} />

		<SelectField
			label="Kind"
			labelPlacement="float"
			options={kindOptions}
			value={attr.kind}
			on:change={(e) => (attr.kind = e.detail.value ?? "")}
		/>
	</div>

	<div class="flex flex-col gap-2 p-1 overflow-y-auto flex-1 min-h-0 max-h-full">
		{#snippet panel(title: string, subheading: string, icon: string, content: Snippet)}
			<div class="p-2 border rounded">
				<Collapse open classes={{ root: "overflow-x-hidden", content: "p-2" }}>
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

		{#snippet attributeListItem(title: string, subheading: string, onClick: VoidFunction)}
			<ListItem
				{title}
				{subheading}
				noShadow
				class="flex-1"
				classes={{ root: "border first:border-t rounded elevation-0" }}
			>
				<div slot="actions">
					<Button iconOnly icon={mdiPencil} on:click={onClick} />
				</div>
			</ListItem>
		{/snippet}

		{#snippet constraintsPanel()}
			{#if !editConstraint}
				<div class="flex flex-col gap-2">
					{#each attr.constraints as constraint}
						{@const { label, description } = constraint.attributes}
						{@render attributeListItem(label, description, () => setEditConstraint(constraint))}
					{/each}

					<Button on:click={() => setEditConstraint()}>Add Constraint</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField
						label="Label"
						labelPlacement="float"
						bind:value={editConstraint.attributes.label}
					/>
					<TextField
						label="Description"
						labelPlacement="float"
						bind:value={editConstraint.attributes.description}
					/>

					<ConfirmButtons
						onClose={() => {
							editConstraint = undefined;
						}}
						onConfirm={saveEditConstraint}
						saveEnabled={!!editConstraint.attributes.label}
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
			{#if !editSignal}
				<div class="flex flex-col gap-2">
					{#each attr.signals as signal}
						{@render attributeListItem(
							signal.attributes.label,
							signal.attributes.description,
							() => setEditSignal(signal)
						)}
					{/each}

					<Button on:click={() => setEditSignal()}>Add Signal</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField
						label="Label"
						labelPlacement="float"
						bind:value={editSignal.attributes.label}
					/>
					<TextField
						label="Description"
						labelPlacement="float"
						bind:value={editSignal.attributes.description}
					/>

					<ConfirmButtons
						onClose={() => {
							editSignal = undefined;
						}}
						onConfirm={saveEditSignal}
						saveEnabled={!!editSignal.attributes.label}
					>
						{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
						{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
					</ConfirmButtons>
				</div>
			{/if}
		{/snippet}
		{@render panel("Signals", "Feedback from this component", mdiBroadcast, signalsPanel)}

		{#snippet controlsPanel()}
			{#if !editControl}
				<div class="flex flex-col gap-2">
					{#each attr.controls as control}
						{@render attributeListItem(
							control.attributes.label,
							control.attributes.description,
							() => setEditControl(control)
						)}
					{/each}

					<Button on:click={() => setEditControl()}>Add Control</Button>
				</div>
			{:else}
				<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
					<TextField
						label="Label"
						labelPlacement="float"
						bind:value={editControl.attributes.label}
					/>
					<TextField
						label="Description"
						labelPlacement="float"
						bind:value={editControl.attributes.description}
					/>

					<ConfirmButtons
						onClose={() => {
							editControl = undefined;
						}}
						onConfirm={saveEditControl}
						saveEnabled={!!editControl.attributes.label}
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
