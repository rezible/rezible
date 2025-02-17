<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog, Header } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import ComponentAttributesEditor from "./ComponentAttributesEditor.svelte";
	import ComponentSelector from "./ComponentSelector.svelte";
	import { componentDialog } from "./dialogState.svelte";
	
	componentDialog.setup();

	const view = $derived(componentDialog.view);
	const viewActionLabels: Record<typeof view, {action: string, confirm: string}> = {
		add: {action: "Add", confirm: "Add"},
		create: {action: "Create", confirm: "Create Component"},
		edit: {action: "Edit", confirm: "Save Changes"},
		closed: {action: "", confirm: ""},
	}
	const labels = $derived(viewActionLabels[view]);

	const creatingToAdd = $derived(view === "create" && componentDialog.previousView === "add");
	const cancelLabel = $derived(creatingToAdd ? "Cancel" : "Cancel");
</script>

<Dialog
	open={componentDialog.open}
	on:close={componentDialog.clear}
	loading={componentDialog.loading}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="{labels.action} System Component">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid">
		{#if componentDialog.open}
			{#if view === "add"}
				<ComponentSelector />
			{:else}
				<ComponentAttributesEditor />
			{/if}
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			loading={componentDialog.loading}
			closeText={cancelLabel}
			confirmText={labels.confirm}
			saveEnabled={componentDialog.stateValid}
			onClose={componentDialog.goBack}
			onConfirm={componentDialog.confirm}
		/>
	</svelte:fragment>
</Dialog>
