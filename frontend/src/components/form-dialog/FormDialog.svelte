<script lang="ts">
	import type { Snippet } from "svelte";
	import { Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";

	type Props = {
		title: string;
		open: boolean;
		loading?: boolean;
		children: Snippet;
		titleActions?: Snippet;
		onClose: () => void;
		onConfirm: () => void;
		confirmText?: string;
		saveEnabled?: boolean;
	};
	const {
		title,
		open,
		loading,
		children,
		titleActions,
		onConfirm,
		onClose,
		confirmText = "Confirm",
		saveEnabled = true,
	}: Props = $props();
</script>

<Dialog
	{open}
	on:close={() => onClose()}
	{loading}
	persistent
	portal
	classes={{ root: "p-8", dialog: "flex flex-col w-full max-w-3xl max-h-full h-fit" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">{title}</span>
		{#if titleActions}
			<div class="">
				{@render titleActions()}
			</div>
		{/if}
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto flex">
		{#if open}
			{@render children()}
		{/if}
	</div>

	<svelte:fragment slot="actions">
		<ConfirmButtons {loading} {onClose} {confirmText} {onConfirm} {saveEnabled} />
	</svelte:fragment>
</Dialog>
