<script lang="ts">
	import type { Snippet } from "svelte";
	import { Toaster } from "$components/ui/sonner";
	import { initSettingsViewController } from "./controller.svelte";
	import SettingsAccessDenied from "./settings-access-denied/SettingsAccessDenied.svelte";

	type Props = {
		children: Snippet;
	};
	const { children }: Props = $props();

	const controller = initSettingsViewController();
</script>

<Toaster />

<div class="min-h-0 min-w-0 flex-1 overflow-y-auto [overflow-wrap:anywhere] p-4 md:p-6">
	{#if controller.showAccessDenied}
		<SettingsAccessDenied />
	{:else}
		{@render children()}
	{/if}
</div>
