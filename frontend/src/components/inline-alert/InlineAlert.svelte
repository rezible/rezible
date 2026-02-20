<script lang="ts">
	import { Button } from "$components/ui/button";
	import type { ErrorModel } from "$src/lib/api";

	type Props = {
		error: ErrorModel;
		onDismiss?: () => void;
	};
	const { error, onDismiss }: Props = $props();
</script>

<div class="mb-2 rounded border border-danger/40 bg-danger/10 p-3" role="alert">
	<div class="flex items-start justify-between gap-3">
		<div class="flex flex-col">
			<span class="font-semibold text-danger">{error.title ?? "An Error Occurred"}</span>
			<span class="text-sm text-danger/90">{error.detail ?? ""}</span>
		</div>
		{#if onDismiss}
			<Button onclick={() => onDismiss()}>Dismiss</Button>
		{/if}
	</div>
</div>
