{#snippet defaultHeader()}
	<Header title="Filters" />
{/snippet}

<script lang="ts">
	import type { Snippet } from "svelte";
	import Header from "../header/Header.svelte";

	type Props = {
		filters: Snippet;
		header?: Snippet;
		children: Snippet;
	};
	const { filters, header = defaultHeader, children }: Props = $props();
</script>

<div class="flex gap-3 h-full max-h-full">
	<div class="w-full h-full max-h-full overflow-y-auto max-w-md flex flex-col gap-1">
		{@render header()}
		<div class="flex flex-col gap-4 border p-2 rounded overflow-x-hidden overflow-y-auto">
			{@render filters()}
		</div>
	</div>

	<div class="flex-1 block min-h-0 max-h-full overflow-y-auto">
		{@render children()}
	</div>
</div>