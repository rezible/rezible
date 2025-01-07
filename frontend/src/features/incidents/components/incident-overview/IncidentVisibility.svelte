<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { Button, Field, Header, Icon, Switch } from 'svelte-ux';
	import { mdiPencil, mdiPencilOutline } from '@mdi/js';
	import { updateIncidentMutation, type Incident } from '$lib/api';
	import ConfirmChangeButtons from '$components/confirm-buttons/ConfirmButtons.svelte';
    import { incidentCtx } from '$features/incidents/lib/context';

	type Props = {};
	const {}: Props = $props();

	const incident = incidentCtx.get();

	let editing = $state(false);
	let newPrivacy = $state(incident.attributes.private);
	let viewers = $state([]);

	const resetState = () => {
		editing = false;
		viewers = [];
	};

	const changed = $derived(newPrivacy !== incident.attributes.private);

	const update = createMutation(() => ({
		...updateIncidentMutation(),
		onSuccess: () => {
			resetState();
		}
	}));
	const doUpdate = () => update.mutate({path: {id: incident.id}, body: {attributes: { private: newPrivacy }}});
</script>

{#if !editing}
	<Header
		title="Incident Visibility"
		classes={{ root: 'min-h-8', title: 'text-md text-neutral-100' }}
	>
		<div slot="actions" class:hidden={editing}>
			<Button
				size="sm"
				classes={{ root: 'h-8' }}
				on:click={() => {
					editing = true;
				}}
			>
				<Icon data={mdiPencil} />
			</Button>
		</div>
	</Header>
	{#if incident.attributes.private}
		<span class="text-neutral-content">Restricted</span>
	{:else}
		<span class="text-neutral-content">Public</span>
	{/if}
{:else}
	<div class="flex flex-col gap-2 w-64">
		<Field label="Restrict Visibility" let:id>
			<Switch {id} bind:checked={newPrivacy} />
		</Field>

		<!-- TODO: add viewers -->

		<ConfirmChangeButtons
			disabled={update.isPending}
			saveEnabled={changed}
			onClose={() => resetState()}
			onConfirm={doUpdate}
		/>
	</div>
{/if}
