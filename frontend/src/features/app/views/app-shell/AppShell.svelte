<script lang="ts">
	import { session } from "$lib/auth.svelte";
	import { setToastState } from "$features/app/lib/toasts.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import Toaster from "$features/app/components/toaster/Toaster.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import PageContainer from "./PageContainer.svelte";

	const { children } = $props();

	appShell.setup();
	setToastState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	{#if session.user}
		<Sidebar />

		<main class="w-full p-2">
			<PageContainer>
				{@render children()}
			</PageContainer>
		</main>
	{:else}
		<main class="w-full p-2">
			{@render children()}
		</main>
	{/if}
</div>

<Toaster />