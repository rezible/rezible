<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog, Header } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { relationshipDialog } from "./relationshipDialog.svelte";
	
	relationshipDialog.setup();

	const view = $derived(relationshipDialog.view);
	const viewActionLabels: Record<typeof view, {action: string, confirm: string}> = {
		create: {action: "Create", confirm: "Create Relationship"},
		edit: {action: "Edit", confirm: "Save Changes"},
		closed: {action: "", confirm: ""},
	}
	const labels = $derived(viewActionLabels[view]);
</script>

<Dialog
	open={relationshipDialog.open}
	on:close={relationshipDialog.clear}
	loading={relationshipDialog.loading}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2 min-h-0",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="{labels.action} Relationship">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<div slot="default" class="p-2 flex-1 min-h-0 max-h-full grid">
		{#if relationshipDialog.open}
			<span>{relationshipDialog.view}</span>
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			loading={relationshipDialog.loading}
			closeText="Cancel"
			confirmText={labels.confirm}
			saveEnabled={relationshipDialog.stateValid}
			onClose={relationshipDialog.goBack}
			onConfirm={relationshipDialog.confirm}
		/>
	</svelte:fragment>
</Dialog>
