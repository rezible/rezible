<script lang="ts">
	import type { Snippet } from "svelte";
	import { Header } from "svelte-ux";

	type Props = {
		nav: Snippet;
		navClosed?: boolean;
		children: Snippet;
	};
	const { nav, navClosed, children }: Props = $props();
</script>

<div class="flex gap-2 w-full h-full max-h-full">
	<div class="flex flex-col min-h-0 gap-2 max-h-full overflow-hidden pl-1 w-72">
		{@render nav()}
	</div>

	<div class="flex-1 flex flex-col min-h-0 max-h-full px-2 overflow-y-auto border-surface-content/10">
		{@render children()}
	</div>
</div>