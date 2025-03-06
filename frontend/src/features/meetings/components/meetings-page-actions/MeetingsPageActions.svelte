<script lang="ts">
	import { mdiChevronDown, mdiCalendar } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	import CreateMeetingDialog from "$features/meetings/components/create-meeting-dialog/CreateMeetingDialog.svelte";

	let createOpen = $state(false);

	const onMeetingCreated = () => {
		createOpen = false;
	};
</script>

<div class="flex gap-2">
	<Button
		classes={{ root: "w-fit h-fit" }}
		variant="fill"
		color="primary"
		on:click={() => {createOpen = true}}
	>
		Create New Meeting
		<Icon data={mdiChevronDown} />
	</Button>
</div>

<CreateMeetingDialog bind:open={createOpen} onCreated={onMeetingCreated} />