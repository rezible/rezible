<script lang="ts">
	import type { Snippet } from "svelte";
	import { Header } from "svelte-ux";

	type Props = {
		nav: Snippet;
		children: Snippet;
	};
	const { nav, children }: Props = $props();
</script>

<div class="grid grid-cols-5 gap-2 h-full max-h-full divide-x">
	<div class="flex flex-col min-h-0 gap-2 max-h-full overflow-hidden pl-1">
		{@render nav()}
	</div>

	<div class="col-span-4 flex flex-col min-h-0 gap-2 max-h-full px-2">
		{@render children()}
	</div>
</div>