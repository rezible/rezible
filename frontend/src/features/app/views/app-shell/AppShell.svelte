<script lang="ts">
	import { session } from "$lib/auth.svelte";
	import { setToastState } from "$features/app/lib/toasts.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import Toaster from "$features/app/components/toaster/Toaster.svelte";
	import Sidebar from "./Sidebar.svelte";
	import Header from "./Header.svelte";
	import PageContainer from "./PageContainer.svelte";
	import HeaderLogo from "./HeaderLogo.svelte";

	const { children } = $props();

	appShell.setup();
	setToastState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-200 text-surface-content">
	{#if !session.user}
		<div class="grid grid-rows-app-shell-layout flex-1">
			<nav class="w-full h-16 border-b">
				<HeaderLogo />
			</nav>

			<main class="w-full overflow-y-auto p-2">
				{@render children()}
			</main>
		</div>
	{:else}
		<Sidebar />

		<div class="grid grid-rows-app-shell-layout flex-1">
			<nav class="w-full h-16">
				<Header />
			</nav>

			<main class="w-full overflow-y-auto px-2 pb-2">
				<PageContainer>
					{@render children()}
				</PageContainer>
			</main>
		</div>
	{/if}
</div>

<Toaster />