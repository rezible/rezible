<script lang="ts">
    import { mdiClose } from "@mdi/js";
    import { Button, Dialog, Header } from "svelte-ux";
    import type { SystemAnalysisComponent } from "$lib/api";
    import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";

    type Props = {open: boolean}
    let { open = $bindable() }: Props = $props();
</script>

<Dialog
	bind:open
	on:close={() => {open = false}}
	persistent
	portal
	classes={{ dialog: 'flex flex-col max-h-full w-5/6 max-w-7xl my-2', root: "p-2" }}
	>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="Add Component">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({force: true})} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<svelte:fragment slot="default">
		{#if open}
			<span>component</span>
		{/if}
	</svelte:fragment>

	<svelte:fragment slot="actions">
		<ConfirmButtons
			onClose={() => {open = false}}
			onConfirm={() => {console.log("confirm")}}
			saveEnabled={false}
		/>
	</svelte:fragment>
</Dialog>