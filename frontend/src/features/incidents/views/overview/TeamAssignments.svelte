<script lang="ts">
	import type { Incident } from '$lib/api';
    import Avatar from '$components/avatar/Avatar.svelte';
    import { Button, Header, Icon } from 'svelte-ux';
    import { mdiPencil } from '@mdi/js';

	interface Props { incident: Incident };
	let { incident }: Props = $props();
	
	const assignments = $derived(incident.attributes.teams);
</script>

<Header title="Teams" classes={{ root: 'min-h-8', title: 'text-md text-neutral-100' }}>
	<div slot="actions">
		<Button
			size="sm"
			classes={{ root: 'h-8 text-neutral-200' }}
			on:click={() => {}}
		>
			<Icon data={mdiPencil} />
		</Button>
	</div>
</Header>

<div class="flex flex-col gap-2">
	{#each assignments as assignment}
		{@const team = assignment.team}
		<div class="">
			<span class="items-center flex flex-row gap-2">
				<Avatar kind="team" id={team.id} />
				<div class="flex flex-col">
					<span class="text-lg">{team.attributes.name}</span>
				</div>
			</span>
		</div>
	{/each}
</div>
