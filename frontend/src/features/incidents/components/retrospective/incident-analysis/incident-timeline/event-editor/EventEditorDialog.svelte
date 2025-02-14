<script lang="ts">
	import { Button, Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import EventAttributesEditor from "./EventAttributesEditor.svelte";
	import { eventDialog } from "./eventEditorDialog.svelte";
</script>

<Dialog
	open={eventDialog.open}
	on:close={() => eventDialog.clear}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-7xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2">
		<span class="text-xl">Create Event</span>
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto flex">
		{#if eventDialog.open}
			<EventAttributesEditor />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			onClose={eventDialog.clear}
			confirmText={eventDialog.view === "create" ? "Create" : "Save"}
			onConfirm={eventDialog.confirm}
			saveEnabled={eventDialog.saveEnabled}
		/>
	</svelte:fragment>
</Dialog>
