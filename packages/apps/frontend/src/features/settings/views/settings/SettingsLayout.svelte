<script lang="ts">
	import type { Snippet } from "svelte";
	import { initSettingsViewController } from "./controller.svelte";

	type Props = {
		children: Snippet;
	};
	const { children }: Props = $props();

    initSettingsViewController();
</script>

<div class="flex min-h-0 flex-1 flex-col overflow-y-auto p-4">
	{@render children()}
</div>
