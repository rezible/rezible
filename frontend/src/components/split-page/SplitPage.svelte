<script lang="ts">
	import type { Snippet } from "svelte";

	type Props = {
		nav: Snippet;
		children: Snippet;
	};
	const { nav, children }: Props = $props();
</script>

<div class="flex gap-2 w-full h-full max-h-full">
	<div class="max-h-full overflow-hidden w-full max-w-md">
		{@render nav()}
	</div>

	<div class="flex-1 flex flex-col min-h-0 max-h-full px-2 overflow-y-auto border-surface-content/10">
		{@render children()}
	</div>
</div>