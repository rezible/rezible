<script lang="ts">
	import { Button, Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import EventAnnotationForm from "./EventAnnotationForm.svelte";
	import { useAnnotationDialogState } from "./dialogState.svelte";
	import DialogTitle from "./DialogTitle.svelte";

	const dialog = useAnnotationDialogState();

	const formAction = $derived(!!dialog.annotation ? "Update" : "Create");
</script>

<Dialog
	open={dialog.open}
	on:close={() => dialog.onClose()}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-5xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" let:close>
		{#if !!dialog.event}
			<DialogTitle event={dialog.event} close={() => close({ force: true })} />
		{/if}
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid overflow-y-auto">
		{#if !!dialog.event}
			<EventAnnotationForm event={dialog.event} current={dialog.annotation} />
		{/if}
	</div>

	<svelte:fragment slot="actions">
		{#if dialog.view !== "view"}
			<ConfirmButtons
				loading={false}
				closeText="Cancel"
				confirmText={formAction}
				onClose={() => dialog.onClose()}
				onConfirm={() => dialog.onConfirm()}
			/>
		{:else}
			<Button on:click={() => dialog.onClose()}>Close</Button>
		{/if}
	</svelte:fragment>
</Dialog>
