<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button } from 'svelte-ux';

	type Props = {
		onClose: VoidFunction;
		onConfirm: VoidFunction;
		saveEnabled?: boolean;
		disabled?: boolean;
		loading?: boolean;
		alignRight?: boolean;
		confirmText?: string;
		confirmButtonContent?: Snippet;
		closeText?: string;
		closeButtonContent?: Snippet;
	};
	let {
		saveEnabled = true,
		disabled = false,
		loading = false,
		confirmText = "Confirm",
		closeText = "Close",
		alignRight = false,
		confirmButtonContent,
		closeButtonContent,
		onClose,
		onConfirm,
	}: Props = $props();
</script>

<div class="flex flex-row gap-2" class:justify-end={alignRight}>
	<Button
		on:click={() => {onClose()}}
		disabled={disabled || loading}
	>
		{#if closeButtonContent}
			{@render closeButtonContent()}
		{:else}
			{closeText}
		{/if}
	</Button>
	<Button
		{loading}
		variant="fill"
		color="warning"
		on:click={() => {onConfirm()}}
		disabled={!saveEnabled || disabled}
	>
		{#if confirmButtonContent}
			{@render confirmButtonContent()}
		{:else}
			{confirmText}
		{/if}
	</Button>
</div>
