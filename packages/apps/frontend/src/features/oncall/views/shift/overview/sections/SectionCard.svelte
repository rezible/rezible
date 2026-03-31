<script lang="ts">
	import type { Snippet } from "svelte";

	const { children }: { children: Snippet } = $props();
</script>

<div class="flex flex-col gap-2 w-full p-2 rounded border border-surface-content/10">
	{@render children()}
</div>