<script lang="ts">
	import { Button, Dialog, Icon } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import MilestoneAttributesEditor from "./MilestoneAttributesEditor.svelte";
	import { milestoneDialog } from "./dialogState.svelte";
	import { milestoneAttributes } from "./milestoneAttributes.svelte";

	milestoneDialog.setup();

	const creating = $derived(milestoneDialog.view === "create");
</script>

<Dialog
	open={milestoneDialog.open}
	on:close={() => milestoneDialog.clear}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-5xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">{creating ? "Create" : "Edit"} Incident Milestone</span>
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto flex">
		{#if milestoneDialog.open}
			<MilestoneAttributesEditor />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			onClose={milestoneDialog.clear}
			confirmText={creating ? "Create" : "Save"}
			onConfirm={milestoneDialog.confirm}
			saveEnabled={milestoneAttributes.valid}
		/>
	</svelte:fragment>
</Dialog>
