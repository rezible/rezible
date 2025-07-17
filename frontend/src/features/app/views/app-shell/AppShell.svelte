<script lang="ts">
	import { session } from "$lib/auth.svelte";
	import { settings } from "$lib/settings.svelte";
	import { setUserOncallInformationState } from "$lib/userOncall.svelte";
	import { setToastState } from "$features/app/lib/toasts.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import Toaster from "$features/app/components/toaster/Toaster.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import PageContainer from "./PageContainer.svelte";

	const { children } = $props();

	appShell.setup();
	settings.setup();

	setToastState();
	setUserOncallInformationState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	{#if session.user}
		<Sidebar />
	{/if}

	<main class="w-full h-full p-2">
		<PageContainer>
			{@render children()}
		</PageContainer>
	</main>
</div>

<Toaster />