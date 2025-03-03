<script lang="ts">
	import ConfirmButtons from '$src/components/confirm-buttons/ConfirmButtons.svelte';
	import { mdiCheck, mdiClose } from '@mdi/js';
	import { Icon, TextField } from 'svelte-ux';

	type Props = {
		label?: string;
		description: string;
		onCancel: VoidFunction;
		onConfirm: VoidFunction;
	}

	let { label = $bindable(), description = $bindable(), onCancel, onConfirm }: Props = $props();

	const saveEnabled = $derived((label !== undefined) ? !!label : !!description);
</script>

<div class="w-full flex flex-col border rounded-lg p-2 gap-2">
	{#if label !== undefined}
		<TextField
			label="Label"
			labelPlacement="float"
			bind:value={label}
		/>
	{/if}

	<TextField
		label="Description"
		labelPlacement="float"
		bind:value={description}
	/>
	
	<ConfirmButtons onClose={onCancel} {onConfirm} {saveEnabled}>
		{#snippet closeButtonContent()}<Icon data={mdiClose} />{/snippet}
		{#snippet confirmButtonContent()}<Icon data={mdiCheck} />{/snippet}
	</ConfirmButtons>
</div>
