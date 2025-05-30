<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import ComponentAttributesEditor from "./ComponentAttributesEditor.svelte";
	import ComponentSelector from "./ComponentSelector.svelte";
	import { useComponentDialog } from "./dialogState.svelte";
	import Header from "$src/components/header/Header.svelte";
	
	const componentDialog = useComponentDialog();

	const viewActionLabels: Record<typeof componentDialog.view, { action: string; confirm: string }> = {
		add: { action: "Add", confirm: "Add" },
		create: { action: "Create", confirm: "Create Component" },
		edit: { action: "Edit", confirm: "Save Changes" },
		closed: { action: "", confirm: "" },
	};
	const labels = $derived(viewActionLabels[componentDialog.view]);

	const creatingToAdd = $derived(
		componentDialog.view === "create" && componentDialog.previousView === "add"
	);
	const cancelLabel = $derived(creatingToAdd ? "Cancel" : "Cancel");
</script>

<Dialog
	open={componentDialog.open}
	on:close={() => componentDialog.clear()}
	loading={componentDialog.loading}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-7xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="{labels.action} System Component">
			{#snippet actions()}
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			{/snippet}
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 overflow-y-auto">
		{#if componentDialog.open}
			{#if componentDialog.view === "add"}
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
			saveEnabled={componentDialog.valid}
			onClose={() => componentDialog.goBack()}
			onConfirm={() => componentDialog.onConfirm()}
		/>
	</svelte:fragment>
</Dialog>
