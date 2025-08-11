<script lang="ts">
	import type { Snippet } from "svelte";

	type Props = {
		filters: Snippet;
		children: Snippet;
	};
	const { filters, children }: Props = $props();
</script>

<div class="flex gap-3 w-full h-full max-h-full pt-2 pl-2">
	<div class="w-full h-full max-h-full overflow-y-auto max-w-md flex flex-col">
		<span class="text-sm uppercase font-semibold text-surface-content/60">Filters</span>
		<div class="flex flex-col flex-0 gap-1 border p-2 rounded overflow-x-hidden overflow-y-auto">
			{@render filters()}
		</div>
	</div>

	<div class="flex-1 block min-h-0 max-h-full overflow-y-auto">
		{@render children()}
	</div>
</div>