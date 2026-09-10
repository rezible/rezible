<script lang="ts">
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import { Button } from "$components/ui/button";
	import CreateMeetingDialog from "$features/meetings/components/create-meeting-dialog/CreateMeetingDialog.svelte";

	let createOpen = $state(false);

	const onMeetingCreated = () => {
		createOpen = false;
	};
</script>

<div class="flex gap-2">
	<Button
		color="primary"
		onclick={() => {
			createOpen = true;
		}}
	>
		<!-- classes={{ root: "w-fit h-fit" }} -->
		Create New Meeting
		<RiArrowDownSLine class="" aria-hidden="true" />
	</Button>
</div>

<CreateMeetingDialog bind:open={createOpen} onCreated={onMeetingCreated} />
