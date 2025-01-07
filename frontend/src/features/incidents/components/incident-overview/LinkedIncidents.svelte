<script lang="ts">
	import { mdiPencil } from '@mdi/js';
	import { Header, Button, Icon } from 'svelte-ux';
    import { incidentCtx } from '$features/incidents/lib/context';

	type Props = {};
	const {}: Props = $props();

	const incident = incidentCtx.get();

	let editing = $state(false);
	const linkedIncidents = $derived(incident.attributes.linked_incidents);
</script>

<Header title="Linked Incidents" classes={{ root: 'min-h-8', title: 'text-md text-neutral-100' }}>
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

<div class="flex flex-col gap-2">
	{#each linkedIncidents as linked}
		<a href="#/incidents/{linked.incident_id}">
			<div class="border p-2 hover:bg-accent cursor-pointer">
				<div class="text-lg">{linked.incident_title}</div>
				<div class="text-md">{linked.incident_summary}</div>
			</div>
		</a>
	{/each}
	{#if linkedIncidents.length === 0}
		<span>no linked incidents</span>
	{/if}
</div>
