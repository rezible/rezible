<script lang="ts">
	import Toaster, { setToastState } from "$components/toaster";
	import Sidebar from "./Sidebar.svelte";
	import Header from "./Header.svelte";

	const { children } = $props();

	setToastState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	<Sidebar />

	<div class="grid grid-rows-layout flex-1">
		<nav class="w-full h-16 bg-surface-200">
			<Header />
		</nav>

		<main class="w-full p-2 overflow-y-auto">
			{@render children()}
		</main>
	</div>
</div>

<Toaster />
