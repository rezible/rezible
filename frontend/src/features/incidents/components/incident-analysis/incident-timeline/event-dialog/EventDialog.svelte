<script lang="ts">
	import { Button, Dialog, Icon } from "svelte-ux";
	import { mdiMagicStaff } from "@mdi/js";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import EventAttributesEditor from "./EventAttributesEditor.svelte";
	import { eventDialog } from "./dialogState.svelte";
	import { eventAttributes } from "./attribute-panels/eventAttributes.svelte";

	eventDialog.setup();

	const creating = $derived(eventDialog.view === "create");
</script>

<Dialog
	open={eventDialog.open}
	on:close={() => eventDialog.clear}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-7xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">{creating ? "Create" : "Edit"} Event</span>
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
			confirmText={creating ? "Create" : "Save"}
			onConfirm={eventDialog.confirm}
			saveEnabled={eventAttributes.valid}
		/>
	</svelte:fragment>
</Dialog>
