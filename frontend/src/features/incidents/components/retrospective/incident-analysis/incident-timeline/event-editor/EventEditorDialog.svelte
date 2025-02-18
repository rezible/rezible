<script lang="ts">
	import { Button, Dialog, Icon, Tooltip } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import EventAttributesEditor from "./EventAttributesEditor.svelte";
	import { eventDialog } from "./dialogState.svelte";
	import { mdiMagicStaff } from "@mdi/js";
</script>

<Dialog
	open={eventDialog.open}
	on:close={() => eventDialog.clear}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-7xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">Create Event</span>
		<div class="">
			<Button variant="fill-light" color="secondary">
				<span class="flex gap-2 items-center">
					AI Draft
					<Icon data={mdiMagicStaff} />
				</span>
			</Button>
		</div>
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
