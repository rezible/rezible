<script lang="ts">
	import { Button, Icon } from 'svelte-ux';
    import { mdiPlus } from '@mdi/js';
	
	import CreateMeetingDialog from './CreateMeetingDialog.svelte';
    import ScheduledMeetingsList from './ScheduledMeetingsList.svelte';
	import UpcomingSessionsList from './UpcomingSessionsList.svelte';

	let createOpen = $state(false);

	const onMeetingCreated = () => {
		createOpen = false;
	}
</script>

<Button 
	classes={{root: "w-fit"}}
	variant="fill" 
	color="primary" 
	size="lg" 
	on:click={() => {createOpen = true}}>
	Create New Meeting
	<Icon data={mdiPlus} />
</Button>

<div class="flex gap-2">
	<div class="max-w-xl w-full border mt-2 p-2">
		<UpcomingSessionsList />
	</div>

	<div class="max-w-xl w-full border mt-2 p-2">
		<ScheduledMeetingsList />
	</div>
</div>

<CreateMeetingDialog bind:open={createOpen} onCreated={onMeetingCreated} />