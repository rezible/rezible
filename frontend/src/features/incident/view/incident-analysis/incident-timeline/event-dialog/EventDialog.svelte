<script lang="ts">
	import { Dialog } from "svelte-ux";
	import { mdiMagicStaff } from "@mdi/js";
	import Button from "$components/button/Button.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import EventAttributesEditor from "./EventAttributesEditor.svelte";
	import { eventAttributes } from "./attribute-panels/eventAttributesState.svelte";
	import { useEventDialog } from "./dialogState.svelte";

	const eventDialog = useEventDialog();

	const attributesValid = $derived(!!eventAttributes.title);

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
			onClose={() => eventDialog.clear()}
			confirmText={creating ? "Create" : "Save"}
			onConfirm={() => eventDialog.confirm()}
			saveEnabled={attributesValid}
		/>
	</svelte:fragment>
</Dialog>
