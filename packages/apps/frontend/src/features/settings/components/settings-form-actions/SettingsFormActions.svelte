<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	type Props = {
		onSave: () => void;
		onCancel: () => void;
		dirty: boolean;
		pending?: boolean;
		disabled?: boolean;
		saveDisabled?: boolean;
		saveLabel?: string;
		// Inline editors close on cancel, so their Cancel stays available without changes.
		cancelRequiresChanges?: boolean;
	};
	const {
		onSave,
		onCancel,
		dirty,
		pending = false,
		disabled = false,
		saveDisabled = false,
		saveLabel = "Save",
		cancelRequiresChanges = true,
	}: Props = $props();

	const cancelDisabled = $derived(disabled || pending || (cancelRequiresChanges && !dirty));
</script>

<div class="flex flex-wrap items-center gap-2">
	<Button onclick={onSave} disabled={disabled || saveDisabled || pending || !dirty}>
		{#if pending}
			<Spinner />
		{/if}
		{pending ? "Saving…" : saveLabel}
	</Button>
	<Button variant="outline" onclick={onCancel} disabled={cancelDisabled}>Cancel</Button>
</div>
