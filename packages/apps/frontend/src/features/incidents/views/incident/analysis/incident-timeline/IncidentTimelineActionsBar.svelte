<script lang="ts">
	import RiFlagLine from "remixicon-svelte/icons/flag-line";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import { Button } from "$components/ui/button";

	import { useMilestonesDialog } from "./milestones-dialog/controller.svelte";
	import { useEventDialog } from "./event-dialog/controller.svelte";

	const milestonesDialog = useMilestonesDialog();
	const eventDialog = useEventDialog();
</script>

<div
	class="w-fit group flex flex-col gap-2 justify-center mx-auto p-1 rounded-lg border border-surface bg-surface-100/20"
>
	<Button
		color="primary"
		onclick={() => {
			milestonesDialog.open = true;
		}}
	>
		<span class="flex items-center gap-2">
			<span class="group-hover:inline hidden">Edit Milestones</span>
			<RiFlagLine aria-hidden="true" />
		</span>
	</Button>

	<Button color="primary" onclick={() => eventDialog.setCreating()}>
		<span class="flex items-center gap-2">
			<span class="group-hover:inline hidden">Create Event</span>
			<RiAddLine aria-hidden="true" />
		</span>
	</Button>
</div>
