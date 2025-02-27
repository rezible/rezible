<script lang="ts">
	import { mdiShapeSquareRoundedPlus, mdiTimelinePlus } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	import { eventDialog } from "./event-dialog/dialogState.svelte";

	type Props = {};
	let {}: Props = $props();
</script>

<div class="absolute top-2 right-2 w-fit mx-auto">
	<div
		class="w-fit group flex gap-2 justify-center mx-auto p-2 rounded-lg border border-surface bg-surface-100/20"
	>
		<Button color="primary" variant="fill" on:click={eventDialog.setCreating}>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Create Event</span>
				<Icon data={mdiTimelinePlus} classes={{root: "-rotate-90"}} />
			</span>
		</Button>
	</div>
</div>
