<script lang="ts">
	import {
		Button,
		Checkbox,
		cls,
		Collapse,
		Header,
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
		getSystemComponentOptions,
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
	import { relationshipDialog } from "./relationshipDialog.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import { SvelteSet } from "svelte/reactivity";

	const attr = $derived(relationshipDialog.attributes);

	const sourceComponentQuery = createQuery(() => ({
		...getSystemComponentOptions({path: {id: attr.sourceId}}),
		enabled: !!attr.sourceId,
	}));

	const targetComponentQuery = createQuery(() => ({
		...getSystemComponentOptions({path: {id: attr.targetId}}),
		enabled: !!attr.targetId,
	}));

	const sourceComponent = $derived(sourceComponentQuery.data?.data);
	const targetComponent = $derived(targetComponentQuery.data?.data);

	const includedSignals = $derived(new SvelteSet(attr.feedbackSignals.map(v => v.id)));
	const includedControls = $derived(new SvelteSet(attr.controlActions.map(v => v.id)));

	const toggleSignalIncluded = (id: string) => {
		if (!includedSignals.delete(id)) includedSignals.add(id);
	}

	const toggleControlIncluded = (id: string) => {
		if (!includedControls.delete(id)) includedControls.add(id);
	}
</script>

{#snippet signalListItem(signal: SystemComponentSignal)}
{@const included = includedSignals.has(signal.id)}
	<div>
		<ListItem
			title={signal.attributes.label}
			subheading={signal.attributes.description}
			on:click={() => {toggleSignalIncluded(signal.id)}}
			noShadow
			class={cls(
				"px-4 py-2",
				"cursor-pointer transition-shadow duration-100",
				"hover:bg-surface-100 hover:outline",
				included ? "bg-surface-100 shadow-md" : "",
			)}
		>
			<div slot="actions">
				<Checkbox dense checked={included} on:change={() => {toggleSignalIncluded(signal.id)}} />
			</div>
		</ListItem>
	</div>
{/snippet}

{#snippet controlListItem(control: SystemComponentControl)}
{@const included = includedControls.has(control.id)}
	<div>
		<ListItem
			title={control.attributes.label}
			subheading={control.attributes.description}
			on:click={() => {toggleControlIncluded(control.id)}}
			noShadow
			class={cls(
				"px-4 py-2",
				"cursor-pointer transition-shadow duration-100",
				"hover:bg-surface-100 hover:outline",
				included ? "bg-surface-100 shadow-md" : "",
			)}
		>
			<div slot="actions">
				<Checkbox dense checked={included} on:change={() => {toggleControlIncluded(control.id)}} />
			</div>
		</ListItem>
	</div>
{/snippet}

{#snippet relationshipComponentColumn(cmp: SystemComponent)}
	<div class="p-2">
		<Header title={cmp.attributes.name} />
	</div>

	<div class="p-2 border">
		<Header title="Signals">
			<svelte:fragment slot="actions">
				<Button size="sm">Create New</Button>
			</svelte:fragment>
		</Header>
		<div class="flex flex-col gap-2">
			{#each cmp.attributes.signals as signal}
				{@render signalListItem(signal)}
			{/each}
		</div>
	</div>

	<div class="p-2 border">
		<Header title="Controls">
			<svelte:fragment slot="actions">
				<Button size="sm">Create New</Button>
			</svelte:fragment>
		</Header>
		<div class="flex flex-col gap-2">
			{#each cmp.attributes.controls as control}
				{@render controlListItem(control)}
			{/each}
		</div>
	</div>
{/snippet}

<div class="grid grid-cols-2 min-h-0 max-h-full h-full gap-2">
	<div class="border overflow-y-auto">
		<LoadingQueryWrapper query={sourceComponentQuery} view={relationshipComponentColumn} />
	</div>
	<div class="border overflow-y-auto">
		<LoadingQueryWrapper query={targetComponentQuery} view={relationshipComponentColumn} />
	</div>
</div>
