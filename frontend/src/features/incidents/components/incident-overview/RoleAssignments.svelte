<script lang="ts">
	import type { Incident } from '$lib/api';
	import Avatar from '$components/avatar/Avatar.svelte';
    import { Button, Header, Icon } from 'svelte-ux';
    import { mdiPencil } from '@mdi/js';

	interface Props { incident: Incident };
	let { incident }: Props = $props();

	const assignments = $derived(incident.attributes.roles);
	let editing = $state(false);
</script>

<Header title="Responders" classes={{ root: 'min-h-8', title: 'text-md text-neutral-100' }}>
	<div slot="actions" class:hidden={editing}>
		<Button
			size="sm"
			classes={{ root: 'h-8 text-neutral-200' }}
			on:click={() => {}}
		>
			<Icon data={mdiPencil} />
		</Button>
	</div>
</Header>

{#each assignments as assignment}
	<div class="">
		<span class="items-center flex flex-row gap-2">
			<Avatar kind="user" id={assignment.user.attributes.name} />
			<div class="flex flex-col">
				<span class="text-lg">{assignment.user.attributes.name}</span>
				<span class="text-gray-700">{assignment.role.attributes.name}</span>
			</div>
		</span>
	</div>
{/each}
