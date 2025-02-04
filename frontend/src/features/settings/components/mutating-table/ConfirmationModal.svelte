<script lang="ts">
	import { Dialog } from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";

	interface Props {
		open: boolean;
		title: string;
		text: string;
		loading?: boolean;
		onConfirm: () => void;
		onClose: () => void;
	}
	let { open, title, text, loading, onConfirm, onClose }: Props = $props();
</script>

<Dialog {open} on:close={onClose} {loading}>
	<div slot="title">{title}</div>
	<div class="px-6 py-3">{text}</div>
	<div slot="actions">
		<ConfirmButtons {onConfirm} {onClose} />
	</div>
</Dialog>
