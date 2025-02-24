<script lang="ts">
	import type { SystemComponent, SystemComponentSignal, SystemComponentControl } from "$lib/api";
	import { Button, Header, ListItem } from "svelte-ux";
	import { relationshipDialog } from "./relationshipDialogState.svelte";
	import { mdiPencil, mdiPlus } from "@mdi/js";
	import { SvelteSet } from "svelte/reactivity";
	import LabelDescriptionEditor from "./LabelDescriptionEditor.svelte";

	type Props = {
		component: SystemComponent;
	};
	const { component }: Props = $props();

	const attr = $derived(component.attributes);
	const relAttr = $derived(relationshipDialog.attributes);

	const includedSignalIds = $derived(
		new SvelteSet(relAttr.feedbackSignals.map((v) => v.attributes.signalId))
	);
	const excludedSignals = $derived(attr.signals.filter((s) => !includedSignalIds.has(s.id)));

	const includedControlIds = $derived(
		new SvelteSet(relAttr.controlActions.map((v) => v.attributes.controlId))
	);
	const excludedControls = $derived(attr.controls.filter((s) => !includedControlIds.has(s.id)));

	const includeFeedbackSignal = (signalId: string) =>
		relAttr.setFeedbackSignal({ signalId, description: "" });

	const includeControlAction = (controlId: string) =>
		relAttr.setControlAction({ controlId, description: "" });

	let editingSignal = $state<SystemComponentSignal>();
	const setEditingSignal = (s?: SystemComponentSignal) => {
		editingSignal = !!s ? $state.snapshot(s) : { id: "", attributes: { label: "", description: "" } };
	};
	const cancelEditingSignal = () => (editingSignal = undefined);

	const saveEditingSignal = () => {
		const sig = $state.snapshot(editingSignal);
		alert("save signal edit");
		editingSignal = undefined;
	}

	let editingControl = $state<SystemComponentControl>();
	const setEditingControl = (s?: SystemComponentControl) => {
		editingControl = !!s ? $state.snapshot(s) : { id: "", attributes: { label: "", description: "" } };
	};
	const cancelEditingControl = () => (editingControl = undefined);
	const saveEditingControl = () => {
		const ctrl = $state.snapshot(editingControl);
		alert("save control edit");
		editingControl = undefined;
	};

	interface RelationshipListItem {
		id: string;
		attributes: {
			label: string;
			description: string;
		};
	}
</script>

<div class="flex flex-col h-full min-h-0">
	<div class="p-1">
		<Header title={attr.name} />
	</div>

	<div class="flex flex-col flex-1 gap-2 p-1 overflow-y-auto min-h-0">
		<div class="border p-2 flex flex-col gap-2 overflow-y-auto">
			<Header title="Signals">
				<svelte:fragment slot="actions">
					{#if excludedSignals.length > 0 && !editingSignal}
						<Button size="sm" on:click={() => setEditingSignal()}>Create New</Button>
					{/if}
				</svelte:fragment>
			</Header>

			<div class="flex flex-col gap-1 min-h-0 overflow-x-hidden overflow-y-auto">
				{#if editingSignal}
					<LabelDescriptionEditor
						bind:label={editingSignal.attributes.label}
						bind:description={editingSignal.attributes.description}
						onCancel={cancelEditingSignal}
						onConfirm={saveEditingSignal}
					/>
				{:else}
					<div class="flex flex-col min-h-0 overflow-y-auto gap-2">
						{#each excludedSignals as signal}
							{@render listItem(signal, setEditingSignal, includeFeedbackSignal)}
						{/each}

						{#if excludedSignals.length === 0}
							<Button on:click={() => setEditingSignal()}>Create Signal</Button>
						{/if}
					</div>
				{/if}
			</div>
		</div>

		<div class="border p-2 flex flex-col gap-2 overflow-y-auto">
			<Header title="Controls">
				<svelte:fragment slot="actions">
					{#if excludedControls.length > 0 && !editingControl}
						<Button size="sm" on:click={() => setEditingControl()}>Create New</Button>
					{/if}
				</svelte:fragment>
			</Header>

			<div class="flex flex-col gap-1 overflow-x-hidden overflow-y-auto">
				{#if editingControl}
					<LabelDescriptionEditor
						bind:label={editingControl.attributes.label}
						bind:description={editingControl.attributes.description}
						onCancel={cancelEditingControl}
						onConfirm={saveEditingControl}
					/>
				{:else}
					<div class="flex flex-col min-h-0 overflow-y-auto gap-2">
						{#each excludedControls as control}
							{@render listItem(control, setEditingControl, includeControlAction)}
						{/each}

						{#if excludedControls.length === 0}
							<Button on:click={() => setEditingControl()}>Create Control</Button>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>

{#snippet listItem(item: RelationshipListItem, editFunc: (item?: RelationshipListItem) => void, includeFunc: (id: string) => void)}
	<div class="p-1">
		<ListItem
			title={item.attributes.label}
			subheading={item.attributes.description}
			noShadow
			noBackground
			class="px-4 py-2 transition-shadow duration-100 hover:bg-surface-100 hover:outline"
		>
			<svelte:fragment slot="actions">
				<Button size="sm" iconOnly icon={mdiPencil} on:click={() => {editFunc(item)}} />
				<Button size="sm" iconOnly icon={mdiPlus} on:click={() => {includeFunc(item.id)}} />
			</svelte:fragment>
		</ListItem>
	</div>
{/snippet}
