<script lang="ts">
	import { mdiClose } from "@mdi/js";
	import { Button, Dialog, Header } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { analysis } from "../analysis.svelte";

	type Props = {};
	let {}: Props = $props();

	const title = "Edit Relationship";
</script>

<Dialog
	open={analysis.componentDialogOpen}
	on:close={() => {
		analysis.setComponentDialogOpen(false);
	}}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2",
		root: "p-2",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header {title}>
			<svelte:fragment slot="actions">
				<Button
					on:click={() => close({ force: true })}
					iconOnly
					icon={mdiClose}
				/>
			</svelte:fragment>
		</Header>
	</div>

	<svelte:fragment slot="default">
		{#if analysis.componentDialogOpen}
			<span>component: {analysis.editingComponent?.id}</span>
		{/if}
	</svelte:fragment>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			onClose={() => {
				analysis.setComponentDialogOpen(false);
			}}
			onConfirm={() => {
				console.log("confirm");
			}}
			saveEnabled={false}
		/>
	</svelte:fragment>
</Dialog>
