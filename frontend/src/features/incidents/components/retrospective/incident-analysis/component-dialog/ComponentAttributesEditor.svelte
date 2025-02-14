<script lang="ts">
	import { Button, Collapse, Icon, ListItem, SelectField, TextField, type MenuOption } from "svelte-ux";
	import { type Snippet } from "svelte";
	import { mdiBroadcast, mdiCheck, mdiClose, mdiPencil, mdiStateMachine, mdiTune } from "@mdi/js";
	import { createQuery } from "@tanstack/svelte-query";
	import { listSystemComponentKindsOptions } from "$lib/api";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { componentDialog } from "./componentDialog.svelte";
	import { componentTraits } from "./componentTraitMutations.svelte";
	import { systemComponentKindQueryMenuOptionSelect } from "$lib/systemComponents";

	const attr = $derived(componentDialog.componentAttributes);

	componentTraits.setupMutations();

	const listComponentKindOptions = createQuery(() => ({
		...listSystemComponentKindsOptions({}),
		select: systemComponentKindQueryMenuOptionSelect,
	}));
</script>

<div class="flex flex-row min-h-0 max-h-full h-full gap-2">
	<div class="flex flex-col gap-2 w-2/5">
		<TextField label="Name" labelPlacement="float" bind:value={attr.name} />

		<TextField label="Description" labelPlacement="float" bind:value={attr.description} />

		<SelectField
			label="Kind"
			labelPlacement="float"
			options={listComponentKindOptions.data}
			loading={listComponentKindOptions.isFetching}
			value={attr.kind}
			on:change={(e) => (attr.kind = e.detail.value ?? "")}
		/>
	</div>

	<div class="flex flex-col gap-2 p-1 overflow-y-auto flex-1 min-h-0 max-h-full">
		{@render traitPanel(
			"Constraints",
			"Conditions under which this component operates normally",
			mdiStateMachine,
			constraintsPanel
		)}
		{@render traitPanel("Signals", "Feedback supplied by this component", mdiBroadcast, signalsPanel)}
		{@render traitPanel(
			"Controls",
			"Methods that can alter the behaviour of this component",
			mdiTune,
			controlsPanel
		)}
	</div>
</div>

{#snippet traitPanel(title: string, subheading: string, icon: string, content: Snippet)}
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
	{@const editConstraint = componentTraits.constraint}
	{#if !editConstraint}
		<div class="flex flex-col gap-2">
			{#each attr.constraints as c}
				{@const { label, description } = c.attributes}
				{@render attributeListItem(label, description, () => componentTraits.editConstraint(c))}
			{/each}

			<Button on:click={() => componentTraits.editConstraint()}>Add Constraint</Button>
		</div>
	{:else}
		<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
			<TextField label="Label" labelPlacement="float" bind:value={editConstraint.attributes.label} />
			<TextField
				label="Description"
				labelPlacement="float"
				bind:value={editConstraint.attributes.description}
			/>

			<ConfirmButtons
				onClose={componentTraits.clearConstraint}
				loading={componentTraits.pending}
				onConfirm={componentTraits.saveConstraint}
				saveEnabled={!!editConstraint.attributes.label}
			>
				{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
				{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
			</ConfirmButtons>
		</div>
	{/if}
{/snippet}

{#snippet signalsPanel()}
	{@const editSignal = componentTraits.signal}
	{#if !editSignal}
		<div class="flex flex-col gap-2">
			{#each attr.signals as signal}
				{@render attributeListItem(signal.attributes.label, signal.attributes.description, () =>
					componentTraits.editSignal(signal)
				)}
			{/each}

			<Button on:click={() => componentTraits.editSignal()}>Add Signal</Button>
		</div>
	{:else}
		<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
			<TextField label="Label" labelPlacement="float" bind:value={editSignal.attributes.label} />
			<TextField
				label="Description"
				labelPlacement="float"
				bind:value={editSignal.attributes.description}
			/>

			<ConfirmButtons
				onClose={componentTraits.clearSignal}
				loading={componentTraits.pending}
				onConfirm={componentTraits.saveSignal}
				saveEnabled={!!editSignal.attributes.label}
			>
				{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
				{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
			</ConfirmButtons>
		</div>
	{/if}
{/snippet}

{#snippet controlsPanel()}
	{@const editControl = componentTraits.control}
	{#if !editControl}
		<div class="flex flex-col gap-2">
			{#each attr.controls as control}
				{@render attributeListItem(control.attributes.label, control.attributes.description, () =>
					componentTraits.editControl(control)
				)}
			{/each}

			<Button on:click={() => componentTraits.editControl()}>Add Control</Button>
		</div>
	{:else}
		<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
			<TextField label="Label" labelPlacement="float" bind:value={editControl.attributes.label} />
			<TextField
				label="Description"
				labelPlacement="float"
				bind:value={editControl.attributes.description}
			/>

			<ConfirmButtons
				onClose={componentTraits.clearControl}
				onConfirm={componentTraits.saveControl}
				loading={componentTraits.pending}
				saveEnabled={!!editControl.attributes.label}
			>
				{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
				{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
			</ConfirmButtons>
		</div>
	{/if}
{/snippet}
