<script lang="ts">
	import { mdiFlagPlus, mdiTimelineClock, mdiTimelinePlus } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	import { eventDialog } from "./event-dialog/dialogState.svelte";
	import { milestonesDialog } from "./milestones-dialog/dialogState.svelte";

	type Props = {};
	let {}: Props = $props();
</script>

<div class="absolute top-2 right-2 w-fit mx-auto">
	<div
		class="w-fit group flex flex-col gap-2 justify-center mx-auto p-1 rounded-lg border border-surface bg-surface-100/20"
	>
		<Button color="primary" variant="fill" on:click={() => {milestonesDialog.open = true}}>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Edit Milestones</span>
				<Icon data={mdiFlagPlus} classes={{root: ""}} />
			</span>
		</Button>

		<Button color="primary" variant="fill" on:click={eventDialog.setCreating}>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Create Event</span>
				<Icon data={mdiTimelinePlus} classes={{root: "-rotate-90"}} />
			</span>
		</Button>
	</div>
</div>
