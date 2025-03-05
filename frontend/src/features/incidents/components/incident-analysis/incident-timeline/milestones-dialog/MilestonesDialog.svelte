<script lang="ts">
	import { Button, Dialog } from "svelte-ux";
	import { milestonesDialog } from "./dialogState.svelte";
	import MilestonesEditor from "./MilestonesEditor.svelte";
	import { mdiClose } from "@mdi/js";

	const editAction = $derived(milestonesDialog.editingMilestone ? "Edit" : "Create");
	const title = $derived(milestonesDialog.editorOpen ? `${editAction} Milestone` : "Incident Milestones");
</script>

<Dialog
	open={milestonesDialog.open}
	on:close={milestonesDialog.close}
	portal
	persistent
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-5xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">{title}</span>
		<Button size="sm" icon={mdiClose} iconOnly on:click={milestonesDialog.close} />
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto flex overflow-x-hidden">
		{#if milestonesDialog.open}
			<MilestonesEditor />
		{/if}
	</div>
</Dialog>
