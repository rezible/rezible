<script lang="ts">
	import { session } from '$lib/auth.svelte';
    import Toaster, { setToastState } from '$src/components/toaster';
	
	import Sidebar from "./Sidebar.svelte";
	import Header from './Header.svelte';

	setToastState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	<Sidebar />

	<div class="grid grid-rows-layout flex-1">
		<nav class="w-full h-16 bg-surface-200">
			<Header />
		</nav>

		<main class="w-full p-2 overflow-y-auto">
			{#if !session.isValid}
				<div class="m-4 text-lg text-danger">auth session invalid</div>
			{/if}

			<slot />
		</main>
	</div>
</div>

<Toaster />