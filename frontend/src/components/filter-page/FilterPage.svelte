<script lang="ts">
	import type { Snippet } from "svelte";

	type Props = {
		filters: Snippet;
		children: Snippet;
	};
	const { filters, children }: Props = $props();
</script>

<div class="flex gap-2 w-full h-full max-h-full">
	<div class="max-h-full overflow-x-hidden overflow-y-auto w-full max-w-sm flex flex-col gap-1">
		{@render filters()}
	</div>

	<div class="flex-1 flex flex-col min-h-0 max-h-full overflow-y-auto border-surface-content/10">
		{@render children()}
	</div>
</div>