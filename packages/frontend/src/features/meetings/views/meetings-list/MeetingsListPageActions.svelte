<script lang="ts">
	import { mdiChevronDown } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import CreateMeetingDialog from "$features/meetings/components/create-meeting-dialog/CreateMeetingDialog.svelte";

	let createOpen = $state(false);

	const onMeetingCreated = () => {
		createOpen = false;
	};
</script>

<div class="flex gap-2">
	<Button
		color="primary"
		onclick={() => {createOpen = true}}
	> <!-- classes={{ root: "w-fit h-fit" }} -->
		Create New Meeting
		<Icon data={mdiChevronDown} />
	</Button>
</div>

<CreateMeetingDialog bind:open={createOpen} onCreated={onMeetingCreated} />