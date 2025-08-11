<script lang="ts">
	import { mdiChevronDown } from "@mdi/js";
	import Button from "$components/button/Button.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import CreateMeetingDialog from "$features/meetings-list/components/create-meeting-dialog/CreateMeetingDialog.svelte";

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