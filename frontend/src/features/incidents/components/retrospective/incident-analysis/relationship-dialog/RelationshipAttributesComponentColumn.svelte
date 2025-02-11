<script lang="ts">
	import { type SystemComponent, type SystemComponentSignal, type SystemComponentControl } from "$lib/api";
	import { Button, Checkbox, cls, Header, Icon, ListItem, TextField } from "svelte-ux";
	import { relationshipDialog } from "./relationshipDialog.svelte";
	import ConfirmButtons from "$src/components/confirm-buttons/ConfirmButtons.svelte";
	import { mdiCheck, mdiClose } from "@mdi/js";
	import { SvelteSet } from "svelte/reactivity";

	type Props = {
		component: SystemComponent;
	};
	const { component }: Props = $props();

	const attr = $derived(component.attributes);
	const relAttr = $derived(relationshipDialog.attributes);

	const includedSignalIds = $derived(new SvelteSet(relAttr.feedbackSignals.map((v) => v.attributes.signal_id)));
	const excludedSignals = $derived(attr.signals.filter(s => !includedSignalIds.has(s.id)));

	const includedControlIds = $derived(new SvelteSet(relAttr.controlActions.map((v) => v.attributes.control_id)));
	const excludedControls = $derived(attr.controls.filter(s => !includedControlIds.has(s.id)));

	const includeFeedbackSignal = (signal_id: string) => {
		relAttr.setFeedbackSignal({ signal_id, description: "" });
	};

	const includeControlAction = (control_id: string) => {
		relAttr.setControlAction({ control_id, description: "" });
	};

	let editingSignal = $state<SystemComponentSignal>();
	const setEditingSignal = (s?: SystemComponentSignal) => {
		editingSignal = { id: "", attributes: { label: "", description: "" } };
	};
	const cancelEditingSignal = () => {
		editingSignal = undefined;
	};
	const saveEditingSignal = () => {
		editingSignal = undefined;
	};

	let editingControl = $state<SystemComponentControl>();
	const setEditingControl = (s?: SystemComponentControl) => {
		editingControl = { id: "", attributes: { label: "", description: "" } };
	};
	const cancelEditingControl = () => {
		editingControl = undefined;
	};
	const saveEditingControl = () => {
		editingControl = undefined;
	};

	interface RelationshipListItem {
		id: string;
		attributes: {
			label: string;
			description: string;
		}
	}
</script>

<div class="p-2">
	<Header title={attr.name} />
</div>

<div class="p-2 border">
	<Header title="Signals">
		<svelte:fragment slot="actions">
			{#if excludedSignals.length > 0 && !editingSignal}
				<Button size="sm">Create New</Button>
			{/if}
		</svelte:fragment>
	</Header>

	<div class="flex flex-col gap-2 overflow-x-hidden p-1">
		{#if !editingSignal}
			<div class="flex flex-col gap-2">
				{#each excludedSignals as signal}
					{@render listItem(signal, includeFeedbackSignal)}
				{/each}

				{#if excludedSignals.length === 0}
					<Button on:click={() => setEditingSignal()}>Create Signal</Button>
				{/if}
			</div>
		{:else}
			{@const saveEnabled = !!editingSignal.attributes.label}
			<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
				<TextField label="Label" labelPlacement="float" bind:value={editingSignal.attributes.label} />
				<TextField
					label="Description"
					labelPlacement="float"
					bind:value={editingSignal.attributes.description}
				/>

				{@render confirmEditButtons(saveEnabled, cancelEditingSignal, saveEditingSignal)}
			</div>
		{/if}
	</div>
</div>

<div class="p-2 border">
	<Header title="Controls">
		<svelte:fragment slot="actions">
			{#if excludedControls.length > 0 && !editingControl}
				<Button size="sm">Create New</Button>
			{/if}
		</svelte:fragment>
	</Header>

	<div class="flex flex-col gap-2 overflow-x-hidden p-1">
		{#if !editingControl}
			<div class="flex flex-col gap-2">
				{#each excludedControls as control}
					{@render listItem(control, includeControlAction)}
				{/each}

				{#if excludedControls.length === 0}
					<Button on:click={() => setEditingControl()}>Create Control</Button>
				{/if}
			</div>
		{:else}
			{@const saveEnabled = !!editingControl.attributes.label}
			<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
				<TextField
					label="Label"
					labelPlacement="float"
					bind:value={editingControl.attributes.label}
				/>
				<TextField
					label="Description"
					labelPlacement="float"
					bind:value={editingControl.attributes.description}
				/>

				{@render confirmEditButtons(saveEnabled, cancelEditingControl, saveEditingControl)}
			</div>
		{/if}
	</div>
</div>

{#snippet confirmEditButtons(saveEnabled: boolean, onClose: VoidFunction, onConfirm: VoidFunction)}
	<ConfirmButtons {onClose} {onConfirm} {saveEnabled}>
		{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
		{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
	</ConfirmButtons>
{/snippet}

{#snippet listItem(item: RelationshipListItem, toggleFunc: (id: string) => void)}
	<div>
		<ListItem
			title={item.attributes.label}
			subheading={item.attributes.description}
			noShadow
			noBackground
			class="px-4 py-2 transition-shadow duration-100 hover:bg-surface-100 hover:outline"
		>
			<div slot="actions">
				<Button dense on:click={() => {toggleFunc(item.id)}}>
					include
				</Button>
			</div>
		</ListItem>
	</div>
{/snippet}
